package main

import (
	"fmt"
	"math"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Fantom-foundation/go-lachesis/logger"
)

type TxMaker func(*ethclient.Client) (*types.Transaction, error)
type TxCallback func(*types.Receipt, error)

type Transaction struct {
	Make     TxMaker
	Callback TxCallback
	Dsc      string
}

type Generator struct {
	tps     uint32
	chainId uint
	signer  types.Signer

	instances      uint
	accs           []*Acc
	offset         uint
	position       uint
	generatorState genState

	work sync.WaitGroup
	done chan struct{}
	sync.Mutex

	logger.Instance
}

func NewTxGenerator(cfg *Config, num, ofTotal uint) *Generator {
	accs := cfg.Accs.Count / ofTotal
	offset := cfg.Accs.Offset + accs*(num-1)
	g := &Generator{
		chainId:   uint(cfg.ChainId),
		signer:    types.NewEIP155Signer(big.NewInt(int64(cfg.ChainId))),
		instances: ofTotal,
		accs:      make([]*Acc, accs),
		offset:    offset,

		Instance: logger.MakeInstance(),
	}

	return g
}

func (g *Generator) Start() (output chan *Transaction) {
	g.Lock()
	defer g.Unlock()

	if g.done != nil {
		return
	}
	g.done = make(chan struct{})

	output = make(chan *Transaction, 100)
	g.work.Add(1)
	go g.background(output)

	g.Log.Info("will use", "accounts", len(g.accs), "from", g.offset, "to", uint(len(g.accs))+g.offset)
	return
}

func (g *Generator) Stop() {
	g.Lock()
	defer g.Unlock()

	if g.done == nil {
		return
	}

	close(g.done)
	g.work.Wait()
	g.done = nil
}

func (g *Generator) getTPS() float64 {
	tps := atomic.LoadUint32(&g.tps)
	return float64(tps)
}

func (g *Generator) SetTPS(tps float64) {
	x := uint32(math.Ceil(tps / float64(g.instances)))
	atomic.StoreUint32(&g.tps, x)
}

func (g *Generator) background(output chan<- *Transaction) {
	defer g.work.Done()
	defer close(output)

	g.Log.Info("started")
	defer g.Log.Info("stopped")

	for {
		begin := time.Now()
		var (
			generating time.Duration
			sending    time.Duration
		)

		tps := g.getTPS()
		for count := tps; count > 0; count-- {
			begin := time.Now()
			tx := g.Yield()
			generating += time.Since(begin)

			begin = time.Now()
			select {
			case output <- tx:
				sending += time.Since(begin)
				continue
			case <-g.done:
				return
			}
		}

		spent := time.Since(begin)
		if spent >= time.Second {
			g.Log.Warn("exceeded performance", "tps", tps, "generating", generating, "sending", sending)
			continue
		}

		select {
		case <-time.After(time.Second - spent):
			continue
		case <-g.done:
			return
		}
	}
}

func (g *Generator) Yield() *Transaction {
	if !g.generatorState.IsReady(g.done) {
		return nil
	}
	tx := g.generate(g.position, &g.generatorState)
	g.Log.Info("generated tx", "position", g.position, "dsc", tx.Dsc)
	g.position++

	return tx
}

type genState struct {
	ready      chan struct{}
	BallotAddr common.Address
}

func (s *genState) NotReady() {
	s.ready = make(chan struct{})
}

func (s *genState) IsReady(done <-chan struct{}) bool {
	if s.ready == nil {
		return true
	}

	select {
	case <-done:
		return false
	case <-s.ready:
		return true
	}
}

func (s *genState) Ready() {
	close(s.ready)
}

func (g *Generator) generate(position uint, state *genState) *Transaction {
	// count := uint(len(g.accs))
	var (
		maker    TxMaker
		callback TxCallback
		dsc      string
	)

	a := position % uint(len(g.accs))

	switch step := (position % 10001); {

	case step == 0:
		dsc = "ballot contract creation"
		maker = g.ballotCreateContract(a)
		state.NotReady()
		callback = func(r *types.Receipt, e error) {
			state.BallotAddr = r.ContractAddress
			state.Ready()
		}

	case 0 < step && step < 10000:
		chose := ballotRandChose()
		dsc = fmt.Sprintf("%d voites for %d", a, chose)
		maker = g.ballotVoite(a, state.BallotAddr, chose)
		break

	case step == 10000:
		dsc = "ballot winner reading"
		maker = g.ballotWinner(state.BallotAddr)

	default:
		panic(fmt.Sprintf("unknown step %d", step))
	}

	return &Transaction{
		Make:     maker,
		Callback: callback,
		Dsc:      dsc,
	}
}

func (g *Generator) Payer(n uint, amounts ...*big.Int) *bind.TransactOpts {
	from := g.accs[n]
	if from == nil {
		from = MakeAcc(n + g.offset)
		g.accs[n] = from
	}

	t := bind.NewKeyedTransactor(from.Key)

	t.Value = big.NewInt(0)
	for _, amount := range amounts {
		t.Value.Add(t.Value, amount)
	}

	return t
}

func (g *Generator) ReadOnly() *bind.CallOpts {
	return &bind.CallOpts{}
}
