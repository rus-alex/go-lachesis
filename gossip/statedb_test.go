package gossip

// Simple ballot contract
//go:generate bash -c "docker run --rm -v $(pwd)/ballot:/src ethereum/solc:0.5.12 -o /src/solc/ --optimize --optimize-runs=2000 --bin --abi --allow-paths /src --overwrite /src/Ballot.sol"
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --bin=ballot/solc/Ballot.bin --abi=ballot/solc/Ballot.abi --pkg=ballot --type=Contract --out=ballot/contract.go

import (
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	eth "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-lachesis/app"
	"github.com/Fantom-foundation/go-lachesis/gossip/ballot"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/leveldb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/nokeyiserr"
	"github.com/Fantom-foundation/go-lachesis/logger"
	"github.com/Fantom-foundation/go-lachesis/utils"
)

func BenchmarkPureDB(b *testing.B) {
	logger.SetTestMode(b)

	data := hash.FakeEvents(100)

	b.Run("MemoryDB", func(b *testing.B) {
		dbs := memorydb.NewProducer("")
		benchmarkPureDB(b, dbs, data)
	})

	b.Run("LevelDB", func(b *testing.B) {
		dbdir, err := ioutil.TempDir("", "benchmark_puredb*")
		require.NoError(b, err)

		dbs := leveldb.NewProducer(dbdir)
		defer os.RemoveAll(dbdir)

		benchmarkPureDB(b, dbs, data)
	})
}

func benchmarkPureDB(b *testing.B, dbs kvdb.DbProducer, data hash.Events) {
	require := require.New(b)
	db := dbs.OpenDb(uniqName())
	defer db.Close()

	x := len(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// write
		key := data[i%x].Bytes()
		val := data[(i+1)%x].Bytes()
		err := db.Put(key, val)
		require.NoError(err)
		// read
		if i < (x / 10) {
			continue
		}
		key = data[(i-10)%x].Bytes()
		_, err = db.Get(key)
		require.NoError(err)
	}
}

func BenchmarkStateDB(b *testing.B) {
	logger.SetTestMode(b)

	data := make([]common.Hash, 100)
	for i := 0; i < len(data); i++ {
		data[i] = hash.FakeHash(int64(i))
	}

	b.Run("MemoryDB", func(b *testing.B) {
		dbs := memorydb.NewProducer("")
		benchmarkStateDB(b, dbs, data)
	})

	b.Run("LevelDB", func(b *testing.B) {
		dbdir, err := ioutil.TempDir("", "benchmark_statedb*")
		defer os.RemoveAll(dbdir)
		require.NoError(b, err)

		dbs := leveldb.NewProducer(dbdir)

		benchmarkStateDB(b, dbs, data)
	})

	b.Run("CachedLevelDB", func(b *testing.B) {
		dbdir, err := ioutil.TempDir("", "benchmark_statedb*")
		require.NoError(b, err)
		defer os.RemoveAll(dbdir)

		dbs := flushable.NewSyncedPool(
			leveldb.NewProducer(dbdir))
		defer dbs.Flush([]byte("0"))

		benchmarkStateDB(b, dbs, data)
	})
}

func benchmarkStateDB(b *testing.B, dbs kvdb.DbProducer, data []common.Hash) {
	db := dbs.OpenDb(uniqName())
	stateStore := state.NewDatabase(
		rawdb.NewDatabase(
			nokeyiserr.Wrap(
				db)))

	b.Run("OverKVDB", func(b *testing.B) {
		stateDB, err := state.New(common.Hash{}, stateStore, nil)
		require.NoError(b, err)

		flatten := dbs.OpenDb(uniqName())
		defer flatten.Close()

		KVDB := app.NewStateDbRedirector(stateDB, flatten)
		benchmarkStateDbOver(b, KVDB, data)
	})

	b.Run("OverMPT", func(b *testing.B) {
		stateDB, err := state.New(common.Hash{}, stateStore, nil)
		require.NoError(b, err)

		MPT := app.NewStateDbRedirector(stateDB, nil)
		benchmarkStateDbOver(b, MPT, data)
	})
}

func benchmarkStateDbOver(b *testing.B, stateDB *app.StateDbRedirector, data []common.Hash) {
	require := require.New(b)

	b.ResetTimer()
	defer b.StopTimer()

	x := len(data)
	for i := 0; i < b.N; i++ {
		// write
		loc := data[i%x]
		addr := common.BytesToAddress(loc[:common.AddressLength])
		val := data[(i+1)%x]
		stateDB.SetState(addr, loc, val)
		// read
		if i < (x / 10) {
			continue
		}
		loc = data[(i-10)%x]
		addr = common.BytesToAddress(loc[:common.AddressLength])
		val = stateDB.GetState(addr, loc)
		require.NotEmpty(val)
	}
	root, err := stateDB.Commit(true)
	require.NoError(err)
	require.NotEmpty(root)
}

func BenchmarkStateDbWithBallot(b *testing.B) {
	logger.SetLevel("warn")
	//logger.SetTestMode(b)

	b.Run("overMPT", func(b *testing.B) {
		env := newTestEnv(false)
		defer env.Close()
		benchmarkStateDbWithBallot(b, env)
	})

	b.Run("Flattened", func(b *testing.B) {
		env := newTestEnv(true)
		defer env.Close()
		benchmarkStateDbWithBallot(b, env)
	})
}

func benchmarkStateDbWithBallot(b *testing.B, env *testEnv) {
	require := require.New(b)

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

func uniqName() string {
	return hash.FakeHash(rand.Int63()).Hex()
}
