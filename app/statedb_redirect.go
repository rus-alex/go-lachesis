package app

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

type StateDbRedirector struct {
	*state.StateDB
	flatten kvdb.KeyValueStore
}

func (r *StateDbRedirector) GetState(addr common.Address, loc common.Hash) common.Hash {
	if r.flatten != nil {
		key := append(addr.Bytes(), loc.Bytes()...)
		val, err := r.flatten.Get(key)
		if err != nil {
			panic(err)
		}
		return common.BytesToHash(val)
	}
	return r.StateDB.GetState(addr, loc)
}

func (r *StateDbRedirector) SetState(addr common.Address, loc common.Hash, val common.Hash) {
	if r.flatten != nil {
		key := append(addr.Bytes(), loc.Bytes()...)
		err := r.flatten.Put(key, val.Bytes())
		if err != nil {
			panic(err)
		}
		return
	}
	r.StateDB.SetState(addr, loc, val)
}
