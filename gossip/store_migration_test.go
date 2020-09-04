package gossip

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-lachesis/app"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

func TestLegacyStructSerialization(t *testing.T) {
	require := require.New(t)

	for i, src := range []*legacySfcConstants1{
		&legacySfcConstants1{
			ShortGasPowerAllocPerSec: 0,
			LongGasPowerAllocPerSec:  0,
		},
		&legacySfcConstants1{
			ShortGasPowerAllocPerSec: 0xFFFFFFFFFFFFFFFF,
			LongGasPowerAllocPerSec:  0xFFFFFFFFFFFFFFFF,
			BaseRewardPerSec:         big.NewInt(0xFFFFFFFFFFFFFF),
		},
	} {

		dst1 := &legacySfcConstants2{
			ShortGasPowerAllocPerSec: src.ShortGasPowerAllocPerSec,
			LongGasPowerAllocPerSec:  src.LongGasPowerAllocPerSec,
			BaseRewardPerSec:         src.BaseRewardPerSec,
		}

		dst2 := app.SfcConstants{
			ShortGasPowerAllocPerSec: src.ShortGasPowerAllocPerSec,
			LongGasPowerAllocPerSec:  src.LongGasPowerAllocPerSec,
			BaseRewardPerSec:         src.BaseRewardPerSec,
		}

		bytes1, err := rlp.EncodeToBytes(dst1)
		require.NoError(err)

		bytes2, err := rlp.EncodeToBytes(dst2)
		require.NoError(err)

		require.EqualValues(bytes1, bytes2, i)
	}
}

func TestDropTableApproachesAreTheSame(t *testing.T) {
	require := require.New(t)

	db := memorydb.New()
	defer db.Close()

	count := func(t kvdb.KeyValueStore) int {
		it := t.NewIterator()
		defer it.Release()

		var i int
		for i = 0; it.Next(); i++ {
		}
		return i
	}

	rmByName := func(db, t kvdb.KeyValueStore, tname string) {
		it := db.NewIteratorWithPrefix([]byte(tname))
		defer it.Release()
		dropTable(it, db)
	}

	rmTable := func(db, t kvdb.KeyValueStore, tname string) {
		it := t.NewIterator()
		defer it.Release()
		dropTable(it, t)
	}

	var testdata = []string{
		"",
		"0",
		"sdjfuviaiew",
		"_",
	}

	for name, approach := range map[string]func(db, t kvdb.KeyValueStore, name string){
		"drop by name": rmByName,
		"drop table":   rmTable,
	} {
		tableName := "table" + name
		t := table.New(db, []byte(tableName))

		for _, x := range testdata {
			key := []byte(x)
			val := []byte(x)
			t.Put(key, val)
		}
		require.Equal(len(testdata), count(t), name)

		approach(db, t, tableName)
		require.Equal(0, count(t), name)
	}
}

func dropTable(it ethdb.Iterator, t kvdb.KeyValueStore) {
	var s *Store
	s.dropTable(it, t)
}
