package main

import (
	"github.com/mkyas/minichord"
	"sync"
)

var port = "2077" // TODO: replace with flag
var regResponse *minichord.RegistrationResponse

var wg sync.WaitGroup

var nodeAddr string
var userChan = make(chan string, 1)
var regChan = make(chan *minichord.MiniChord, 100)
