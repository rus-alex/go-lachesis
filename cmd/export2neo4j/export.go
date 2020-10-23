package main

import (
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"gopkg.in/urfave/cli.v1"

	appdb "github.com/Fantom-foundation/go-lachesis/app"
	"github.com/Fantom-foundation/go-lachesis/gossip"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/integration"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
)

// statsReportLimit is the time limit during import and export after which we
// always print out progress. This avoids the user wondering what's going on.
const statsReportLimit = 8 * time.Second

var (
	// DataDirFlag defines directory to store Lachesis state and user's wallets
	DataDirFlag = utils.DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases of lachesis node",
		Value: utils.DirectoryString(DefaultDataDir()),
	}

	// Neo4jFlag defines db uri
	Neo4jFlag = cli.StringFlag{
		Name:  "neo4j",
		Usage: "Neo4j url",
		Value: DefaultDb(),
	}
)

func exportChain(ctx *cli.Context) error {
	dataDir := DefaultDataDir()
	if ctx.GlobalIsSet(utils.DataDirFlag.Name) {
		dataDir = ctx.GlobalString(DataDirFlag.Name)
	}
	gdb := makeGossipStore(dataDir)
	defer gdb.Close()

	dbUri := DefaultDb()
	if ctx.GlobalIsSet(Neo4jFlag.Name) {
		dbUri = ctx.GlobalString(Neo4jFlag.Name)
	}
	db, err := neo4j.NewDriver(dbUri, neo4j.NoAuth(), func(c *neo4j.Config) {
		c.Encrypted = false
	})
	if err != nil {
		return err
	}
	defer db.Close()

	from := idx.Epoch(1)
	if len(ctx.Args()) > 1 {
		n, err := strconv.ParseUint(ctx.Args().Get(1), 10, 32)
		if err != nil {
			return err
		}
		from = idx.Epoch(n)
	}
	to := idx.Epoch(0)
	if len(ctx.Args()) > 2 {
		n, err := strconv.ParseUint(ctx.Args().Get(2), 10, 32)
		if err != nil {
			return err
		}
		to = idx.Epoch(n)
	}

	err = exportTo(db, gdb, from, to)
	if err != nil {
		utils.Fatalf("Export error: %v\n", err)
	}

	return nil
}

func makeGossipStore(dataDir string) *gossip.Store {
	dbs := flushable.NewSyncedPool(integration.DBProducer(dataDir))
	gdb := gossip.NewStore(dbs, gossip.LiteStoreConfig(), appdb.LiteStoreConfig())
	gdb.SetName("gossip-db")
	return gdb
}

// exportTo writer the active chain.
func exportTo(db neo4j.Driver, gdb *gossip.Store, from, to idx.Epoch) (err error) {
	start, reported := time.Now(), time.Time{}

	session, err := db.Session(neo4j.AccessModeWrite)
	if err != nil {
		return
	}
	defer session.Close()

	var (
		counter int
		last    hash.Event
	)
	gdb.ForEachEvent(from, func(event *inter.Event) bool {
		if to >= from && event.Epoch > to {
			return false
		}

		log.Info(">>>", "event", event.Hash())

		_, err := session.WriteTransaction(func(ctx neo4j.Transaction) (interface{}, error) {
			result, err := ctx.Run(
				"CREATE (e:Event) SET e.hash = $hash",
				map[string]interface{}{
					"hash": event.Hash().String(),
				})
			if err != nil {
				return nil, err
			}
			return nil, result.Err()
		})
		if err != nil {
			log.Error("<<<", "err", err)
			return false
		}

		counter++
		last = event.Hash()
		if counter%100 == 1 && time.Since(reported) >= statsReportLimit {
			log.Info("Exporting events", "last", last.String(), "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
		return true
	})

	log.Info("Exported events", "last", last.String(), "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))
	return
}
