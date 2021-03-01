package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/topicsdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func main() {

	db := topicsdb.New(memorydb.New())

	for i := 0; i < 100; i++ {
		blockN := uint64(i / 10)
		txN := uint(i % 10)
		db.Push(&types.Log{
			BlockNumber: blockN,
			Index:       txN,
			TxHash:      hash.FakeHash(int64(i)),
			Topics: []common.Hash{
				hash.FakeHash(int64(10*txN + 0)),
				hash.FakeHash(int64(10*txN + 1)),
				hash.FakeHash(int64(10*txN + 2)),
			},
		})
	}

	wg := &sync.WaitGroup{}
	done := make(chan struct{})

	wg.Add(10)
	for i := uint(0); i < 10; i++ {
		go func(i uint) {
			defer wg.Done()

			txN := uint(i)
			topics := make([][]common.Hash, 3)
			topics[i%3] = []common.Hash{
				hash.FakeHash(int64(10*txN + i%3)),
				hash.FakeHash(int64(i)),
			}
			for {
				select {
				case <-done:
					return
				default:
				}

				logs, err := db.Find(topics)
				if err != nil {
					panic(err)
				}
				if len(logs) < 10 {
					panic("no 10 logs found")
				}
			}
		}(i)
	}

	waitForSignal()
	close(done)
	fmt.Println("Stopping ...")
	wg.Wait()
	fmt.Println("Finish.")
}

func waitForSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
}
