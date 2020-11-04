package evmcore

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/state/snapshot"
	"github.com/ethereum/go-ethereum/core/types"
)

// StateDB is a subset of *state.StateDB methods.
type StateDB interface {
	// vm.StateDB part
	CreateAccount(common.Address)

	SubBalance(common.Address, *big.Int)
	AddBalance(common.Address, *big.Int)
	GetBalance(common.Address) *big.Int
	SetBalance(common.Address, *big.Int)

	GetNonce(common.Address) uint64
	SetNonce(common.Address, uint64)

	GetCodeHash(common.Address) common.Hash
	GetCode(common.Address) []byte
	SetCode(common.Address, []byte)
	GetCodeSize(common.Address) int

	AddRefund(uint64)
	SubRefund(uint64)
	GetRefund() uint64

	GetCommittedState(common.Address, common.Hash) common.Hash
	GetState(common.Address, common.Hash) common.Hash
	SetState(common.Address, common.Hash, common.Hash)

	Suicide(common.Address) bool
	HasSuicided(common.Address) bool

	Exist(common.Address) bool
	Empty(common.Address) bool

	RevertToSnapshot(int)
	Snapshot() int

	AddLog(*types.Log)
	AddPreimage(common.Hash, []byte)

	ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) error

	// other part

	Prepare(thash, bhash common.Hash, ti int)
	Finalise(deleteEmptyObjects bool)
	IntermediateRoot(deleteEmptyObjects bool) common.Hash
	GetLogs(hash common.Hash) []*types.Log
	BlockHash() common.Hash
	TxIndex() int
	Commit(deleteEmptyObjects bool) (common.Hash, error)
	Copy() StateDB
	Database() state.Database
	Error() error
	StorageTrie(common.Address) state.Trie
	GetProof(common.Address) ([][]byte, error)
	SetStorage(common.Address, map[common.Hash]common.Hash)
	GetStorageProof(a common.Address, key common.Hash) ([][]byte, error)

	// workaround

	MPT() *state.StateDB
}

type stateDbWrapper struct {
	*state.StateDB
}

func (w *stateDbWrapper) Copy() StateDB {
	return &stateDbWrapper{
		w.StateDB.Copy(),
	}
}

func (w *stateDbWrapper) MPT() *state.StateDB {
	return w.StateDB
}

func stateNew(root common.Hash, db state.Database, snaps *snapshot.Tree) (StateDB, error) {
	sdb, err := state.New(root, db, snaps)
	if err != nil {
		return nil, err
	}

	return &stateDbWrapper{sdb}, nil
}
