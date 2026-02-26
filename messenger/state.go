package main

import (
	"fmt"
	"github.com/mkyas/minichord"
	"math/rand"
	"sync"
)

var port = "2077" // TODO: replace with flag
var regResponse *minichord.RegistrationResponse

var wg sync.WaitGroup

var nodeAddr string
var userChan = make(chan string, 1)
var regChan = make(chan *minichord.MiniChord, 100)

// ? I don't know if this is a good idea, but it seems sensible to me not to allow infinite nodes.
const MAX_ID int32 = 1023

// RegistryState manages the active nodes.
type RegistryState struct {
	mu    sync.RWMutex
	nodes map[int32]string
}

// NewRegistryState initializes a new thread-safe registry state.
func NewRegistryState() *RegistryState {
	return &RegistryState{
		nodes: make(map[int32]string),
	}
}

func (rs *RegistryState) RegisterNode(addr string) (int32, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if int32(len(rs.nodes)) >= MAX_ID+1 {
		return -1, fmt.Errorf("registry is full, maximum capacity of %d nodes reached", MAX_ID+1)
	}

	for {
		newID := rand.Int31n(MAX_ID + 1)
		if _, exists := rs.nodes[newID]; !exists {
			rs.nodes[newID] = addr
			return newID, nil
		}
	}
}

// DeregisterNode safely removes a node from the registry by its ID.
func (rs *RegistryState) DeregisterNode(id int32) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	delete(rs.nodes, id)
}

// GetNodesSnapshot returns a deep copy of the current nodes to ensure thread-safe
// iteration without holding the lock during network I/O operations.
// ! Adding this function was suggested by Gemini, for now I'll leave it
// ! commented, but it might actually be useful.
// func (rs *RegistryState) GetNodesSnapshot() map[int32]string {
// 	rs.mu.RLock()
// 	defer rs.mu.RUnlock()

// 	snapshot := make(map[int32]string, len(rs.nodes))
// 	for id, addr := range rs.nodes {
// 		snapshot[id] = addr
// 	}
// 	return snapshot
// }
