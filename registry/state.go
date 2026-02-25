package main

import "sync"

type Nodes map[int32]string

var messageLock sync.Mutex
var wg sync.WaitGroup
var nodes = make(Nodes)
