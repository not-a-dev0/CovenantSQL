/*
 * Copyright 2018 The CovenantSQL Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package transport

import (
	"context"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/CovenantSQL/CovenantSQL/crypto/asymmetric"
	"github.com/CovenantSQL/CovenantSQL/kayak"
	"github.com/CovenantSQL/CovenantSQL/proto"
	"github.com/CovenantSQL/CovenantSQL/twopc"
	"github.com/CovenantSQL/CovenantSQL/utils/log"
	"github.com/jordwest/mock-conn"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

type TestConn struct {
	*mock_conn.End
	peerNodeID proto.NodeID
}

type TestStreamRouter struct {
	sync.Mutex
	streamMap map[proto.NodeID]*TestStream
}

type TestStream struct {
	nodeID proto.NodeID
	router *TestStreamRouter
	queue  chan *TestConn
}

func NewTestStreamRouter() *TestStreamRouter {
	return &TestStreamRouter{
		streamMap: make(map[proto.NodeID]*TestStream),
	}
}

func NewTestStream(nodeID proto.NodeID, router *TestStreamRouter) *TestStream {
	return &TestStream{
		nodeID: nodeID,
		router: router,
		queue:  make(chan *TestConn),
	}
}

func NewSocketPair(fromNode proto.NodeID, toNode proto.NodeID) (clientConn *TestConn, serverConn *TestConn) {
	conn := mock_conn.NewConn()
	clientConn = NewTestConn(conn.Server, toNode)
	serverConn = NewTestConn(conn.Client, fromNode)
	return
}

func NewTestConn(endpoint *mock_conn.End, peerNodeID proto.NodeID) *TestConn {
	return &TestConn{
		End:        endpoint,
		peerNodeID: peerNodeID,
	}
}

func (r *TestStreamRouter) Get(id proto.NodeID) *TestStream {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.streamMap[id]; !ok {
		r.streamMap[id] = NewTestStream(id, r)
	}

	return r.streamMap[id]
}

func (c *TestConn) GetPeerNodeID() proto.NodeID {
	return c.peerNodeID
}

func (s *TestStream) Accept() (conn ConnWithPeerNodeID, err error) {
	select {
	case conn := <-s.queue:
		return conn, nil
	}
}

func (s *TestStream) Dial(ctx context.Context, nodeID proto.NodeID) (conn ConnWithPeerNodeID, err error) {
	clientConn, serverConn := NewSocketPair(s.nodeID, nodeID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.router.Get(nodeID).queue <- serverConn:
	}

	return clientConn, nil
}

// MockWorker is an autogenerated mock type for the Worker type
type MockWorker struct {
	mock.Mock
}

// Commit provides a mock function with given fields: ctx, wb
func (_m *MockWorker) Commit(ctx context.Context, wb twopc.WriteBatch) (interface{}, error) {
	ret := _m.Called(context.Background(), wb)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(context.Context, twopc.WriteBatch) interface{}); ok {
		r0 = rf(ctx, wb)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, twopc.WriteBatch) error); ok {
		r1 = rf(ctx, wb)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Prepare provides a mock function with given fields: ctx, wb
func (_m *MockWorker) Prepare(ctx context.Context, wb twopc.WriteBatch) error {
	ret := _m.Called(context.Background(), wb)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, twopc.WriteBatch) error); ok {
		r0 = rf(ctx, wb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rollback provides a mock function with given fields: ctx, wb
func (_m *MockWorker) Rollback(ctx context.Context, wb twopc.WriteBatch) error {
	ret := _m.Called(context.Background(), wb)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, twopc.WriteBatch) error); ok {
		r0 = rf(ctx, wb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type CallCollector struct {
	l         sync.Mutex
	callOrder []string
}

func (c *CallCollector) Append(call string) {
	c.l.Lock()
	defer c.l.Unlock()
	c.callOrder = append(c.callOrder, call)
}

func (c *CallCollector) Get() []string {
	c.l.Lock()
	defer c.l.Unlock()
	return c.callOrder[:]
}

func (c *CallCollector) Reset() {
	c.l.Lock()
	defer c.l.Unlock()
	c.callOrder = c.callOrder[:0]
}

func testPeersFixture(term uint64, servers []*kayak.Server) *kayak.Peers {
	testPriv := []byte{
		0xea, 0xf0, 0x2c, 0xa3, 0x48, 0xc5, 0x24, 0xe6,
		0x39, 0x26, 0x55, 0xba, 0x4d, 0x29, 0x60, 0x3c,
		0xd1, 0xa7, 0x34, 0x7d, 0x9d, 0x65, 0xcf, 0xe9,
		0x3c, 0xe1, 0xeb, 0xff, 0xdc, 0xa2, 0x26, 0x94,
	}
	privKey, pubKey := asymmetric.PrivKeyFromBytes(testPriv)

	newServers := make([]*kayak.Server, 0, len(servers))
	var leaderNode *kayak.Server

	for _, s := range servers {
		newS := &kayak.Server{
			Role:   s.Role,
			ID:     s.ID,
			PubKey: pubKey,
		}
		newServers = append(newServers, newS)
		if newS.Role == proto.Leader {
			leaderNode = newS
		}
	}

	peers := &kayak.Peers{
		Term:    term,
		Leader:  leaderNode,
		Servers: servers,
		PubKey:  pubKey,
	}

	peers.Sign(privKey)

	return peers
}

func testLogFixture(data []byte) (log *kayak.Log) {
	log = &kayak.Log{
		Index: uint64(1),
		Term:  uint64(1),
		Data:  data,
	}

	log.ComputeHash()

	return
}

func TestConnPair(t *testing.T) {
	Convey("test transport", t, FailureContinues, func(c C) {
		router := NewTestStreamRouter()
		stream1 := router.Get("id1")
		stream2 := router.Get("id2")

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			clientConn, err := stream1.Dial(context.Background(), "id2")
			c.So(err, ShouldBeNil)
			_, err = clientConn.Write([]byte("test"))
			c.So(err, ShouldBeNil)
			clientConn.Close()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			var err error
			serverConn, err := stream2.Accept()
			c.So(err, ShouldBeNil)
			buffer, err := ioutil.ReadAll(serverConn)
			c.So(err, ShouldBeNil)
			c.So(buffer, ShouldResemble, []byte("test"))
		}()

		wg.Wait()
	})
}

func TestTransport(t *testing.T) {
	Convey("test transport", t, FailureContinues, func(c C) {
		router := NewTestStreamRouter()
		stream1 := router.Get("id1")
		stream2 := router.Get("id2")
		config1 := NewConfig("id1", stream1)
		config2 := NewConfig("id2", stream2)
		t1 := NewTransport(config1)
		t2 := NewTransport(config2)
		testLog := testLogFixture([]byte("test request"))

		var err error

		// init
		err = t1.Init()
		So(err, ShouldBeNil)
		err = t2.Init()
		So(err, ShouldBeNil)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := t1.Request(context.Background(), "id2", "test method", testLog)
			c.So(err, ShouldBeNil)
			c.So(res, ShouldResemble, []byte("test response"))
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case req := <-t2.Process():
				c.So(req.GetLog(), ShouldResemble, testLog)
				c.So(req.GetMethod(), ShouldEqual, "test method")
				c.So(req.GetPeerNodeID(), ShouldEqual, proto.NodeID("id1"))
				req.SendResponse([]byte("test response"), nil)
			}
		}()

		wg.Wait()

		// shutdown transport
		err = t1.Shutdown()
		So(err, ShouldBeNil)
		err = t2.Shutdown()
		So(err, ShouldBeNil)
	})
}

func TestIntegration(t *testing.T) {
	type createMockRes struct {
		runner    *kayak.TwoPCRunner
		transport *NetworkTransport
		worker    *MockWorker
		config    *kayak.TwoPCConfig
		runtime   *kayak.Runtime
	}

	// router is a dummy channel based local rpc transport router
	mockRouter := NewTestStreamRouter()

	// peers is a simple 3-node peer configuration
	peers := testPeersFixture(1, []*kayak.Server{
		{
			Role: proto.Leader,
			ID:   "leader",
		},
		{
			Role: proto.Follower,
			ID:   "follower1",
		},
		{
			Role: proto.Follower,
			ID:   "follower2",
		},
	})
	// create mock returns basic arguments to prepare for a server
	createMock := func(nodeID proto.NodeID) (res *createMockRes) {
		res = &createMockRes{}
		log.SetLevel(log.FatalLevel)
		d, _ := ioutil.TempDir("", "kayak_test")

		// runner instance
		res.runner = kayak.NewTwoPCRunner()
		// transport for this instance
		res.transport = NewTransport(NewConfig(nodeID, mockRouter.Get(nodeID)))
		// underlying worker
		res.worker = &MockWorker{}
		// runner config including timeout settings, commit log storage, local server id
		res.config = &kayak.TwoPCConfig{
			RuntimeConfig: kayak.RuntimeConfig{
				RootDir:        d,
				LocalID:        nodeID,
				Runner:         res.runner,
				Transport:      res.transport,
				ProcessTimeout: time.Millisecond * 800,
			},
			Storage: res.worker,
		}
		res.runtime, _ = kayak.NewRuntime(res.config, peers)
		return
	}
	// cleanup log storage after execution
	cleanupDir := func(c *createMockRes) {
		os.RemoveAll(c.config.RuntimeConfig.RootDir)
	}

	Convey("integration test", t, FailureContinues, func(c C) {
		var err error

		lMock := createMock("leader")
		f1Mock := createMock("follower1")
		f2Mock := createMock("follower2")
		defer cleanupDir(lMock)
		defer cleanupDir(f1Mock)
		defer cleanupDir(f2Mock)

		// init
		err = lMock.runtime.Init()
		So(err, ShouldBeNil)
		err = f1Mock.runtime.Init()
		So(err, ShouldBeNil)
		err = f2Mock.runtime.Init()
		So(err, ShouldBeNil)

		// payload to send
		testPayload := []byte("test data")

		// underlying worker mock, prepare/commit/rollback with be received the decoded data
		callOrder := &CallCollector{}
		f1Mock.worker.On("Prepare", mock.Anything, testPayload).
			Return(nil).Run(func(args mock.Arguments) {
			callOrder.Append("prepare")
		})
		f2Mock.worker.On("Prepare", mock.Anything, testPayload).
			Return(nil).Run(func(args mock.Arguments) {
			callOrder.Append("prepare")
		})
		f1Mock.worker.On("Commit", mock.Anything, testPayload).
			Return(nil, nil).Run(func(args mock.Arguments) {
			callOrder.Append("commit")
		})
		f2Mock.worker.On("Commit", mock.Anything, testPayload).
			Return(nil, nil).Run(func(args mock.Arguments) {
			callOrder.Append("commit")
		})
		lMock.worker.On("Prepare", mock.Anything, testPayload).
			Return(nil).Run(func(args mock.Arguments) {
			callOrder.Append("prepare")
		})
		lMock.worker.On("Commit", mock.Anything, testPayload).
			Return(nil, nil).Run(func(args mock.Arguments) {
			callOrder.Append("commit")
		})

		// process the encoded data
		_, _, err = lMock.runtime.Apply(testPayload)
		So(err, ShouldBeNil)
		So(callOrder.Get(), ShouldResemble, []string{
			"prepare",
			"prepare",
			"prepare",
			"commit",
			"commit",
			"commit",
		})

		// process the encoded data again
		callOrder.Reset()
		_, _, err = lMock.runtime.Apply(testPayload)
		So(err, ShouldBeNil)
		So(callOrder.Get(), ShouldResemble, []string{
			"prepare",
			"prepare",
			"prepare",
			"commit",
			"commit",
			"commit",
		})

		// shutdown
		lMock.runtime.Shutdown()
		f1Mock.runtime.Shutdown()
		f2Mock.runtime.Shutdown()
	})
}
