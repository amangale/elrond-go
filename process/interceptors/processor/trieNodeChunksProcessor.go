package processor

import (
	"context"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor/chunk"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const minimumRequestTimeInterval = time.Second

type chunkHandler interface {
	Put(chunkIndex uint32, buff []byte)
	TryAssembleAllChunks() []byte
	GetAllMissingChunkIndexes() []uint32
	Size() int
	IsInterfaceNil() bool
}

type checkRequest struct {
	batch        *batch.Batch
	chanResponse chan process.CheckedChunkResult
}

// TrieNodesChunksProcessorArgs is the argument DTO used in the trieNodeChunksProcessor constructor
type TrieNodesChunksProcessorArgs struct {
	Hasher          hashing.Hasher
	ChunksCacher    storage.Cacher
	RequestInterval time.Duration
	RequestHandler  process.RequestHandler
	Topic           string
	ShardID         uint32
}

type trieNodeChunksProcessor struct {
	hasher            hashing.Hasher
	chunksCacher      storage.Cacher
	chanCheckRequests chan checkRequest
	requestInterval   time.Duration
	requestHandler    process.RequestHandler
	topic             string
	shardID           uint32
	cancel            func()
}

// NewTrieNodeChunksProcessor creates a new trieNodeChunksProcessor instance
func NewTrieNodeChunksProcessor(arg TrieNodesChunksProcessorArgs) (*trieNodeChunksProcessor, error) {
	if check.IfNil(arg.Hasher) {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor", process.ErrNilHasher)
	}
	if check.IfNil(arg.ChunksCacher) {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor", process.ErrNilCacher)
	}
	if arg.RequestInterval < minimumRequestTimeInterval {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor, minimum request interval is %v",
			process.ErrInvalidValue, minimumRequestTimeInterval)
	}
	if check.IfNil(arg.RequestHandler) {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor", process.ErrNilRequestHandler)
	}
	if len(arg.Topic) == 0 {
		return nil, fmt.Errorf("%w in NewTrieNodeChunksProcessor", process.ErrEmptyTopic)
	}

	tncp := &trieNodeChunksProcessor{
		hasher:            arg.Hasher,
		chunksCacher:      arg.ChunksCacher,
		chanCheckRequests: make(chan checkRequest),
		requestInterval:   arg.RequestInterval,
		requestHandler:    arg.RequestHandler,
		topic:             arg.Topic,
		shardID:           arg.ShardID,
	}
	var ctx context.Context
	ctx, tncp.cancel = context.WithCancel(context.Background())
	go tncp.processLoop(ctx)

	return tncp, nil
}

func (proc *trieNodeChunksProcessor) processLoop(ctx context.Context) {
	chanDoRequests := time.After(proc.requestInterval)
	for {
		select {
		case <-ctx.Done():
			log.Debug("trieNodeChunksProcessor.processLoop go routine is stopping...")
			return
		case request := <-proc.chanCheckRequests:
			proc.processCheckRequest(request)
		case <-chanDoRequests:
			proc.doRequests(ctx)
			chanDoRequests = time.After(proc.requestInterval)
		}
	}
}

// CheckBatch will check the batch returning a checked chunk result containing result processing
func (proc *trieNodeChunksProcessor) CheckBatch(b *batch.Batch) (process.CheckedChunkResult, error) {
	batchValid, err := proc.batchIsValid(b)
	if !batchValid {
		return process.CheckedChunkResult{
			IsChunk:        false,
			HaveAllChunks:  false,
			CompleteBuffer: nil,
		}, err
	}

	respChan := make(chan process.CheckedChunkResult, 1)
	req := checkRequest{
		batch:        b,
		chanResponse: respChan,
	}

	proc.chanCheckRequests <- req
	response := <-respChan

	return response, nil
}

func (proc *trieNodeChunksProcessor) processCheckRequest(cr checkRequest) {
	chunkObject, found := proc.chunksCacher.Get(cr.batch.Reference)
	if !found {
		chunkObject = chunk.NewChunk(cr.batch.MaxChunks)
	}
	chunkData, ok := chunkObject.(chunkHandler)
	if !ok {
		chunkData = chunk.NewChunk(cr.batch.MaxChunks)
	}

	chunkData.Put(cr.batch.ChunkIndex, cr.batch.Data[0])

	buff := chunkData.TryAssembleAllChunks()
	haveAllChunks := len(buff) > 0
	if haveAllChunks {
		proc.chunksCacher.Remove(cr.batch.Reference)
	} else {
		proc.chunksCacher.Put(cr.batch.Reference, chunkData, chunkData.Size())
	}

	cr.chanResponse <- process.CheckedChunkResult{
		IsChunk:        true,
		HaveAllChunks:  haveAllChunks,
		CompleteBuffer: buff,
	}
}

func (proc *trieNodeChunksProcessor) batchIsValid(b *batch.Batch) (bool, error) {
	if b.MaxChunks < 2 {
		return false, nil
	}
	if len(b.Reference) != proc.hasher.Size() {
		return false, process.ErrIncompatibleReference
	}
	if len(b.Data) != 1 {
		return false, nil
	}

	return true, nil
}

func (proc *trieNodeChunksProcessor) doRequests(ctx context.Context) {
	references := proc.chunksCacher.Keys()
	for _, ref := range references {
		select {
		case <-ctx.Done():
			//early exit
			return
		default:
		}

		proc.requestMissingForReference(ref, ctx)
	}
}

func (proc *trieNodeChunksProcessor) requestMissingForReference(reference []byte, ctx context.Context) {
	data, found := proc.chunksCacher.Get(reference)
	if !found {
		return
	}

	chunkData, ok := data.(chunkHandler)
	if !ok {
		return
	}

	missing := chunkData.GetAllMissingChunkIndexes()
	for _, missingChunkIndex := range missing {
		select {
		case <-ctx.Done():
			//early exit
			return
		default:
		}

		proc.requestHandler.RequestTrieNode(proc.shardID, reference, proc.topic, missingChunkIndex)
	}
}

// Close will close the process go routine
func (proc *trieNodeChunksProcessor) Close() error {
	proc.cancel()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (proc *trieNodeChunksProcessor) IsInterfaceNil() bool {
	return proc == nil
}
