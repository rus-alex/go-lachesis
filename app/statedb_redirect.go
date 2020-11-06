package app

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

// EnabledStateDbRedirection is for testing only.
var EnabledStateDbRedirection = true

type StateDbRedirector struct {
	*state.StateDB
	flatten kvdb.KeyValueStore
}

func (r *StateDbRedirector) GetState(addr common.Address, loc common.Hash) common.Hash {
	if EnabledStateDbRedirection && r.flatten != nil {
		log.Debug("enabled StateDbRedirection")
		key := append(addr.Bytes(), loc.Bytes()...)
		val, err := r.flatten.Get(key)
		if err != nil {
			panic(err)
		}
		return common.BytesToHash(val)
	}
	log.Debug("disabled StateDbRedirection")
	return r.StateDB.GetState(addr, loc)
}

func (r *StateDbRedirector) SetState(addr common.Address, loc common.Hash, val common.Hash) {
	if EnabledStateDbRedirection && r.flatten != nil {
		log.Debug("enabled StateDbRedirection")
		key := append(addr.Bytes(), loc.Bytes()...)
		err := r.flatten.Put(key, val.Bytes())
		if err != nil {
			panic(err)
		}
		return
	}
	log.Debug("disabled StateDbRedirection")
	r.StateDB.SetState(addr, loc, val)
}

func (r *StateDbRedirector) Copy() evmcore.StateDB {
	db := memorydb.New()
	err := kvdb.Copy(r.flatten, db, nil)
	if err != nil {
		panic(err)
	}

	return &StateDbRedirector{
		r.StateDB.Copy(),
		db,
	}
}

func (r *StateDbRedirector) MPT() *state.StateDB {
	return r.StateDB
}
