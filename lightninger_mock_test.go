// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package raiju

import (
	"context"
	"github.com/nyonson/raiju/lightning"
	"sync"
	"time"
)

// lightningerMock is a mock implementation of lightninger.
//
//	func TestSomethingThatUseslightninger(t *testing.T) {
//
//		// make and configure a mocked lightninger
//		mockedlightninger := &lightningerMock{
//			AddInvoiceFunc: func(ctx context.Context, amount lightning.Satoshi) (lightning.Invoice, error) {
//				panic("mock out the AddInvoice method")
//			},
//			DescribeGraphFunc: func(ctx context.Context) (*lightning.Graph, error) {
//				panic("mock out the DescribeGraph method")
//			},
//			ForwardingHistoryFunc: func(ctx context.Context, since time.Time) ([]lightning.Forward, error) {
//				panic("mock out the ForwardingHistory method")
//			},
//			GetChannelFunc: func(ctx context.Context, channelID lightning.ChannelID) (lightning.Channel, error) {
//				panic("mock out the GetChannel method")
//			},
//			GetInfoFunc: func(ctx context.Context) (*lightning.Info, error) {
//				panic("mock out the GetInfo method")
//			},
//			ListChannelsFunc: func(ctx context.Context) (lightning.Channels, error) {
//				panic("mock out the ListChannels method")
//			},
//			SendPaymentFunc: func(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubKey lightning.PubKey, maxFee lightning.Satoshi) (lightning.Satoshi, error) {
//				panic("mock out the SendPayment method")
//			},
//			SetFeesFunc: func(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error {
//				panic("mock out the SetFees method")
//			},
//			SubscribeChannelUpdatesFunc: func(ctx context.Context) (<-chan lightning.Channels, <-chan error, error) {
//				panic("mock out the SubscribeChannelUpdates method")
//			},
//		}
//
//		// use mockedlightninger in code that requires lightninger
//		// and then make assertions.
//
//	}
type lightningerMock struct {
	// AddInvoiceFunc mocks the AddInvoice method.
	AddInvoiceFunc func(ctx context.Context, amount lightning.Satoshi) (lightning.Invoice, error)

	// DescribeGraphFunc mocks the DescribeGraph method.
	DescribeGraphFunc func(ctx context.Context) (*lightning.Graph, error)

	// ForwardingHistoryFunc mocks the ForwardingHistory method.
	ForwardingHistoryFunc func(ctx context.Context, since time.Time) ([]lightning.Forward, error)

	// GetChannelFunc mocks the GetChannel method.
	GetChannelFunc func(ctx context.Context, channelID lightning.ChannelID) (lightning.Channel, error)

	// GetInfoFunc mocks the GetInfo method.
	GetInfoFunc func(ctx context.Context) (*lightning.Info, error)

	// ListChannelsFunc mocks the ListChannels method.
	ListChannelsFunc func(ctx context.Context) (lightning.Channels, error)

	// SendPaymentFunc mocks the SendPayment method.
	SendPaymentFunc func(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubKey lightning.PubKey, maxFee lightning.Satoshi) (lightning.Satoshi, error)

	// SetFeesFunc mocks the SetFees method.
	SetFeesFunc func(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error

	// SubscribeChannelUpdatesFunc mocks the SubscribeChannelUpdates method.
	SubscribeChannelUpdatesFunc func(ctx context.Context) (<-chan lightning.Channels, <-chan error, error)

	// calls tracks calls to the methods.
	calls struct {
		// AddInvoice holds details about calls to the AddInvoice method.
		AddInvoice []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Amount is the amount argument value.
			Amount lightning.Satoshi
		}
		// DescribeGraph holds details about calls to the DescribeGraph method.
		DescribeGraph []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// ForwardingHistory holds details about calls to the ForwardingHistory method.
		ForwardingHistory []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Since is the since argument value.
			Since time.Time
		}
		// GetChannel holds details about calls to the GetChannel method.
		GetChannel []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ChannelID is the channelID argument value.
			ChannelID lightning.ChannelID
		}
		// GetInfo holds details about calls to the GetInfo method.
		GetInfo []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// ListChannels holds details about calls to the ListChannels method.
		ListChannels []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// SendPayment holds details about calls to the SendPayment method.
		SendPayment []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Invoice is the invoice argument value.
			Invoice lightning.Invoice
			// OutChannelID is the outChannelID argument value.
			OutChannelID lightning.ChannelID
			// LastHopPubKey is the lastHopPubKey argument value.
			LastHopPubKey lightning.PubKey
			// MaxFee is the maxFee argument value.
			MaxFee lightning.Satoshi
		}
		// SetFees holds details about calls to the SetFees method.
		SetFees []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ChannelID is the channelID argument value.
			ChannelID lightning.ChannelID
			// Fee is the fee argument value.
			Fee lightning.FeePPM
		}
		// SubscribeChannelUpdates holds details about calls to the SubscribeChannelUpdates method.
		SubscribeChannelUpdates []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
	lockAddInvoice              sync.RWMutex
	lockDescribeGraph           sync.RWMutex
	lockForwardingHistory       sync.RWMutex
	lockGetChannel              sync.RWMutex
	lockGetInfo                 sync.RWMutex
	lockListChannels            sync.RWMutex
	lockSendPayment             sync.RWMutex
	lockSetFees                 sync.RWMutex
	lockSubscribeChannelUpdates sync.RWMutex
}

// AddInvoice calls AddInvoiceFunc.
func (mock *lightningerMock) AddInvoice(ctx context.Context, amount lightning.Satoshi) (lightning.Invoice, error) {
	callInfo := struct {
		Ctx    context.Context
		Amount lightning.Satoshi
	}{
		Ctx:    ctx,
		Amount: amount,
	}
	mock.lockAddInvoice.Lock()
	mock.calls.AddInvoice = append(mock.calls.AddInvoice, callInfo)
	mock.lockAddInvoice.Unlock()
	if mock.AddInvoiceFunc == nil {
		var (
			invoiceOut lightning.Invoice
			errOut     error
		)
		return invoiceOut, errOut
	}
	return mock.AddInvoiceFunc(ctx, amount)
}

// AddInvoiceCalls gets all the calls that were made to AddInvoice.
// Check the length with:
//
//	len(mockedlightninger.AddInvoiceCalls())
func (mock *lightningerMock) AddInvoiceCalls() []struct {
	Ctx    context.Context
	Amount lightning.Satoshi
} {
	var calls []struct {
		Ctx    context.Context
		Amount lightning.Satoshi
	}
	mock.lockAddInvoice.RLock()
	calls = mock.calls.AddInvoice
	mock.lockAddInvoice.RUnlock()
	return calls
}

// DescribeGraph calls DescribeGraphFunc.
func (mock *lightningerMock) DescribeGraph(ctx context.Context) (*lightning.Graph, error) {
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockDescribeGraph.Lock()
	mock.calls.DescribeGraph = append(mock.calls.DescribeGraph, callInfo)
	mock.lockDescribeGraph.Unlock()
	if mock.DescribeGraphFunc == nil {
		var (
			graphOut *lightning.Graph
			errOut   error
		)
		return graphOut, errOut
	}
	return mock.DescribeGraphFunc(ctx)
}

// DescribeGraphCalls gets all the calls that were made to DescribeGraph.
// Check the length with:
//
//	len(mockedlightninger.DescribeGraphCalls())
func (mock *lightningerMock) DescribeGraphCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockDescribeGraph.RLock()
	calls = mock.calls.DescribeGraph
	mock.lockDescribeGraph.RUnlock()
	return calls
}

// ForwardingHistory calls ForwardingHistoryFunc.
func (mock *lightningerMock) ForwardingHistory(ctx context.Context, since time.Time) ([]lightning.Forward, error) {
	callInfo := struct {
		Ctx   context.Context
		Since time.Time
	}{
		Ctx:   ctx,
		Since: since,
	}
	mock.lockForwardingHistory.Lock()
	mock.calls.ForwardingHistory = append(mock.calls.ForwardingHistory, callInfo)
	mock.lockForwardingHistory.Unlock()
	if mock.ForwardingHistoryFunc == nil {
		var (
			forwardsOut []lightning.Forward
			errOut      error
		)
		return forwardsOut, errOut
	}
	return mock.ForwardingHistoryFunc(ctx, since)
}

// ForwardingHistoryCalls gets all the calls that were made to ForwardingHistory.
// Check the length with:
//
//	len(mockedlightninger.ForwardingHistoryCalls())
func (mock *lightningerMock) ForwardingHistoryCalls() []struct {
	Ctx   context.Context
	Since time.Time
} {
	var calls []struct {
		Ctx   context.Context
		Since time.Time
	}
	mock.lockForwardingHistory.RLock()
	calls = mock.calls.ForwardingHistory
	mock.lockForwardingHistory.RUnlock()
	return calls
}

// GetChannel calls GetChannelFunc.
func (mock *lightningerMock) GetChannel(ctx context.Context, channelID lightning.ChannelID) (lightning.Channel, error) {
	callInfo := struct {
		Ctx       context.Context
		ChannelID lightning.ChannelID
	}{
		Ctx:       ctx,
		ChannelID: channelID,
	}
	mock.lockGetChannel.Lock()
	mock.calls.GetChannel = append(mock.calls.GetChannel, callInfo)
	mock.lockGetChannel.Unlock()
	if mock.GetChannelFunc == nil {
		var (
			channelOut lightning.Channel
			errOut     error
		)
		return channelOut, errOut
	}
	return mock.GetChannelFunc(ctx, channelID)
}

// GetChannelCalls gets all the calls that were made to GetChannel.
// Check the length with:
//
//	len(mockedlightninger.GetChannelCalls())
func (mock *lightningerMock) GetChannelCalls() []struct {
	Ctx       context.Context
	ChannelID lightning.ChannelID
} {
	var calls []struct {
		Ctx       context.Context
		ChannelID lightning.ChannelID
	}
	mock.lockGetChannel.RLock()
	calls = mock.calls.GetChannel
	mock.lockGetChannel.RUnlock()
	return calls
}

// GetInfo calls GetInfoFunc.
func (mock *lightningerMock) GetInfo(ctx context.Context) (*lightning.Info, error) {
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockGetInfo.Lock()
	mock.calls.GetInfo = append(mock.calls.GetInfo, callInfo)
	mock.lockGetInfo.Unlock()
	if mock.GetInfoFunc == nil {
		var (
			infoOut *lightning.Info
			errOut  error
		)
		return infoOut, errOut
	}
	return mock.GetInfoFunc(ctx)
}

// GetInfoCalls gets all the calls that were made to GetInfo.
// Check the length with:
//
//	len(mockedlightninger.GetInfoCalls())
func (mock *lightningerMock) GetInfoCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockGetInfo.RLock()
	calls = mock.calls.GetInfo
	mock.lockGetInfo.RUnlock()
	return calls
}

// ListChannels calls ListChannelsFunc.
func (mock *lightningerMock) ListChannels(ctx context.Context) (lightning.Channels, error) {
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockListChannels.Lock()
	mock.calls.ListChannels = append(mock.calls.ListChannels, callInfo)
	mock.lockListChannels.Unlock()
	if mock.ListChannelsFunc == nil {
		var (
			channelsOut lightning.Channels
			errOut      error
		)
		return channelsOut, errOut
	}
	return mock.ListChannelsFunc(ctx)
}

// ListChannelsCalls gets all the calls that were made to ListChannels.
// Check the length with:
//
//	len(mockedlightninger.ListChannelsCalls())
func (mock *lightningerMock) ListChannelsCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockListChannels.RLock()
	calls = mock.calls.ListChannels
	mock.lockListChannels.RUnlock()
	return calls
}

// SendPayment calls SendPaymentFunc.
func (mock *lightningerMock) SendPayment(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubKey lightning.PubKey, maxFee lightning.Satoshi) (lightning.Satoshi, error) {
	callInfo := struct {
		Ctx           context.Context
		Invoice       lightning.Invoice
		OutChannelID  lightning.ChannelID
		LastHopPubKey lightning.PubKey
		MaxFee        lightning.Satoshi
	}{
		Ctx:           ctx,
		Invoice:       invoice,
		OutChannelID:  outChannelID,
		LastHopPubKey: lastHopPubKey,
		MaxFee:        maxFee,
	}
	mock.lockSendPayment.Lock()
	mock.calls.SendPayment = append(mock.calls.SendPayment, callInfo)
	mock.lockSendPayment.Unlock()
	if mock.SendPaymentFunc == nil {
		var (
			satoshiOut lightning.Satoshi
			errOut     error
		)
		return satoshiOut, errOut
	}
	return mock.SendPaymentFunc(ctx, invoice, outChannelID, lastHopPubKey, maxFee)
}

// SendPaymentCalls gets all the calls that were made to SendPayment.
// Check the length with:
//
//	len(mockedlightninger.SendPaymentCalls())
func (mock *lightningerMock) SendPaymentCalls() []struct {
	Ctx           context.Context
	Invoice       lightning.Invoice
	OutChannelID  lightning.ChannelID
	LastHopPubKey lightning.PubKey
	MaxFee        lightning.Satoshi
} {
	var calls []struct {
		Ctx           context.Context
		Invoice       lightning.Invoice
		OutChannelID  lightning.ChannelID
		LastHopPubKey lightning.PubKey
		MaxFee        lightning.Satoshi
	}
	mock.lockSendPayment.RLock()
	calls = mock.calls.SendPayment
	mock.lockSendPayment.RUnlock()
	return calls
}

// SetFees calls SetFeesFunc.
func (mock *lightningerMock) SetFees(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error {
	callInfo := struct {
		Ctx       context.Context
		ChannelID lightning.ChannelID
		Fee       lightning.FeePPM
	}{
		Ctx:       ctx,
		ChannelID: channelID,
		Fee:       fee,
	}
	mock.lockSetFees.Lock()
	mock.calls.SetFees = append(mock.calls.SetFees, callInfo)
	mock.lockSetFees.Unlock()
	if mock.SetFeesFunc == nil {
		var (
			errOut error
		)
		return errOut
	}
	return mock.SetFeesFunc(ctx, channelID, fee)
}

// SetFeesCalls gets all the calls that were made to SetFees.
// Check the length with:
//
//	len(mockedlightninger.SetFeesCalls())
func (mock *lightningerMock) SetFeesCalls() []struct {
	Ctx       context.Context
	ChannelID lightning.ChannelID
	Fee       lightning.FeePPM
} {
	var calls []struct {
		Ctx       context.Context
		ChannelID lightning.ChannelID
		Fee       lightning.FeePPM
	}
	mock.lockSetFees.RLock()
	calls = mock.calls.SetFees
	mock.lockSetFees.RUnlock()
	return calls
}

// SubscribeChannelUpdates calls SubscribeChannelUpdatesFunc.
func (mock *lightningerMock) SubscribeChannelUpdates(ctx context.Context) (<-chan lightning.Channels, <-chan error, error) {
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockSubscribeChannelUpdates.Lock()
	mock.calls.SubscribeChannelUpdates = append(mock.calls.SubscribeChannelUpdates, callInfo)
	mock.lockSubscribeChannelUpdates.Unlock()
	if mock.SubscribeChannelUpdatesFunc == nil {
		var (
			channelsChOut <-chan lightning.Channels
			errChOut      <-chan error
			errOut        error
		)
		return channelsChOut, errChOut, errOut
	}
	return mock.SubscribeChannelUpdatesFunc(ctx)
}

// SubscribeChannelUpdatesCalls gets all the calls that were made to SubscribeChannelUpdates.
// Check the length with:
//
//	len(mockedlightninger.SubscribeChannelUpdatesCalls())
func (mock *lightningerMock) SubscribeChannelUpdatesCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockSubscribeChannelUpdates.RLock()
	calls = mock.calls.SubscribeChannelUpdates
	mock.lockSubscribeChannelUpdates.RUnlock()
	return calls
}
