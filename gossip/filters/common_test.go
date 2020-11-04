package filters

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
)

type stateDbWrapper struct {
	*state.StateDB
}

func (w *stateDbWrapper) Copy() evmcore.StateDB {
	return &stateDbWrapper{
		w.StateDB.Copy(),
	}
}

func (w *stateDbWrapper) MPT() *state.StateDB {
	return w.StateDB
}

func stateNew(root common.Hash, db state.Database, snaps *snapshot.Tree) (evmcore.StateDB, error) {
	sdb, err := state.New(root, db, snaps)
	if err != nil {
		return nil, err
	}

	return &stateDbWrapper{sdb}, nil
}
