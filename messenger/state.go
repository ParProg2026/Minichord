package main

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/mkyas/minichord"
)

var port = "2077" // TODO: replace with flag
var regResponse *minichord.RegistrationResponse

var wg sync.WaitGroup

var nodeAddr string
var nodeID int32
var userChan = make(chan string, 1)
var regChan = make(chan *minichord.MiniChord, 10000)

type Finger struct {
	Id   int32
	Addr string
}

var fingerTable []Finger
var allNodes []int32

var regConn net.Conn
var openConnections map[int32]net.Conn

var sendTracker atomic.Uint32
var receiveTracker atomic.Uint32

var relayTracker atomic.Uint32

var sendSummation atomic.Int64
var receiveSummation atomic.Int64
