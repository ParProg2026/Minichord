package main

import (
	"sync"
)

type Nodes map[int32]string

var messageLock sync.Mutex
var wg sync.WaitGroup
var startWg sync.WaitGroup
var summaryWg sync.WaitGroup
var nodes = make(Nodes)

type Summary struct {
	id               int32
	sendTracker      uint32
	receiveTracker   uint32
	sendSummation    int64
	receiveSummation int64
}

var summaries []Summary
