package main

import "sync"

type IDs map[int32]struct{}
type Nodes map[string]int32

var messageLock sync.Mutex
var wg sync.WaitGroup
var usedIDs = make(IDs)
var nodes = make(Nodes)
