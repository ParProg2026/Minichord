package main

import (
	"sync"
)

// Map from node id to adddress
type Nodes map[int32]string

var nodes = make(Nodes)

// Lock to for processing messages -> avoid collisions
var messageLock sync.Mutex
var wg sync.WaitGroup
var setupWg sync.WaitGroup
var startWg sync.WaitGroup
var summaryWg sync.WaitGroup

// Save the finger tables for all the nodes
var allFingers map[int32][]int32

// Collect the summary for all nodes
type Summary struct {
	id               int32
	sendTracker      uint32
	receiveTracker   uint32
	relayTracker     uint32
	sendSummation    int64
	receiveSummation int64
}

var summaries map[int32]Summary
