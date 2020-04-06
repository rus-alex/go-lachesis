package gossip

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecQueue(t *testing.T) {
	var (
		N        = 10000
		q        = newExecQueue(N)
		counter  int
		execd    = make(chan int)
		testexit = make(chan struct{})
	)
	defer q.quit()
	defer close(testexit)

	check := func(state string, wantOK bool) {
		c := counter
		counter++
		qf := func() {
			select {
			case execd <- c:
			case <-testexit:
			}
		}
		require.Equalf(t, wantOK, q.canQueue(), "canQueue() == %t for %s", !wantOK, state)

		require.Equalf(t, wantOK, q.queue(qf), "canQueue() == %t for %s", !wantOK, state)
	}

	for i := 0; i < N; i++ {
		check("queue below cap", true)
	}
	check("full queue", false)
	for i := 0; i < N; i++ {
		c := <-execd
		require.Equal(t, i, c, "execution out of order")
	}
	q.quit()
	check("closed queue", false)
}
