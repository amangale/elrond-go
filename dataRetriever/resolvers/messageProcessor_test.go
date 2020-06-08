package resolvers

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//------- canProcessMessage

func TestMessageProcessor_CanProcessErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	mp := &messageProcessor{
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return expectedErr
			},
		},
	}

	err := mp.canProcessMessage(&mock.P2PMessageMock{}, "")

	assert.True(t, errors.Is(err, expectedErr))
}

func TestMessageProcessor_CanProcessOnTopicErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	mp := &messageProcessor{
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return nil
			},
			CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64) error {
				return expectedErr
			},
		},
	}

	err := mp.canProcessMessage(&mock.P2PMessageMock{}, "")

	assert.True(t, errors.Is(err, expectedErr))
}

func TestMessageProcessor_CanProcessThrottlerNotAllowingShouldErr(t *testing.T) {
	t.Parallel()

	canProcessWasCalled := false
	mp := &messageProcessor{
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return nil
			},
			CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64) error {
				return nil
			},
		},
		throttler: &mock.ThrottlerStub{
			CanProcessCalled: func() bool {
				canProcessWasCalled = true
				return false
			},
		},
	}

	err := mp.canProcessMessage(&mock.P2PMessageMock{}, "")

	assert.True(t, errors.Is(err, dataRetriever.ErrSystemBusy))
	assert.True(t, canProcessWasCalled)
}

func TestMessageProcessor_CanProcessShouldWork(t *testing.T) {
	t.Parallel()

	canProcessWasCalled := false
	mp := &messageProcessor{
		antifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return nil
			},
			CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64) error {
				return nil
			},
		},
		throttler: &mock.ThrottlerStub{
			CanProcessCalled: func() bool {
				canProcessWasCalled = true
				return true
			},
		},
	}

	err := mp.canProcessMessage(&mock.P2PMessageMock{}, "")

	assert.Nil(t, err)
	assert.True(t, canProcessWasCalled)
}

//------- parseReceivedMessage

func TestMessageProcessor_ParseReceivedMessageMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	mp := &messageProcessor{
		marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return expectedErr
			},
		},
	}

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}
	rd, err := mp.parseReceivedMessage(msg)

	assert.Equal(t, err, expectedErr)
	assert.Nil(t, rd)
}

func TestMessageProcessor_ParseReceivedMessageNilValueFieldShouldErr(t *testing.T) {
	t.Parallel()

	mp := &messageProcessor{
		marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return nil
			},
		},
	}

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}
	rd, err := mp.parseReceivedMessage(msg)

	assert.Equal(t, err, dataRetriever.ErrNilValue)
	assert.Nil(t, rd)
}

func TestMessageProcessor_ParseReceivedMessageShouldWork(t *testing.T) {
	t.Parallel()

	expectedValue := []byte("expected value")
	mp := &messageProcessor{
		marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				rd := obj.(*dataRetriever.RequestData)
				rd.Value = expectedValue

				return nil
			},
		},
	}

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}
	rd, err := mp.parseReceivedMessage(msg)

	assert.Nil(t, err)
	require.NotNil(t, rd)
	assert.Equal(t, expectedValue, rd.Value)
}