package gossip

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-lachesis/app"
)

func TestLegasyStructSerialization(t *testing.T) {
	require := require.New(t)

	for _, src := range []*legacySfcConstants1{
		&legacySfcConstants1{
			ShortGasPowerAllocPerSec: 0,
			LongGasPowerAllocPerSec:  0,
		},
		&legacySfcConstants1{
			ShortGasPowerAllocPerSec: 0xFFFFFFFFFFFFFFFF,
			LongGasPowerAllocPerSec:  0xFFFFFFFFFFFFFFFF,
			BaseRewardPerSec:         big.NewInt(0xFFFFFFFFFFFFFF),
		},
	} {

		dst1 := &legacySfcConstants2{
			ShortGasPowerAllocPerSec: src.ShortGasPowerAllocPerSec,
			LongGasPowerAllocPerSec:  src.LongGasPowerAllocPerSec,
			BaseRewardPerSec:         src.BaseRewardPerSec,
		}

		dst2 := app.SfcConstants{
			ShortGasPowerAllocPerSec: src.ShortGasPowerAllocPerSec,
			LongGasPowerAllocPerSec:  src.LongGasPowerAllocPerSec,
			BaseRewardPerSec:         src.BaseRewardPerSec,
		}

		bytes1, err := rlp.EncodeToBytes(dst1)
		require.NoError(err)

		bytes2, err := rlp.EncodeToBytes(dst2)
		require.NoError(err)

		require.EqualValues(bytes1, bytes2)
	}
}
