package gossip

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/utils/migration"
)

func (s *Store) Migrate() error {
	versions := migration.NewKvdbIDStore(s.table.Version)
	return s.migrations().Exec(versions)
}

func (s *Store) migrations() *migration.Migration {
	return migration.
		Begin("lachesis-gossip-store").
		Next("remove async data from sync DBs",
			func() error {
				legacyTablePackInfos := table.New(s.mainDb, []byte("p"))
				toolsStore{s}.rmPrefix(legacyTablePackInfos, "serverPool")
				//legacyTablePeers := table.New(s.mainDb, []byte("Z"))
				//toolsStore{s}.rmPrefix(legacyTablePeers, "")
				toolsStore{s}.rmPrefix(s.mainDb, "Z")
				return nil
			}).
		Next("remove legacy genesis field",
			legacyStore1{s}.migrateEraseGenesisField).
		Next("multi-delegations",
			legacyStore1{s}.migrateMultiDelegations).
		Next("adjustable offline pruning time",
			legacyStore1{s}.migrateAdjustableOfflinePeriod)
}

type toolsStore struct {
	*Store
}

func (s toolsStore) rmPrefix(t kvdb.KeyValueStore, prefix string) {
	it := t.NewIteratorWithPrefix([]byte(prefix))
	defer it.Release()

	s.dropTable(it, t)
}

type legacyStore1 struct {
	*Store
}

type legacySfcDelegation1 struct {
	CreatedEpoch idx.Epoch
	CreatedTime  inter.Timestamp

	DeactivatedEpoch idx.Epoch
	DeactivatedTime  inter.Timestamp

	Amount *big.Int

	ToStakerID idx.StakerID
}

func (s legacyStore1) migrateMultiDelegations() error {
	{ // migrate table Delegations
		legacyTableDelegations := table.New(s.mainDb, []byte("3"))
		newKeys := make([][]byte, 0, 10000)
		newValues := make([][]byte, 0, 10000)
		{

			it := legacyTableDelegations.NewIterator()
			defer it.Release()
			for it.Next() {
				delegation := &legacySfcDelegation1{}
				err := rlp.DecodeBytes(it.Value(), delegation)
				if err != nil {
					return errors.Wrap(err, "failed legacy delegation deserialization during migration")
				}

				addr := common.BytesToAddress(it.Key())
				id := sfctype.DelegationID{
					Delegator: addr,
					StakerID:  delegation.ToStakerID,
				}
				newValue, err := rlp.EncodeToBytes(sfctype.SfcDelegation{
					CreatedEpoch:     delegation.CreatedEpoch,
					CreatedTime:      delegation.CreatedTime,
					DeactivatedEpoch: delegation.DeactivatedEpoch,
					DeactivatedTime:  delegation.DeactivatedTime,
					Amount:           delegation.Amount,
				})
				if err != nil {
					return err
				}

				// don't write into DB during iteration
				newKeys = append(newKeys, id.Bytes())
				newValues = append(newValues, newValue)
			}
		}
		{
			it := legacyTableDelegations.NewIterator()
			defer it.Release()
			s.dropTable(it, legacyTableDelegations)
		}
		for i := range newKeys {
			err := legacyTableDelegations.Put(newKeys[i], newValues[i])
			if err != nil {
				return err
			}
		}
	}
	{ // migrate table DelegationOldRewards
		legacyTableDelegationOldRewards := table.New(s.mainDb, []byte("6"))
		newKeys := make([][]byte, 0, 10000)
		newValues := make([][]byte, 0, 10000)
		{
			it := legacyTableDelegationOldRewards.NewIterator()
			defer it.Release()
			for it.Next() {
				addr := common.BytesToAddress(it.Key())
				delegations := s.getSfcDelegationsByAddr(addr, 2)
				if len(delegations) > 1 {
					return errors.New("more than one delegation during multi-delegation migration")
				}
				if len(delegations) == 0 {
					continue
				}
				toStakerID := delegations[0].ID.StakerID
				id := sfctype.DelegationID{
					Delegator: addr,
					StakerID:  toStakerID,
				}

				// don't write into DB during iteration
				newKeys = append(newKeys, id.Bytes())
				newValues = append(newKeys, it.Value())
			}
		}
		{
			it := legacyTableDelegationOldRewards.NewIterator()
			defer it.Release()
			s.dropTable(it, legacyTableDelegationOldRewards)
		}
		for i := range newKeys {
			err := legacyTableDelegationOldRewards.Put(newKeys[i], newValues[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// getSfcDelegationsByAddr returns a lsit of delegations by address
func (s legacyStore1) getSfcDelegationsByAddr(addr common.Address, limit int) []sfctype.SfcDelegationAndID {
	legacyTableDelegations := table.New(s.mainDb, []byte("3"))
	it := legacyTableDelegations.NewIteratorWithPrefix(addr.Bytes())
	defer it.Release()
	res := make([]sfctype.SfcDelegationAndID, 0, limit)
	s.forEachSfcDelegation(it, func(id sfctype.SfcDelegationAndID) bool {
		if limit == 0 {
			return false
		}
		limit--
		res = append(res, id)
		return true
	})
	return res
}

func (s legacyStore1) forEachSfcDelegation(it ethdb.Iterator, do func(sfctype.SfcDelegationAndID) bool) {
	_continue := true
	for _continue && it.Next() {
		delegation := &sfctype.SfcDelegation{}
		err := rlp.DecodeBytes(it.Value(), delegation)
		if err != nil {
			s.Log.Crit("Failed to decode rlp while iteration", "err", err)
		}

		addr := it.Key()[len(it.Key())-sfctype.DelegationIDSize:]
		_continue = do(sfctype.SfcDelegationAndID{
			ID:         sfctype.BytesToDelegationID(addr),
			Delegation: delegation,
		})
	}
}

func (s legacyStore1) migrateEraseGenesisField() error {
	it := s.mainDb.NewIteratorWithPrefix([]byte("G"))
	defer it.Release()
	s.dropTable(it, s.mainDb)
	return nil
}

type legacySfcConstants1 struct {
	ShortGasPowerAllocPerSec uint64
	LongGasPowerAllocPerSec  uint64
	BaseRewardPerSec         *big.Int
}

type legacySfcConstants2 struct {
	ShortGasPowerAllocPerSec uint64
	LongGasPowerAllocPerSec  uint64
	BaseRewardPerSec         *big.Int
	OfflinePenaltyThreshold  struct {
		Num    idx.Block
		Period inter.Timestamp
	}
}

func (s legacyStore1) migrateAdjustableOfflinePeriod() error {
	{ // migrate table SfcConstants
		newKeys := make([][]byte, 0, 10000)
		newValues := make([][]byte, 0, 10000)
		legacyTableSfcConstants := table.New(s.mainDb, []byte("4"))
		{
			it := legacyTableSfcConstants.NewIterator()
			defer it.Release()
			for it.Next() {
				constants := &legacySfcConstants1{}
				err := rlp.DecodeBytes(it.Value(), constants)
				if err != nil {
					return errors.Wrap(err, "failed legacy constants deserialization during migration")
				}

				newConstants := legacySfcConstants2{
					ShortGasPowerAllocPerSec: constants.ShortGasPowerAllocPerSec,
					LongGasPowerAllocPerSec:  constants.LongGasPowerAllocPerSec,
					BaseRewardPerSec:         constants.BaseRewardPerSec,
				}
				newValue, err := rlp.EncodeToBytes(newConstants)
				if err != nil {
					return err
				}

				// don't write into DB during iteration
				newKeys = append(newKeys, it.Key())
				newValues = append(newValues, newValue)
			}
		}
		{
			it := legacyTableSfcConstants.NewIterator()
			defer it.Release()
			s.dropTable(it, legacyTableSfcConstants)
		}
		for i := range newKeys {
			err := legacyTableSfcConstants.Put(newKeys[i], newValues[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
