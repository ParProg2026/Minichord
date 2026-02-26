package main

import (
	"sync"

	"github.com/mkyas/minichord"
)

var port = "2077" // TODO: replace with flag
var regResponse *minichord.RegistrationResponse

var wg sync.WaitGroup

var nodeAddr string
var nodeID int32
var userChan = make(chan string, 1)
var regChan = make(chan *minichord.MiniChord, 100)
