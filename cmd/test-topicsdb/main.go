package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/topicsdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func main() {
	fmt.Println("Start")
	defer fmt.Println("Stop")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	db := topicsdb.New(memorydb.New())
	const (
		cicle = 10000
		size  = 10 * 1024
	)

	for i := int64(0); true; i++ {
		select {
		case <-done:
			return
		default:
		}

		n := i % cicle
		rec := &types.Log{
			BlockNumber: uint64(n / 10),
			BlockHash:   hash.FakeHash((n)),
			TxHash:      hash.FakeHash((n + 1)),
			Index:       uint(n % 10),
			Data:        make([]byte, size),
			Topics:      make([]common.Hash, 10),
		}
		for j := range rec.Topics {
			rec.Topics[j] = hash.FakeHash(n*10 + int64(j))
		}

		// write
		if i < cicle {
			rand.Read(rec.Data)
			db.MustPush(rec)
			continue
		}

		// read
		for j := 0; j < len(rec.Topics)-1; j += 2 {
			templ := make([][]common.Hash, len(rec.Topics))
			templ[j] = []common.Hash{rec.Topics[j], hash.FakeHash(int64(j))}
			templ[j+1] = []common.Hash{hash.FakeHash(int64(j)), rec.Topics[j+1]}
			got, err := db.Find(templ)
			if err != nil {
				panic(err)
			}

			fmt.Printf("found: %d\n", len(got))
		}

		if i%cicle == 1 {
			fmt.Printf("progress: %d\n", i)
		}

	}

}
