package evmcore

import (
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
)

type StateDbRedirector struct {
	*state.StateDB
	flattened kvdb.KeyValueStore
}

func (r *StateDbRedirector) GetState(addr common.Address, loc common.Hash) common.Hash {
	if r.flattened != nil {
		key := append(addr.Bytes(), loc.Bytes()...)
		val, err := r.flattened.Get(key)
		if err != nil {
			panic(err)
		}
		return common.BytesToHash(val)
	}
	return r.StateDB.GetState(addr, loc)
}

func (r *StateDbRedirector) SetState(addr common.Address, loc common.Hash, val common.Hash) {
	if r.flattened != nil {
		key := append(addr.Bytes(), loc.Bytes()...)
		err := r.flattened.Put(key, val.Bytes())
		if err != nil {
			panic(err)
		}
	}
	r.StateDB.SetState(addr, loc, val)
}
