package main

//go:generate bash -c "docker run --rm -v $(pwd)/ballot:/src -v $(pwd)/ballot:/dst ethereum/solc:0.5.12 -o /dst/ --optimize --optimize-runs=2000 --bin --abi --overwrite /src/Ballot.sol"
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=./ballot/Ballot.bin --abi=./ballot/Ballot.abi --pkg=ballot --type=Contract --out=ballot/contract.go

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/Fantom-foundation/go-lachesis/cmd/api-caller/ballot"
	"github.com/Fantom-foundation/go-lachesis/hash"
)

func (g *Generator) ballotCreateContract(admin uint) TxMaker {
	payer := g.Payer(admin)
	return func(client *ethclient.Client) (*types.Transaction, error) {
		_, tx, _, err := ballot.DeployContract(payer, client, [][32]byte{
			ballotProposal("option 1"),
			ballotProposal("option 2"),
			ballotProposal("option 3"),
		})
		if err != nil {
			panic(err)
		}

		return tx, err
	}
}

func (g *Generator) ballotRight(admin uint, addr common.Address, voiter uint) TxMaker {
	payer := g.Payer(admin)
	to := g.Payer(voiter).From
	return func(client *ethclient.Client) (*types.Transaction, error) {
		transactor, err := ballot.NewContractTransactor(addr, client)
		if err != nil {
			panic(err)
		}

		return transactor.GiveRightToVote(payer, to)
	}
}

func (g *Generator) ballotVoite(voiter uint, addr common.Address, proposal int64) TxMaker {
	payer := g.Payer(voiter)
	return func(client *ethclient.Client) (*types.Transaction, error) {
		transactor, err := ballot.NewContractTransactor(addr, client)
		if err != nil {
			panic(err)
		}

		return transactor.Vote(payer, big.NewInt(proposal))
	}
}

func ballotProposal(s string) [32]byte {
	return hash.Of([]byte(s))
}
