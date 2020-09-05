package app

import (
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

type (
	Config struct {
		Net lachesis.Config
		StoreConfig

		// TxIndex enables indexing transactions and receipts
		TxIndex bool
		// EpochDowntimeIndex enables indexing downtime by epoch
		EpochDowntimeIndex bool
		// EpochActiveValidationScoreIndex enables indexing validation score by epoch
		EpochActiveValidationScoreIndex bool
	}

	// StoreConfig is a config for store db.
	StoreConfig struct {
		// Cache size for Receipts.
		ReceiptsCacheSize int
		// Cache size for Stakers.
		StakersCacheSize int
		// Cache size for Delegations.
		DelegationsCacheSize int
		// Cache size for EpochStats.
		EpochStatsCacheSize int
	}
)

// DefaultConfig for product.
func DefaultConfig(network lachesis.Config) Config {
	return Config{
		Net:                             network,
		TxIndex:                         true,
		EpochDowntimeIndex:              false,
		EpochActiveValidationScoreIndex: false,
		StoreConfig:                     DefaultStoreConfig(),
	}
}

// DefaultStoreConfig for product.
func DefaultStoreConfig() StoreConfig {
	return StoreConfig{
		ReceiptsCacheSize:    100,
		DelegationsCacheSize: 4000,
		StakersCacheSize:     4000,
		EpochStatsCacheSize:  100,
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return StoreConfig{
		ReceiptsCacheSize:    100,
		DelegationsCacheSize: 400,
		StakersCacheSize:     400,
		EpochStatsCacheSize:  100,
	}
}
