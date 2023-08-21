// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package lightning

import (
	"context"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/routing/route"
	"sync"
)

// channelerMock is a mock implementation of channeler.
//
//	func TestSomethingThatUseschanneler(t *testing.T) {
//
//		// make and configure a mocked channeler
//		mockedchanneler := &channelerMock{
//			DescribeGraphFunc: func(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error) {
//				panic("mock out the DescribeGraph method")
//			},
//			ForwardingHistoryFunc: func(ctx context.Context, req lndclient.ForwardingHistoryRequest) (*lndclient.ForwardingHistoryResponse, error) {
//				panic("mock out the ForwardingHistory method")
//			},
//			GetChanInfoFunc: func(ctx context.Context, chanId uint64) (*lndclient.ChannelEdge, error) {
//				panic("mock out the GetChanInfo method")
//			},
//			GetInfoFunc: func(ctx context.Context) (*lndclient.Info, error) {
//				panic("mock out the GetInfo method")
//			},
//			GetNodeInfoFunc: func(ctx context.Context, pubkey route.Vertex, includeChannels bool) (*lndclient.NodeInfo, error) {
//				panic("mock out the GetNodeInfo method")
//			},
//			ListChannelsFunc: func(ctx context.Context, activeOnly bool, publicOnly bool) ([]lndclient.ChannelInfo, error) {
//				panic("mock out the ListChannels method")
//			},
//			UpdateChanPolicyFunc: func(ctx context.Context, req lndclient.PolicyUpdateRequest, chanPoint *wire.OutPoint) error {
//				panic("mock out the UpdateChanPolicy method")
//			},
//		}
//
//		// use mockedchanneler in code that requires channeler
//		// and then make assertions.
//
//	}
type channelerMock struct {
	// DescribeGraphFunc mocks the DescribeGraph method.
	DescribeGraphFunc func(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error)

	// ForwardingHistoryFunc mocks the ForwardingHistory method.
	ForwardingHistoryFunc func(ctx context.Context, req lndclient.ForwardingHistoryRequest) (*lndclient.ForwardingHistoryResponse, error)

	// GetChanInfoFunc mocks the GetChanInfo method.
	GetChanInfoFunc func(ctx context.Context, chanId uint64) (*lndclient.ChannelEdge, error)

	// GetInfoFunc mocks the GetInfo method.
	GetInfoFunc func(ctx context.Context) (*lndclient.Info, error)

	// GetNodeInfoFunc mocks the GetNodeInfo method.
	GetNodeInfoFunc func(ctx context.Context, pubkey route.Vertex, includeChannels bool) (*lndclient.NodeInfo, error)

	// ListChannelsFunc mocks the ListChannels method.
	ListChannelsFunc func(ctx context.Context, activeOnly bool, publicOnly bool) ([]lndclient.ChannelInfo, error)

	// UpdateChanPolicyFunc mocks the UpdateChanPolicy method.
	UpdateChanPolicyFunc func(ctx context.Context, req lndclient.PolicyUpdateRequest, chanPoint *wire.OutPoint) error

	// calls tracks calls to the methods.
	calls struct {
		// DescribeGraph holds details about calls to the DescribeGraph method.
		DescribeGraph []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// IncludeUnannounced is the includeUnannounced argument value.
			IncludeUnannounced bool
		}
		// ForwardingHistory holds details about calls to the ForwardingHistory method.
		ForwardingHistory []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Req is the req argument value.
			Req lndclient.ForwardingHistoryRequest
		}
		// GetChanInfo holds details about calls to the GetChanInfo method.
		GetChanInfo []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ChanId is the chanId argument value.
			ChanId uint64
		}
		// GetInfo holds details about calls to the GetInfo method.
		GetInfo []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// GetNodeInfo holds details about calls to the GetNodeInfo method.
		GetNodeInfo []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Pubkey is the pubkey argument value.
			Pubkey route.Vertex
			// IncludeChannels is the includeChannels argument value.
			IncludeChannels bool
		}
		// ListChannels holds details about calls to the ListChannels method.
		ListChannels []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ActiveOnly is the activeOnly argument value.
			ActiveOnly bool
			// PublicOnly is the publicOnly argument value.
			PublicOnly bool
		}
		// UpdateChanPolicy holds details about calls to the UpdateChanPolicy method.
		UpdateChanPolicy []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Req is the req argument value.
			Req lndclient.PolicyUpdateRequest
			// ChanPoint is the chanPoint argument value.
			ChanPoint *wire.OutPoint
		}
	}
	lockDescribeGraph     sync.RWMutex
	lockForwardingHistory sync.RWMutex
	lockGetChanInfo       sync.RWMutex
	lockGetInfo           sync.RWMutex
	lockGetNodeInfo       sync.RWMutex
	lockListChannels      sync.RWMutex
	lockUpdateChanPolicy  sync.RWMutex
}

// DescribeGraph calls DescribeGraphFunc.
func (mock *channelerMock) DescribeGraph(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error) {
	callInfo := struct {
		Ctx                context.Context
		IncludeUnannounced bool
	}{
		Ctx:                ctx,
		IncludeUnannounced: includeUnannounced,
	}
	mock.lockDescribeGraph.Lock()
	mock.calls.DescribeGraph = append(mock.calls.DescribeGraph, callInfo)
	mock.lockDescribeGraph.Unlock()
	if mock.DescribeGraphFunc == nil {
		var (
			graphOut *lndclient.Graph
			errOut   error
		)
		return graphOut, errOut
	}
	return mock.DescribeGraphFunc(ctx, includeUnannounced)
}

// DescribeGraphCalls gets all the calls that were made to DescribeGraph.
// Check the length with:
//
//	len(mockedchanneler.DescribeGraphCalls())
func (mock *channelerMock) DescribeGraphCalls() []struct {
	Ctx                context.Context
	IncludeUnannounced bool
} {
	var calls []struct {
		Ctx                context.Context
		IncludeUnannounced bool
	}
	mock.lockDescribeGraph.RLock()
	calls = mock.calls.DescribeGraph
	mock.lockDescribeGraph.RUnlock()
	return calls
}

// ForwardingHistory calls ForwardingHistoryFunc.
func (mock *channelerMock) ForwardingHistory(ctx context.Context, req lndclient.ForwardingHistoryRequest) (*lndclient.ForwardingHistoryResponse, error) {
	callInfo := struct {
		Ctx context.Context
		Req lndclient.ForwardingHistoryRequest
	}{
		Ctx: ctx,
		Req: req,
	}
	mock.lockForwardingHistory.Lock()
	mock.calls.ForwardingHistory = append(mock.calls.ForwardingHistory, callInfo)
	mock.lockForwardingHistory.Unlock()
	if mock.ForwardingHistoryFunc == nil {
		var (
			forwardingHistoryResponseOut *lndclient.ForwardingHistoryResponse
			errOut                       error
		)
		return forwardingHistoryResponseOut, errOut
	}
	return mock.ForwardingHistoryFunc(ctx, req)
}

// ForwardingHistoryCalls gets all the calls that were made to ForwardingHistory.
// Check the length with:
//
//	len(mockedchanneler.ForwardingHistoryCalls())
func (mock *channelerMock) ForwardingHistoryCalls() []struct {
	Ctx context.Context
	Req lndclient.ForwardingHistoryRequest
} {
	var calls []struct {
		Ctx context.Context
		Req lndclient.ForwardingHistoryRequest
	}
	mock.lockForwardingHistory.RLock()
	calls = mock.calls.ForwardingHistory
	mock.lockForwardingHistory.RUnlock()
	return calls
}

// GetChanInfo calls GetChanInfoFunc.
func (mock *channelerMock) GetChanInfo(ctx context.Context, chanId uint64) (*lndclient.ChannelEdge, error) {
	callInfo := struct {
		Ctx    context.Context
		ChanId uint64
	}{
		Ctx:    ctx,
		ChanId: chanId,
	}
	mock.lockGetChanInfo.Lock()
	mock.calls.GetChanInfo = append(mock.calls.GetChanInfo, callInfo)
	mock.lockGetChanInfo.Unlock()
	if mock.GetChanInfoFunc == nil {
		var (
			channelEdgeOut *lndclient.ChannelEdge
			errOut         error
		)
		return channelEdgeOut, errOut
	}
	return mock.GetChanInfoFunc(ctx, chanId)
}

// GetChanInfoCalls gets all the calls that were made to GetChanInfo.
// Check the length with:
//
//	len(mockedchanneler.GetChanInfoCalls())
func (mock *channelerMock) GetChanInfoCalls() []struct {
	Ctx    context.Context
	ChanId uint64
} {
	var calls []struct {
		Ctx    context.Context
		ChanId uint64
	}
	mock.lockGetChanInfo.RLock()
	calls = mock.calls.GetChanInfo
	mock.lockGetChanInfo.RUnlock()
	return calls
}

// GetInfo calls GetInfoFunc.
func (mock *channelerMock) GetInfo(ctx context.Context) (*lndclient.Info, error) {
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
			infoOut *lndclient.Info
			errOut  error
		)
		return infoOut, errOut
	}
	return mock.GetInfoFunc(ctx)
}

// GetInfoCalls gets all the calls that were made to GetInfo.
// Check the length with:
//
//	len(mockedchanneler.GetInfoCalls())
func (mock *channelerMock) GetInfoCalls() []struct {
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

// GetNodeInfo calls GetNodeInfoFunc.
func (mock *channelerMock) GetNodeInfo(ctx context.Context, pubkey route.Vertex, includeChannels bool) (*lndclient.NodeInfo, error) {
	callInfo := struct {
		Ctx             context.Context
		Pubkey          route.Vertex
		IncludeChannels bool
	}{
		Ctx:             ctx,
		Pubkey:          pubkey,
		IncludeChannels: includeChannels,
	}
	mock.lockGetNodeInfo.Lock()
	mock.calls.GetNodeInfo = append(mock.calls.GetNodeInfo, callInfo)
	mock.lockGetNodeInfo.Unlock()
	if mock.GetNodeInfoFunc == nil {
		var (
			nodeInfoOut *lndclient.NodeInfo
			errOut      error
		)
		return nodeInfoOut, errOut
	}
	return mock.GetNodeInfoFunc(ctx, pubkey, includeChannels)
}

// GetNodeInfoCalls gets all the calls that were made to GetNodeInfo.
// Check the length with:
//
//	len(mockedchanneler.GetNodeInfoCalls())
func (mock *channelerMock) GetNodeInfoCalls() []struct {
	Ctx             context.Context
	Pubkey          route.Vertex
	IncludeChannels bool
} {
	var calls []struct {
		Ctx             context.Context
		Pubkey          route.Vertex
		IncludeChannels bool
	}
	mock.lockGetNodeInfo.RLock()
	calls = mock.calls.GetNodeInfo
	mock.lockGetNodeInfo.RUnlock()
	return calls
}

// ListChannels calls ListChannelsFunc.
func (mock *channelerMock) ListChannels(ctx context.Context, activeOnly bool, publicOnly bool) ([]lndclient.ChannelInfo, error) {
	callInfo := struct {
		Ctx        context.Context
		ActiveOnly bool
		PublicOnly bool
	}{
		Ctx:        ctx,
		ActiveOnly: activeOnly,
		PublicOnly: publicOnly,
	}
	mock.lockListChannels.Lock()
	mock.calls.ListChannels = append(mock.calls.ListChannels, callInfo)
	mock.lockListChannels.Unlock()
	if mock.ListChannelsFunc == nil {
		var (
			channelInfosOut []lndclient.ChannelInfo
			errOut          error
		)
		return channelInfosOut, errOut
	}
	return mock.ListChannelsFunc(ctx, activeOnly, publicOnly)
}

// ListChannelsCalls gets all the calls that were made to ListChannels.
// Check the length with:
//
//	len(mockedchanneler.ListChannelsCalls())
func (mock *channelerMock) ListChannelsCalls() []struct {
	Ctx        context.Context
	ActiveOnly bool
	PublicOnly bool
} {
	var calls []struct {
		Ctx        context.Context
		ActiveOnly bool
		PublicOnly bool
	}
	mock.lockListChannels.RLock()
	calls = mock.calls.ListChannels
	mock.lockListChannels.RUnlock()
	return calls
}

// UpdateChanPolicy calls UpdateChanPolicyFunc.
func (mock *channelerMock) UpdateChanPolicy(ctx context.Context, req lndclient.PolicyUpdateRequest, chanPoint *wire.OutPoint) error {
	callInfo := struct {
		Ctx       context.Context
		Req       lndclient.PolicyUpdateRequest
		ChanPoint *wire.OutPoint
	}{
		Ctx:       ctx,
		Req:       req,
		ChanPoint: chanPoint,
	}
	mock.lockUpdateChanPolicy.Lock()
	mock.calls.UpdateChanPolicy = append(mock.calls.UpdateChanPolicy, callInfo)
	mock.lockUpdateChanPolicy.Unlock()
	if mock.UpdateChanPolicyFunc == nil {
		var (
			errOut error
		)
		return errOut
	}
	return mock.UpdateChanPolicyFunc(ctx, req, chanPoint)
}

// UpdateChanPolicyCalls gets all the calls that were made to UpdateChanPolicy.
// Check the length with:
//
//	len(mockedchanneler.UpdateChanPolicyCalls())
func (mock *channelerMock) UpdateChanPolicyCalls() []struct {
	Ctx       context.Context
	Req       lndclient.PolicyUpdateRequest
	ChanPoint *wire.OutPoint
} {
	var calls []struct {
		Ctx       context.Context
		Req       lndclient.PolicyUpdateRequest
		ChanPoint *wire.OutPoint
	}
	mock.lockUpdateChanPolicy.RLock()
	calls = mock.calls.UpdateChanPolicy
	mock.lockUpdateChanPolicy.RUnlock()
	return calls
}

// routerMock is a mock implementation of router.
//
//	func TestSomethingThatUsesrouter(t *testing.T) {
//
//		// make and configure a mocked router
//		mockedrouter := &routerMock{
//			SendPaymentFunc: func(ctx context.Context, request lndclient.SendPaymentRequest) (chan lndclient.PaymentStatus, chan error, error) {
//				panic("mock out the SendPayment method")
//			},
//			SubscribeHtlcEventsFunc: func(ctx context.Context) (<-chan *routerrpc.HtlcEvent, <-chan error, error) {
//				panic("mock out the SubscribeHtlcEvents method")
//			},
//		}
//
//		// use mockedrouter in code that requires router
//		// and then make assertions.
//
//	}
type routerMock struct {
	// SendPaymentFunc mocks the SendPayment method.
	SendPaymentFunc func(ctx context.Context, request lndclient.SendPaymentRequest) (chan lndclient.PaymentStatus, chan error, error)

	// SubscribeHtlcEventsFunc mocks the SubscribeHtlcEvents method.
	SubscribeHtlcEventsFunc func(ctx context.Context) (<-chan *routerrpc.HtlcEvent, <-chan error, error)

	// calls tracks calls to the methods.
	calls struct {
		// SendPayment holds details about calls to the SendPayment method.
		SendPayment []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Request is the request argument value.
			Request lndclient.SendPaymentRequest
		}
		// SubscribeHtlcEvents holds details about calls to the SubscribeHtlcEvents method.
		SubscribeHtlcEvents []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
	lockSendPayment         sync.RWMutex
	lockSubscribeHtlcEvents sync.RWMutex
}

// SendPayment calls SendPaymentFunc.
func (mock *routerMock) SendPayment(ctx context.Context, request lndclient.SendPaymentRequest) (chan lndclient.PaymentStatus, chan error, error) {
	callInfo := struct {
		Ctx     context.Context
		Request lndclient.SendPaymentRequest
	}{
		Ctx:     ctx,
		Request: request,
	}
	mock.lockSendPayment.Lock()
	mock.calls.SendPayment = append(mock.calls.SendPayment, callInfo)
	mock.lockSendPayment.Unlock()
	if mock.SendPaymentFunc == nil {
		var (
			paymentStatusChOut chan lndclient.PaymentStatus
			errChOut           chan error
			errOut             error
		)
		return paymentStatusChOut, errChOut, errOut
	}
	return mock.SendPaymentFunc(ctx, request)
}

// SendPaymentCalls gets all the calls that were made to SendPayment.
// Check the length with:
//
//	len(mockedrouter.SendPaymentCalls())
func (mock *routerMock) SendPaymentCalls() []struct {
	Ctx     context.Context
	Request lndclient.SendPaymentRequest
} {
	var calls []struct {
		Ctx     context.Context
		Request lndclient.SendPaymentRequest
	}
	mock.lockSendPayment.RLock()
	calls = mock.calls.SendPayment
	mock.lockSendPayment.RUnlock()
	return calls
}

// SubscribeHtlcEvents calls SubscribeHtlcEventsFunc.
func (mock *routerMock) SubscribeHtlcEvents(ctx context.Context) (<-chan *routerrpc.HtlcEvent, <-chan error, error) {
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockSubscribeHtlcEvents.Lock()
	mock.calls.SubscribeHtlcEvents = append(mock.calls.SubscribeHtlcEvents, callInfo)
	mock.lockSubscribeHtlcEvents.Unlock()
	if mock.SubscribeHtlcEventsFunc == nil {
		var (
			htlcEventChOut <-chan *routerrpc.HtlcEvent
			errChOut       <-chan error
			errOut         error
		)
		return htlcEventChOut, errChOut, errOut
	}
	return mock.SubscribeHtlcEventsFunc(ctx)
}

// SubscribeHtlcEventsCalls gets all the calls that were made to SubscribeHtlcEvents.
// Check the length with:
//
//	len(mockedrouter.SubscribeHtlcEventsCalls())
func (mock *routerMock) SubscribeHtlcEventsCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockSubscribeHtlcEvents.RLock()
	calls = mock.calls.SubscribeHtlcEvents
	mock.lockSubscribeHtlcEvents.RUnlock()
	return calls
}

// invoicerMock is a mock implementation of invoicer.
//
//	func TestSomethingThatUsesinvoicer(t *testing.T) {
//
//		// make and configure a mocked invoicer
//		mockedinvoicer := &invoicerMock{
//			AddInvoiceFunc: func(ctx context.Context, in *invoicesrpc.AddInvoiceData) (lntypes.Hash, string, error) {
//				panic("mock out the AddInvoice method")
//			},
//		}
//
//		// use mockedinvoicer in code that requires invoicer
//		// and then make assertions.
//
//	}
type invoicerMock struct {
	// AddInvoiceFunc mocks the AddInvoice method.
	AddInvoiceFunc func(ctx context.Context, in *invoicesrpc.AddInvoiceData) (lntypes.Hash, string, error)

	// calls tracks calls to the methods.
	calls struct {
		// AddInvoice holds details about calls to the AddInvoice method.
		AddInvoice []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// In is the in argument value.
			In *invoicesrpc.AddInvoiceData
		}
	}
	lockAddInvoice sync.RWMutex
}

// AddInvoice calls AddInvoiceFunc.
func (mock *invoicerMock) AddInvoice(ctx context.Context, in *invoicesrpc.AddInvoiceData) (lntypes.Hash, string, error) {
	callInfo := struct {
		Ctx context.Context
		In  *invoicesrpc.AddInvoiceData
	}{
		Ctx: ctx,
		In:  in,
	}
	mock.lockAddInvoice.Lock()
	mock.calls.AddInvoice = append(mock.calls.AddInvoice, callInfo)
	mock.lockAddInvoice.Unlock()
	if mock.AddInvoiceFunc == nil {
		var (
			hashOut lntypes.Hash
			sOut    string
			errOut  error
		)
		return hashOut, sOut, errOut
	}
	return mock.AddInvoiceFunc(ctx, in)
}

// AddInvoiceCalls gets all the calls that were made to AddInvoice.
// Check the length with:
//
//	len(mockedinvoicer.AddInvoiceCalls())
func (mock *invoicerMock) AddInvoiceCalls() []struct {
	Ctx context.Context
	In  *invoicesrpc.AddInvoiceData
} {
	var calls []struct {
		Ctx context.Context
		In  *invoicesrpc.AddInvoiceData
	}
	mock.lockAddInvoice.RLock()
	calls = mock.calls.AddInvoice
	mock.lockAddInvoice.RUnlock()
	return calls
}