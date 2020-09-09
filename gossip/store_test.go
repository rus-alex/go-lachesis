package gossip

import (
	"time"

	"github.com/Fantom-foundation/go-lachesis/app"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/leveldb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

func cachedStore() *Store {
	mems := memorydb.NewProducer("", withDelay)
	dbs := flushable.NewSyncedPool(mems)
	adb := app.NewStore(dbs, app.LiteStoreConfig())

	cfg := LiteStoreConfig()
	s := NewStore(dbs, cfg, adb)

	return s
}

func nonCachedStore() *Store {
	mems := memorydb.NewProducer("", withDelay)
	dbs := flushable.NewSyncedPool(mems)
	adb := app.NewStore(dbs, app.LiteStoreConfig())

	cfg := StoreConfig{}
	s := NewStore(dbs, cfg, adb)
	return s
}

func realStore(dir string) *Store {
	disk := leveldb.NewProducer(dir)
	dbs := flushable.NewSyncedPool(disk)
	adb := app.NewStore(dbs, app.LiteStoreConfig())

	cfg := LiteStoreConfig()
	s := NewStore(dbs, cfg, adb)
	return s
}

func withDelay(db kvdb.KeyValueStore) kvdb.KeyValueStore {
	mem, ok := db.(*memorydb.Database)
	if ok {
		mem.SetDelay(time.Millisecond)

	}

	return db
}
