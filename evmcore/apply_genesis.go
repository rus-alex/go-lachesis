// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evmcore

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

// StateDB is a subset of *state.StateDB methods.

// ApplyGenesis writes or updates the genesis block in db.
func ApplyGenesis(statedb StateDB, net *lachesis.Config) (*EvmBlock, error) {
	if net == nil {
		return nil, ErrNoGenesis
	}

	// state
	for addr, account := range net.Genesis.Alloc.Accounts {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}

	// initial block
	root, err := statedb.Commit(true)
	if err != nil {
		return nil, err
	}
	block := genesisBlock(net, root)

	return block, nil
}

// genesisBlock makes genesis block with pretty hash.
func genesisBlock(net *lachesis.Config, root common.Hash) *EvmBlock {

	block := &EvmBlock{
		EvmHeader: EvmHeader{
			Number:   big.NewInt(0),
			Time:     net.Genesis.Time,
			GasLimit: math.MaxUint64,
			Root:     root,
			TxHash:   inter.EmptyTxHash,
		},
	}

	return block
}

// MustApplyGenesis writes the genesis block and state to db, panicking on error.
func MustApplyGenesis(net *lachesis.Config, statedb StateDB) *EvmBlock {
	block, err := ApplyGenesis(statedb, net)
	if err != nil {
		log.Crit("ApplyGenesis", "err", err)
	}
	return block
}
