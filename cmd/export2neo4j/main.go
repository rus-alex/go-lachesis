package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/params"

	"github.com/Fantom-foundation/go-lachesis/flags"
	_ "github.com/Fantom-foundation/go-lachesis/version"
)

var (
	// Git SHA1 commit hash of the release (set via linker flags).
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = flags.NewApp(gitCommit, gitDate, "the go-lachesis command line interface")
)

// init the CLI app.
func init() {
	app.Action = utils.MigrateFlags(exportChain)
	app.Version = params.VersionWithCommit(gitCommit, gitDate)
	app.Flags = append(app.Flags,
		DataDirFlag,
		Neo4jFlag,
	)
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
