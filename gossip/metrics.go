package gossip

import (
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	confirmBlocksMeter = metrics.NewRegisteredCounter("confirm/blocks", nil)
	confirmTxnsMeter   = metrics.NewRegisteredCounter("confirm/transactions", nil)
	txTtfMeter         = metrics.NewRegisteredHistogram("tx_ttf", nil, metrics.NewUniformSample(500))
)

var (
	txLatency = newTxs()

	errUnknownTx = errors.New("unknown tx")
)

type Txs struct {
	txs map[common.Hash]time.Time
	sync.Mutex
}

func newTxs() *Txs {
	return &Txs{
		txs: make(map[common.Hash]time.Time),
	}
}

func (d *Txs) Start(tx common.Hash) {
	d.Lock()
	d.txs[tx] = time.Now()
	d.Unlock()
}

func (d *Txs) Finish(tx common.Hash) (latency time.Duration, err error) {
	d.Lock()
	defer d.Unlock()

	start, ok := d.txs[tx]
	if !ok {
		err = errUnknownTx
		return
	}
	delete(d.txs, tx)

	latency = time.Since(start)
	return
}
