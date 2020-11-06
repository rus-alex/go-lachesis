package gossip

// Simple ballot contract
//go:generate bash -c "docker run --rm -v $(pwd)/ballot:/src ethereum/solc:0.5.12 -o /src/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src --overwrite /src/Ballot.sol"
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=ballot/solc/Ballot.bin --abi=ballot/solc/Ballot.abi --pkg=ballot --type=Contract --out=ballot/contract.go

import (
	"testing"
)

func BenchmarkStateDB(b *testing.B) {
	b.Log("OK")
}
