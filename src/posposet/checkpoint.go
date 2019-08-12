package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/election"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/vector"
)

// checkpoint is for persistent storing.
type checkpoint struct {
	// fields can change only after a frame is decided
	SuperFrameN       idx.SuperFrame
	LastDecidedFrame  idx.Frame
	LastBlockN        idx.Block
	TotalCap          inter.Stake
	LastConsensusTime inter.Timestamp
	NextMembers       internal.Members
	Balances          hash.Hash
}

/*
 * Poset's methods:
 */

// State saves checkpoint.
func (p *Poset) saveCheckpoint() {
	p.store.SetCheckpoint(p.checkpoint)
}

// Bootstrap restores checkpoint from store.
func (p *Poset) Bootstrap() {
	if p.checkpoint != nil {
		return
	}
	// restore checkpoint
	p.checkpoint = p.store.GetCheckpoint()
	if p.checkpoint == nil {
		p.Fatal("Apply genesis for store first")
	}

	// restore current super-frame
	p.loadSuperFrame()
	p.events = vector.NewIndex(p.Members, p.store.epochTable.VectorIndex)
	p.election = election.New(p.Members, p.LastDecidedFrame+1, p.rootStronglySeeRoot)

	// events reprocessing
	p.handleElection(nil)
}

// GetGenesisHash is a genesis getter.
func (p *Poset) GetGenesisHash() hash.Hash {
	return p.Genesis.Hash()
}