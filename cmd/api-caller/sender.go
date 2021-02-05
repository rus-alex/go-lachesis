package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type Sender struct {
	url       string
	input     chan *Transaction
	callbacks map[common.Hash]TxCallback
	headers   chan *types.Header

	cfg struct {
		listenToBlocks bool
		sendTrustedTx  bool
	}

	done chan struct{}
	work sync.WaitGroup

	logger.Instance
}

func NewSender(url string) *Sender {
	s := &Sender{
		url:       url,
		input:     make(chan *Transaction, 10),
		callbacks: make(map[common.Hash]TxCallback),
		headers:   make(chan *types.Header, 1),
		done:      make(chan struct{}),

		Instance: logger.MakeInstance(),
	}

	s.work.Add(1)
	go s.background()

	return s
}

func (s *Sender) Close() {
	if s.done == nil {
		return
	}
	close(s.done)
	s.done = nil

	s.work.Wait()
	close(s.input)
}

func (s *Sender) Send(tx *Transaction) {
	s.input <- tx
}

func (s *Sender) background() {
	defer s.work.Done()
	s.Log.Info("started")
	defer s.Log.Info("stopped")

	var (
		client *ethclient.Client
		err    error
		tx     *Transaction
		info   string
		sbscr  ethereum.Subscription
	)

	disconnect := func() {
		if sbscr != nil {
			sbscr.Unsubscribe()
			sbscr = nil
		}
		if client != nil {
			client.Close()
			client = nil
			s.Log.Error("disonnect from", "url", s.url)
		}
	}
	defer disconnect()

	for {

		for tx == nil {
			select {
			case b := <-s.headers:
				err = s.onNewHeader(client, b)
				if err != nil {
					disconnect()
				}
			case <-s.done:
				return
			case tx = <-s.input:
			}
		}

		for client == nil {
			client = s.connect()
		}

		if s.cfg.listenToBlocks && sbscr == nil {
			sbscr = s.subscribe(client)
			if sbscr == nil {
				disconnect()
				continue
			}
		}

		txHash := tx.Raw.Hash()
		if tx.Callback != nil {
			s.callbacks[txHash] = tx.Callback
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = client.SendTransaction(ctx, tx.Raw)
		cancel()
		if err == nil {
			txCountSentMeter.Inc(1)
			s.Log.Debug("tx sending ok", "hash", txHash, "amount", tx.Raw.Value(), "nonce", tx.Raw.Nonce())
			tx = nil
			continue
		}

		if tx.Callback != nil {
			delete(s.callbacks, txHash)
			tx.Callback(nil, err)
		}

		switch err.Error() {
		case fmt.Sprintf("known transaction: %x", txHash),
			evmcore.ErrNonceTooLow.Error(),
			evmcore.ErrReplaceUnderpriced.Error():
			s.Log.Warn("tx sending skip", "info", info, "amount", tx.Raw.Value(), "cause", err, "nonce", tx.Raw.Nonce())
			tx = nil
			continue
		default:
			s.Log.Error("tx sending err", "info", info, "amount", tx.Raw.Value(), "cause", err, "nonce", tx.Raw.Nonce())
			disconnect()
			s.delay()
			continue
		}
	}
}

func (s *Sender) connect() *ethclient.Client {
	client, err := ethclient.Dial(s.url)
	if err != nil {
		s.Log.Error("connect to", "url", s.url, "err", err)
		s.delay()
		return nil
	}
	s.Log.Info("connect to", "url", s.url)
	return client
}

func (s *Sender) subscribe(client *ethclient.Client) ethereum.Subscription {
	sbscr, err := client.SubscribeNewHead(context.Background(), s.headers)
	if err != nil {
		s.Log.Error("subscribe to", "url", s.url, "err", err)
		s.delay()
		return nil
	}
	s.Log.Info("subscribe to", "url", s.url)
	return sbscr
}

func (s *Sender) onNewHeader(client *ethclient.Client, h *types.Header) error {
	b := evmcore.ConvertFromEthHeader(h)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	txsCount, err := client.TransactionCount(ctx, b.Hash)
	if err != nil {
		s.Log.Error("new block", "number", b.Number, "block", b.Hash, "err", err)
		return err
	}

	for index := uint(0); index < txsCount; index++ {
		tx, err := client.TransactionInBlock(ctx, b.Hash, index)
		if err != nil {
			s.Log.Error("new tx", "number", b.Number, "block", b.Hash, "index", index, "err", err)
			return err
		}
		txHash := tx.Hash()

		callback := s.callbacks[txHash]
		if callback == nil {
			continue
		}

		r, err := client.TransactionReceipt(ctx, txHash)
		if err != nil {
			s.Log.Error("new recepie", "number", b.Number, "block", b.Hash, "tx", txHash, "err", err)
			return err
		}

		callback(r, nil)
	}

	return nil
}

func (s *Sender) delay() {
	select {
	case <-time.After(2 * time.Second):
		return
	case <-s.done:
		return
	}
}
