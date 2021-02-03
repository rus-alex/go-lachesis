package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/VictoriaMetrics/fastcache"
)

func main() {
	fmt.Println("Start")
	defer fmt.Println("Stop")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	cache := fastcache.New(16 * 1024 * 1024)
	const size = 10 * 1024

	for i := int64(0); true; i++ {
		select {
		case <-done:
			return
		default:
		}

		val := make([]byte, size)
		_, err := rand.Read(val)
		if err != nil {
			panic(err)
		}

		key := hash.FakeHash(i)
		cache.Set(key[:], val)

		if i > 10 && i%10 == 1 {
			fmt.Printf("reading of step %d on step %d\n", i-10, i)
			prev := hash.FakeHash(i - 10)
			got := cache.Get(nil, prev[:])
			if got == nil {
				panic("Key not found!")
			}
			if len(got) != size {
				panic("Invalid size!")
			}
		}

		if i%10000 == 1 {
			fmt.Printf("progress: %d\n", i)
		}

	}

}
