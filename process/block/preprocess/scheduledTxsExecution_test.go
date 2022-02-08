package preprocess

import (
	"encoding/json"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	storageMocks "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilTxProcessor(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		nil,
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilTxProcessor, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilTxCoordinator(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		nil,
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilTransactionCoordinator, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilStorer(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		nil,
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilMarshaller(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		nil,
		&mock.ShardCoordinatorStub{},
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionNilShardCoordinator(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		nil,
	)

	assert.True(t, check.IfNil(scheduledTxsExec))
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestScheduledTxsExecution_NewScheduledTxsExecutionOk(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, err := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, scheduledTxsExec)
}

func TestScheduledTxsExecution_InitShouldWork(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	scheduledTxsExec.AddScheduledTx([]byte("txHash2"), &transaction.Transaction{Nonce: 1})
	scheduledTxsExec.AddScheduledTx([]byte("txHash3"), &transaction.Transaction{Nonce: 2})

	assert.Equal(t, 3, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 3, len(scheduledTxsExec.scheduledTxs))

	scheduledTxsExec.Init()

	assert.Equal(t, 0, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 0, len(scheduledTxsExec.scheduledTxs))
}

func TestScheduledTxsExecution_AddShouldWork(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	res := scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	assert.True(t, res)
	assert.Equal(t, 1, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 1, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	assert.False(t, res)
	assert.Equal(t, 1, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 1, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.AddScheduledTx([]byte("txHash2"), &transaction.Transaction{Nonce: 1})
	assert.True(t, res)
	assert.Equal(t, 2, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 2, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.AddScheduledTx([]byte("txHash3"), &transaction.Transaction{Nonce: 1})
	assert.True(t, res)
	assert.Equal(t, 3, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 3, len(scheduledTxsExec.scheduledTxs))

	res = scheduledTxsExec.AddScheduledTx([]byte("txHash2"), &transaction.Transaction{Nonce: 2})
	assert.False(t, res)
	assert.Equal(t, 3, len(scheduledTxsExec.mapScheduledTxs))
	assert.Equal(t, 3, len(scheduledTxsExec.scheduledTxs))
}

func TestScheduledTxsExecution_ExecuteShouldErrMissingTransaction(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.True(t, errors.Is(err, process.ErrMissingTransaction))
}

func TestScheduledTxsExecution_ExecuteShouldErr(t *testing.T) {
	t.Parallel()

	localError := errors.New("error")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, localError
			},
		},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.Equal(t, localError, err)
}

func TestScheduledTxsExecution_ExecuteShouldWorkOnErrFailedTransaction(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, process.ErrFailedTransaction
			},
		},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.Nil(t, err)
}

func TestScheduledTxsExecution_ExecuteShouldWork(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, nil
			},
		},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	err := scheduledTxsExec.Execute([]byte("txHash1"))
	assert.Nil(t, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrNilHaveTimeHandler(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.ExecuteAll(nil)
	assert.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrTimeIsOut(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(-1) }
	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction)
	assert.Equal(t, process.ErrTimeIsOut, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldErrFailedTransaction(t *testing.T) {
	t.Parallel()

	localError := errors.New("error")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, localError
			},
		},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(100) }
	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction)
	assert.Equal(t, localError, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldWorkOnErrFailedTransaction(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, process.ErrFailedTransaction
			},
		},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(100) }
	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction)
	assert.Nil(t, err)
}

func TestScheduledTxsExecution_ExecuteAllShouldWork(t *testing.T) {
	t.Parallel()

	numTxsExecuted := 0
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				numTxsExecuted++
				return vmcommon.Ok, nil
			},
		},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	haveTimeFunction := func() time.Duration { return time.Duration(100) }
	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})
	scheduledTxsExec.AddScheduledTx([]byte("txHash2"), &transaction.Transaction{Nonce: 1})
	scheduledTxsExec.AddScheduledTx([]byte("txHash3"), &transaction.Transaction{Nonce: 2})

	err := scheduledTxsExec.ExecuteAll(haveTimeFunction)
	assert.Nil(t, err)
	assert.Equal(t, 3, numTxsExecuted)
}

func TestScheduledTxsExecution_executeShouldErr(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.execute(nil)
	assert.True(t, errors.Is(err, process.ErrWrongTypeAssertion))
}

func TestScheduledTxsExecution_executeShouldWork(t *testing.T) {
	t.Parallel()

	response := errors.New("response")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				return vmcommon.Ok, response
			},
		},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.execute(&transaction.Transaction{Nonce: 0})
	assert.Equal(t, response, err)
}

func TestScheduledTxsExecution_computeScheduledSCRsShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorStub{
		SameShardCalled: func(_, _ []byte) bool {
			return false
		},
	}

	mapAllIntermediateTxsBeforeScheduledExecution := map[block.Type]map[string]data.TransactionHandler{
		0: {
			"txHash1": &transaction.Transaction{Nonce: 1},
			"txHash2": &transaction.Transaction{Nonce: 2},
		},
	}
	mapAllIntermediateTxsAfterScheduledExecution := map[block.Type]map[string]data.TransactionHandler{
		1: {
			"txHash3": &transaction.Transaction{Nonce: 3},
			"txHash4": &transaction.Transaction{Nonce: 4},
		},
	}

	t.Run("nil maps, empty scheduled scrs", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			shardCoordinator,
		)

		scheduledTxsExec.ComputeScheduledIntermediateTxs(nil, nil)

		assert.Equal(t, 0, len(scheduledTxsExec.GetMapScheduledIntermediateTxs()))
	})
	t.Run("nil map after txs execition, empty scheduled scrs", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			shardCoordinator,
		)

		scheduledTxsExec.ComputeScheduledIntermediateTxs(mapAllIntermediateTxsBeforeScheduledExecution, nil)

		assert.Equal(t, 0, len(scheduledTxsExec.GetMapScheduledIntermediateTxs()))
	})
	t.Run("nil map after txs execition, empty scheduled scrs", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			shardCoordinator,
		)

		localMapAllIntermediateTxsAfterScheduledExecution := map[block.Type]map[string]data.TransactionHandler{
			0: {
				"txHash1": &transaction.Transaction{Nonce: 1},
				"txHash2": &transaction.Transaction{Nonce: 2},
			},
		}
		scheduledTxsExec.ComputeScheduledIntermediateTxs(
			mapAllIntermediateTxsBeforeScheduledExecution,
			localMapAllIntermediateTxsAfterScheduledExecution,
		)

		assert.Equal(t, 0, len(scheduledTxsExec.mapScheduledIntermediateTxs))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			shardCoordinator,
		)

		scheduledTxsExec.ComputeScheduledIntermediateTxs(
			mapAllIntermediateTxsBeforeScheduledExecution,
			mapAllIntermediateTxsAfterScheduledExecution,
		)

		mapScheduledSCRs := scheduledTxsExec.GetMapScheduledIntermediateTxs()
		assert.Equal(t, 1, len(mapScheduledSCRs))
		assert.Equal(t, 2, len(mapScheduledSCRs[1]))
	})
}

func TestScheduledTxsExecution_getAllIntermediateTxsAfterScheduledExecution(t *testing.T) {
	t.Parallel()

	allTxsBeforeExec := map[block.Type]map[string]data.TransactionHandler{
		0: {
			"txHash1": &transaction.Transaction{Nonce: 1},
			"txHash2": &transaction.Transaction{Nonce: 2},
		},
	}
	allTxsAfterExec := map[string]data.TransactionHandler{
		"txHash3": &transaction.Transaction{Nonce: 3},
		"txHash4": &transaction.Transaction{Nonce: 4},
	}

	t.Run("not already existing txs, different shard", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return false
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			0,
		)

		assert.Equal(t, 2, len(scrsInfo))
	})
	t.Run("not already existing txs, same shard, scr", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return true
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			block.SmartContractResultBlock,
		)

		assert.Equal(t, 0, len(scrsInfo))
	})
	t.Run("not already existing txs, same shard, receipt", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return true
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			block.ReceiptBlock,
		)

		assert.Equal(t, 0, len(scrsInfo))
	})
	t.Run("not already existing txs, same shard, transaction", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return true
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			block.TxBlock,
		)

		assert.Equal(t, 2, len(scrsInfo))
	})
	t.Run("not already existing txs, same shard, invalid", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return true
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			block.InvalidBlock,
		)

		assert.Equal(t, 2, len(scrsInfo))
	})
	t.Run("not existing block type, different shard", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return false
				},
			},
		)

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			0,
		)

		assert.Equal(t, 2, len(scrsInfo))
	})
	t.Run("already existing txs, different shard", func(t *testing.T) {
		t.Parallel()

		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&genericMocks.StorerMock{},
			&marshal.GogoProtoMarshalizer{},
			&mock.ShardCoordinatorStub{
				SameShardCalled: func(_, _ []byte) bool {
					return false
				},
			},
		)

		allTxsAfterExec := map[string]data.TransactionHandler{
			"txHash1": &transaction.Transaction{Nonce: 1},
			"txHash2": &transaction.Transaction{Nonce: 2},
		}

		scrsInfo := scheduledTxsExec.getAllIntermediateTxsAfterScheduledExecution(
			allTxsBeforeExec[0],
			allTxsAfterExec,
			0,
		)

		assert.Equal(t, 0, len(scrsInfo))
	})
}

func TestScheduledTxsExecution_GetScheduledIntermediateTxsNonEmptySCRsMap(t *testing.T) {
	t.Parallel()

	allTxsAfterExec := map[block.Type]map[string]data.TransactionHandler{
		0: {
			"txHash1": &transaction.Transaction{Nonce: 1},
			"txHash2": &transaction.Transaction{Nonce: 2},
		},
		1: {
			"txHash3": &transaction.Transaction{Nonce: 3},
			"txHash4": &transaction.Transaction{Nonce: 4},
		},
	}

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{
			SameShardCalled: func(_, _ []byte) bool {
				return false
			},
		},
	)

	scheduledTxsExec.ComputeScheduledIntermediateTxs(
		nil,
		allTxsAfterExec,
	)

	scheduledIntermediateTxs := scheduledTxsExec.GetScheduledIntermediateTxs()

	assert.Equal(t, 2, len(scheduledIntermediateTxs))
	assert.Equal(t, 2, len(scheduledIntermediateTxs[0]))
	assert.Equal(t, 2, len(scheduledIntermediateTxs[1]))
}

func TestScheduledTxsExecution_GetScheduledIntermediateTxsEmptySCRsMap(t *testing.T) {
	t.Parallel()

	allTxsAfterExec := make(map[block.Type]map[string]data.TransactionHandler)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{
			SameShardCalled: func(_, _ []byte) bool {
				return false
			},
		},
	)

	scheduledTxsExec.ComputeScheduledIntermediateTxs(
		nil,
		allTxsAfterExec,
	)

	scheduledIntermediateTxs := scheduledTxsExec.GetScheduledIntermediateTxs()

	assert.Equal(t, 0, len(scheduledIntermediateTxs))
}

func TestScheduledTxsExecution_SetScheduledInfo(t *testing.T) {
	t.Parallel()

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	rootHash := []byte("root hash")
	gasAndFees := scheduled.GasAndFees{}
	intermediateTxs := map[block.Type][]data.TransactionHandler{
		0: {
			&transaction.Transaction{Nonce: 1},
			&transaction.Transaction{Nonce: 2},
		},
		1: {
			&transaction.Transaction{Nonce: 3},
			&transaction.Transaction{Nonce: 4},
		},
	}
	mbs := block.MiniBlockSlice{
		0: {
			Type: block.InvalidBlock,
		},
	}

	scheduledInfo := &process.ScheduledInfo{
		RootHash:        rootHash,
		IntermediateTxs: intermediateTxs,
		GasAndFees:      gasAndFees,
		MiniBlocks:      mbs,
	}
	scheduledTxsExec.SetScheduledInfo(scheduledInfo)

	assert.Equal(t, rootHash, scheduledTxsExec.GetScheduledRootHash())
	assert.Equal(t, gasAndFees, scheduledTxsExec.GetScheduledGasAndFees())
	assert.Equal(t, intermediateTxs, scheduledTxsExec.GetScheduledIntermediateTxs())
	assert.Equal(t, mbs, scheduledTxsExec.GetScheduledMBs())
}

func TestScheduledTxsExecution_Setters(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	gasAndFees := scheduled.GasAndFees{}

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)
	scheduledTxsExec.SetTransactionCoordinator(&mock.TransactionCoordinatorMock{})
	scheduledTxsExec.SetTransactionProcessor(&testscommon.TxProcessorMock{})

	scheduledTxsExec.SetScheduledGasAndFees(gasAndFees)
	assert.Equal(t, gasAndFees, scheduledTxsExec.GetScheduledGasAndFees())

	scheduledTxsExec.SetScheduledRootHash(rootHash)
	assert.Equal(t, rootHash, scheduledTxsExec.GetScheduledRootHash())
}

func TestScheduledTxsExecution_getScheduledInfoForHeaderShouldFail(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")

	t.Run("failed to get SCRs saved data from storage", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("storer err")
		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&storageMocks.StorerStub{
				GetCalled: func(_ []byte) ([]byte, error) {
					return nil, expectedErr
				},
			},
			&marshal.GogoProtoMarshalizer{},
			&mock.ShardCoordinatorStub{},
		)

		scheduledInfo, err := scheduledTxsExec.getScheduledInfoForHeader(rootHash)
		assert.Nil(t, scheduledInfo)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("failed to unmarshal data", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("marshaller err")
		scheduledTxsExec, _ := NewScheduledTxsExecution(
			&testscommon.TxProcessorMock{},
			&mock.TransactionCoordinatorMock{},
			&storageMocks.StorerStub{
				GetCalled: func(_ []byte) ([]byte, error) {
					return nil, nil
				},
			},
			&testscommon.MarshalizerStub{
				UnmarshalCalled: func(_ interface{}, _ []byte) error {
					return expectedErr
				},
			},
			&mock.ShardCoordinatorStub{},
		)

		scheduledInfo, err := scheduledTxsExec.getScheduledInfoForHeader(rootHash)
		assert.Nil(t, scheduledInfo)
		assert.Equal(t, expectedErr, err)
	})
}

func TestScheduledTxsExecution_getScheduledInfoForHeaderShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("root hash")
	expectedGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(102),
		GasProvided:     103,
		GasPenalized:    104,
		GasRefunded:     105,
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash: headerHash,
		Scrs: map[int32]scheduled.SmartContractResults{
			0: {
				TxHandlers: []*smartContractResult.SmartContractResult{},
			},
		},
		GasAndFees: &expectedGasAndFees,
	}
	marshalledSCRsSavedData, _ := json.Marshal(scheduledSCRs)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return marshalledSCRsSavedData, nil
			},
		},
		&testscommon.MarshalizerMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledInfo, _ := scheduledTxsExec.getScheduledInfoForHeader(headerHash)

	assert.Equal(t, headerHash, scheduledInfo.RootHash)
	assert.Equal(t, expectedGasAndFees, scheduledInfo.GasAndFees)
	assert.NotNil(t, scheduledInfo.IntermediateTxs)
	assert.Equal(t, make(block.MiniBlockSlice, 0), scheduledInfo.MiniBlocks)
}

func TestScheduledTxsExecution_getMarshalledScheduledInfoShouldFail(t *testing.T) {
	t.Parallel()

	scheduledRootHash := []byte("root hash")
	mapSCRs := map[block.Type][]data.TransactionHandler{
		0: {
			&transaction.Transaction{Nonce: 1},
		},
	}
	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(102),
		GasProvided:     103,
		GasPenalized:    104,
		GasRefunded:     105,
	}

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&testscommon.MarshalizerMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledInfo := &process.ScheduledInfo{
		RootHash:        scheduledRootHash,
		IntermediateTxs: mapSCRs,
		GasAndFees:      gasAndFees,
	}
	marshalledSCRs, err := scheduledTxsExec.getMarshalledScheduledInfo(scheduledInfo)
	assert.Nil(t, marshalledSCRs)
	assert.Error(t, err)
}

func TestScheduledTxsExecution_getMarshalledScheduledInfoShouldWork(t *testing.T) {
	t.Parallel()

	scheduledRootHash := []byte("root hash")
	mapSCRs := map[block.Type][]data.TransactionHandler{
		0: {
			&smartContractResult.SmartContractResult{
				Nonce: 1,
			},
		},
	}
	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(100),
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash: scheduledRootHash,
		Scrs: map[int32]scheduled.SmartContractResults{
			0: {
				TxHandlers: []*smartContractResult.SmartContractResult{
					{
						Nonce: 1,
					},
				},
			},
		},
		GasAndFees: &gasAndFees,
	}
	expectedScheduledSCRs, _ := json.Marshal(scheduledSCRs)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&testscommon.MarshalizerMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledInfo := &process.ScheduledInfo{
		RootHash:        scheduledRootHash,
		IntermediateTxs: mapSCRs,
		GasAndFees:      gasAndFees,
	}
	marshalledSCRs, err := scheduledTxsExec.getMarshalledScheduledInfo(scheduledInfo)
	assert.Nil(t, err)
	assert.Equal(t, expectedScheduledSCRs, marshalledSCRs)
}

func TestScheduledTxsExecution_RollBackToBlockShouldFail(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")

	expectedErr := errors.New("local err")
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return nil, expectedErr
			},
		},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.RollBackToBlock(rootHash)
	assert.Equal(t, expectedErr, err)
}

func TestScheduledTxsExecution_RollBackToBlockShouldWork(t *testing.T) {
	t.Parallel()

	headerHash := []byte("root hash")
	expectedGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(102),
		GasProvided:     103,
		GasPenalized:    104,
		GasRefunded:     105,
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash: headerHash,
		Scrs: map[int32]scheduled.SmartContractResults{
			0: {
				TxHandlers: []*smartContractResult.SmartContractResult{},
			},
		},
		GasAndFees: &expectedGasAndFees,
	}
	marshalledSCRsSavedData, _ := json.Marshal(scheduledSCRs)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			GetCalled: func(_ []byte) ([]byte, error) {
				return marshalledSCRsSavedData, nil
			},
		},
		&testscommon.MarshalizerMock{},
		&mock.ShardCoordinatorStub{},
	)

	err := scheduledTxsExec.RollBackToBlock(headerHash)
	assert.Nil(t, err)

	scheduledInfo, err := scheduledTxsExec.getScheduledInfoForHeader(headerHash)
	require.Nil(t, err)
	assert.Equal(t, headerHash, scheduledInfo.RootHash)
	assert.Equal(t, expectedGasAndFees, scheduledInfo.GasAndFees)
	assert.NotNil(t, scheduledInfo.IntermediateTxs)
	assert.Equal(t, make(block.MiniBlockSlice, 0), scheduledInfo.MiniBlocks)
}

func TestScheduledTxsExecution_SaveState(t *testing.T) {
	t.Parallel()

	headerHash := []byte("header hash")
	scheduledRootHash := []byte("scheduled root hash")
	mapSCRs := map[block.Type][]data.TransactionHandler{
		0: {
			&smartContractResult.SmartContractResult{
				Nonce: 1,
			},
		},
	}
	gasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(100),
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash: scheduledRootHash,
		Scrs: map[int32]scheduled.SmartContractResults{
			0: {
				TxHandlers: []*smartContractResult.SmartContractResult{
					{
						Nonce: 1,
					},
				},
			},
		},
		GasAndFees: &gasAndFees,
	}
	marshalledScheduledData, _ := json.Marshal(scheduledSCRs)

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			PutCalled: func(key, data []byte) error {
				require.Equal(t, headerHash, key)
				require.Equal(t, marshalledScheduledData, data)
				return nil
			},
		},
		&testscommon.MarshalizerMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledInfo := &process.ScheduledInfo{
		RootHash:        scheduledRootHash,
		IntermediateTxs: mapSCRs,
		GasAndFees:      gasAndFees,
	}
	scheduledTxsExec.SaveState(headerHash, scheduledInfo)
}

func TestScheduledTxsExecution_SaveStateIfNeeded(t *testing.T) {
	t.Parallel()

	headerHash := []byte("header hash")

	wasCalled := false
	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&storageMocks.StorerStub{
			PutCalled: func(key, _ []byte) error {
				wasCalled = true
				require.Equal(t, headerHash, key)
				return nil
			},
		},
		&testscommon.MarshalizerMock{},
		&mock.ShardCoordinatorStub{},
	)

	scheduledTxsExec.SaveStateIfNeeded(headerHash)
	assert.False(t, wasCalled)

	scheduledTxsExec.AddScheduledTx([]byte("txHash1"), &transaction.Transaction{Nonce: 0})

	scheduledTxsExec.SaveStateIfNeeded(headerHash)
	assert.True(t, wasCalled)
}

func TestScheduledTxsExecution_IsScheduledTx(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")

	scheduledTxsExec, _ := NewScheduledTxsExecution(
		&testscommon.TxProcessorMock{},
		&mock.TransactionCoordinatorMock{},
		&genericMocks.StorerMock{},
		&marshal.GogoProtoMarshalizer{},
		&mock.ShardCoordinatorStub{},
	)
	scheduledTxsExec.AddScheduledTx(txHash1, &transaction.Transaction{Nonce: 0})

	ok := scheduledTxsExec.IsScheduledTx(txHash1)
	assert.True(t, ok)

	ok = scheduledTxsExec.IsScheduledTx(txHash2)
	assert.False(t, ok)
}
