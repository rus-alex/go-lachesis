package main

//go:generate bash -c "docker run --rm -v $(pwd)/ballot:/src -v $(pwd)/ballot:/dst ethereum/solc:0.5.12 -o /dst/ --optimize --optimize-runs=2000 --bin --abi --overwrite /src/Ballot.sol"
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./ballot/Ballot.bin --abi=./ballot/Ballot.abi --pkg=ballot --type=Contract --out=ballot/contract.go

import (
	"crypto/ecdsa"
	"math/big"

	// "github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/Fantom-foundation/go-lachesis/cmd/api-caller/ballot"
)

func (g *Generator) createContract(key *ecdsa.PrivateKey, nonce uint, amount *big.Int) *types.Transaction {
	bin := common.FromHex(ballot.ContractBin)
	tx := types.NewContractCreation(uint64(nonce), amount, gasLimit*10000, gasPrice, bin)
	tx, err := types.SignTx(tx, g.signer, key)
	if err != nil {
		panic(err)
	}

	return tx
}

/*
func (env *testEnv) Payer(n int, amounts ...*big.Int) *bind.TransactOpts {
	key := env.privateKey(n)
	t := bind.NewKeyedTransactor(key)
	nonce, _ := env.PendingNonceAt(nil, env.Address(n))
	t.Nonce = big.NewInt(int64(nonce))
	t.Value = big.NewInt(0)
	for _, amount := range amounts {
		t.Value.Add(t.Value, amount)
	}
	return t
}
*/
