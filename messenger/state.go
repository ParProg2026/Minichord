package main

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/mkyas/minichord"
)

var regAddress string
var regResponse *minichord.RegistrationResponse

var wg sync.WaitGroup

var nodeAddr string
var nodeID int32
var userChan = make(chan string, 1)
var comChan = make(chan *minichord.MiniChord, 100)

// Save the fingers and the existing nodes
type Finger struct {
	Id   int32
	Addr string
}

var fingerTable []Finger
var allNodes []int32

// Connection to register
var regConn net.Conn

// Open connection to other fingers
type Conn struct {
	conn net.Conn
	lock sync.Mutex
}

var openConnections map[int32]*Conn

// All the tracking for the summary
var sendTracker atomic.Uint32
var receiveTracker atomic.Uint32

var relayTracker atomic.Uint32

var sendSummation atomic.Int64
var receiveSummation atomic.Int64
