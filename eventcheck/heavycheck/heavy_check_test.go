package heavycheck

import (
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/Fantom-foundation/go-lachesis/eventcheck/epochcheck"
	"github.com/Fantom-foundation/go-lachesis/eventcheck/tests"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
)

type testDagReader struct {
	epochPubKeys map[idx.StakerID]common.Address
	epoch        idx.Epoch
}

// GetEpochPubKeys returns pubkeys of a special testing dagReader
func (t *testDagReader) GetEpochPubKeys() (map[idx.StakerID]common.Address, idx.Epoch) {
	return t.epochPubKeys, t.epoch
}

// getTestDagReaders creates dag readers for testing purposes
func getTestDagReaders(t *testing.T) []testDagReader {
	var tdr []testDagReader
	epochPubKeys1 := map[idx.StakerID]common.Address{}
	epochPubKeys1[1] = tests.NewWallet().Address
	epochPubKeys1[2] = tests.NewWallet().Address

	epochPubKeys2 := map[idx.StakerID]common.Address{}
	epochPubKeys2[11] = tests.NewWallet().Address
	epochPubKeys2[22] = tests.NewWallet().Address
	epochs := []idx.Epoch{0, 1, 10}
	for _, epoch := range epochs {
		tdr = append(tdr, testDagReader{epochPubKeys1, epoch})
		tdr = append(tdr, testDagReader{epochPubKeys2, epoch})
	}
	return tdr
}

// TestHeavyCheck main testing func
func TestHeavyCheck(t *testing.T) {
	configs := tests.DagConfigs()
	dagReaders := getTestDagReaders(t)
	for _, cfg := range configs {
		for _, dagReader := range dagReaders {
			net := lachesis.FakeNetConfig(genesis.FakeAccounts(0, 5, big.NewInt(0), pos.StakeToBalance(1)))
			ledgerID := net.EvmChainConfig().ChainID

			testEvents := tests.Events()
			testEvent := testEvents[0]
			tw := tests.NewWallet()
			sig, err := crypto.Sign(testEvent.Hash().Bytes(), &tw.PrivateKey)
			require.Nil(t, err)
			testEvent.Sig = sig
			signedHash := crypto.Keccak256(testEvent.DataToSign())
			pk, err := crypto.SigToPub(signedHash, testEvent.Sig)
			require.Nil(t, err)
			dagReader.epochPubKeys[1] = crypto.PubkeyToAddress(*pk)

			checker := NewDefault(cfg, &dagReader, types.NewEIP155Signer(ledgerID))

			testChecker(t, checker, testEvents)
		}
	}

	net := lachesis.FakeNetConfig(genesis.FakeAccounts(0, 5, big.NewInt(0), pos.StakeToBalance(1)))
	ledgerID := net.EvmChainConfig().ChainID
	checker := NewDefault(configs[1], &dagReaders[0], types.NewEIP155Signer(ledgerID))
	testOverloaded(t, checker)
}

// testChecker runs set of main tests for a checker
func testChecker(t *testing.T, checker *Checker, testEvents []*inter.Event) {
	checker.Start()
	defer checker.Stop()

	for _, event := range testEvents {
		testValidate(t, checker, event)
		testEnqueue(t, checker, inter.Events{event})
	}
}

// testOverloaded is a small testing func for a Overloaded func
func testOverloaded(t *testing.T, checker *Checker) {
	require.False(t, checker.Overloaded())
	taskDatas := makeTaskData(maxQueuedTasks - len(checker.tasksQ))
	for _, taskData := range taskDatas {
		checker.tasksQ <- taskData
	}
	require.True(t, checker.Overloaded())
}

// makeTaskData creates array of taskData objects
func makeTaskData(num int) []*TaskData {
	var taskDatas []*TaskData
	for i := 0; i < num; i++ {
		td := TaskData{}
		td.onValidated = func(ArbitraryTaskData) {}
		taskDatas = append(taskDatas, &td)
	}
	return taskDatas
}

type TestArbitraryTaskData struct {
}

// GetEvents is just an implementation
func (t *TestArbitraryTaskData) GetEvents() inter.Events {
	return nil
}

// GetResult is just an implementation
func (t *TestArbitraryTaskData) GetResult() []error {
	return nil
}

// GetOnValidatedFn is just an implementation
func (t *TestArbitraryTaskData) GetOnValidatedFn() OnValidatedFn {
	return nil
}

// testEnqueue tests Enqueue function
func testEnqueue(t *testing.T, checker *Checker, event inter.Events) {
	onValidatedFns := []func(ArbitraryTaskData){func(ArbitraryTaskData) {}}
	for _, fn := range onValidatedFns {
		err := checker.Enqueue(event, fn)
		require.Nil(t, err)
	}
}

// testValidate tests validate function
func testValidate(t *testing.T, checker *Checker, event *inter.Event) {
	err := checker.Validate(event)
	if event == nil {
		require.Equal(t, ErrEventIsNil, err)
		return
	}

	addrs, epoch := checker.reader.GetEpochPubKeys()
	if event.Epoch != epoch {
		require.Equal(t, epochcheck.ErrNotRelevant, err)
		return
	}

	addr, ok := addrs[event.Creator]
	if !ok {
		require.Equal(t, epochcheck.ErrAuth, err)
		return
	}

	if !event.VerifySignature(addr) {
		require.Equal(t, ErrWrongEventSig, err)
		return
	}

	for _, tx := range event.Transactions {
		_, err2 := types.Sender(checker.txSigner, tx)
		if err2 != nil {
			require.Equal(t, ErrMalformedTxSig, err)
			return
		}
	}

	if event.TxHash != types.DeriveSha(event.Transactions) {
		require.Equal(t, ErrWrongTxHash, err)
		return
	}

	require.Equal(t, nil, err)
}

// TestTaskData is here for test coverage purposes. TaskData's getters actually has no logic for now
func TestTaskData(t *testing.T) {
	td := TaskData{}
	events := inter.Events{inter.NewEvent()}
	result := []error{errors.New("test err")}
	onVal := func(ArbitraryTaskData) {}

	td.Events = events
	td.Result = result
	td.onValidated = onVal

	onvalFnRet := td.GetOnValidatedFn()

	sf1 := reflect.ValueOf(onVal)
	sf2 := reflect.ValueOf(onvalFnRet)

	require.Equal(t, events, td.GetEvents())
	require.Equal(t, result, td.GetResult())
	require.Equal(t, sf1.Pointer(), sf2.Pointer())
}