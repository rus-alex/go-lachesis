package gossip

// Simple ballot contract
//go:generate bash -c "docker run --rm -v $(pwd)/ballot:/src ethereum/solc:0.5.12 -o /src/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src --overwrite /src/Ballot.sol"
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=ballot/solc/Ballot.bin --abi=ballot/solc/Ballot.abi --pkg=ballot --type=Contract --out=ballot/contract.go

import (
	"math/big"
	"testing"

	eth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-lachesis/gossip/ballot"
	"github.com/Fantom-foundation/go-lachesis/logger"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

func BenchmarkStateDB(b *testing.B) {
	logger.SetTestMode(b)
	require := require.New(b)

	env := newTestEnv()
	defer env.Close()

	proposals := [][32]byte{
		ballotOption("Option 1"),
		ballotOption("Option 2"),
		ballotOption("Option 3"),
	}

	// contract deploy
	addr, tx, cBallot, err := ballot.DeployContract(env.Payer(1), env, proposals)
	require.NoError(err)
	require.NotNil(cBallot)
	r := env.ApplyBlock(nextEpoch, tx)
	require.Equal(addr, r[0].ContractAddress)

	admin, err := cBallot.Chairperson(env.ReadOnly())
	require.NoError(err)
	require.Equal(env.Address(1), admin)

	count := b.N

	// Init accounts
	txs := make([]*eth.Transaction, 0, count-1)
	for i := 2; i <= count; i++ {
		tx := env.Transfer(1, i, utils.ToFtm(10))
		require.NoError(err)
		txs = append(txs, tx)
	}
	env.ApplyBlock(nextEpoch, txs...)

	// GiveRightToVote
	txs = make([]*eth.Transaction, 0, count)
	for i := 1; i <= count; i++ {
		tx, err := cBallot.GiveRightToVote(env.Payer(1), env.Address(i))
		require.NoError(err)
		txs = append(txs, tx)
	}
	env.ApplyBlock(nextEpoch, txs...)

	// Vote
	txs = make([]*eth.Transaction, 0, count)
	for i := 1; i <= count; i++ {
		proposal := big.NewInt(int64(i % len(proposals)))
		tx, err := cBallot.Vote(env.Payer(i), proposal)
		require.NoError(err)
		txs = append(txs, tx)
	}
	env.ApplyBlock(nextEpoch, txs...)

	// Winer
	_, err = cBallot.WinnerName(env.ReadOnly())
	require.NoError(err)
}

func ballotOption(str string) (res [32]byte) {
	buf := []byte(str)
	if len(buf) > 32 {
		panic("string too long")
	}
	copy(res[:], buf)
	return
}
