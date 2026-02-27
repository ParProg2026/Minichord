package main

import (
	"bufio"
	"log"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/mkyas/minichord"
)

// go routine takes inputs and passes them to the go routine in Node
func inputParser() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		cmd := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if cmd == "" {
			continue
		}
		userChan <- cmd
	}
}

// Takes functions to send to the Registry, with error handling
func RegistrySend(fn func(conn net.Conn) error) {
	if err := fn(regConn); err != nil {
		log.Fatal("Operation failed:", err)
	}
}

// Takes functions to send at targeted Nodes, with error handling
func NodeSend(id int32, fn func(conn net.Conn) error) {
	openConnections[id].lock.Lock()
	if err := fn(openConnections[id].conn); err != nil {
		log.Fatal("Operation failed:", err)
	}
	openConnections[id].lock.Unlock()
}

func DetermineNextFinger(data *minichord.NodeData) int32 {
	dest := data.Destination
	// Ensure we can increment forward towards the destination
	if dest < nodeID {
		dest += MAX_ID
	}

	biggest := fingerTable[0].Id
	for _, finger := range fingerTable {
		id := finger.Id
		// ensure IDs loop back around
		if id < nodeID {
			id += MAX_ID
		}

		// find latest node not greater than dest
		if id < dest {
			biggest = finger.Id
		}
	}
	// correct for looping increase
	return biggest % MAX_ID
}

// Collect non-user commands for Node
func MessageConnReceive(conn net.Conn) {
	for {
		msg, err := minichord.ReceiveMiniChordMessage(conn)
		if err != nil {
			return
		}
		comChan <- msg
	}
}

// begin listening
func MessageReceive(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}

		go MessageConnReceive(conn)
	}
}

func Node() {
	// start connection
	defer wg.Done()
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	go MessageReceive(listener)

	// fetch address and register
	nodeAddr = listener.Addr().String()
	RegistrySend(HandleRegistration)
	log.Println("Node listening on:", nodeAddr)
	for {
		// handle commands
		select {
		case userCommand := <-userChan:
			switch userCommand {
			case "print":
				log.Printf("============== NODE %d ==============", nodeID)
				log.Printf("|\tsendTracker: %d\t\t\t\t|", sendTracker.Load())
				log.Printf("|\trecvTracker: %d\t\t\t\t|", receiveTracker.Load())
				log.Printf("|\trelayTracker: %d\t\t\t\t|", relayTracker.Load())
				log.Printf("|\tsendSummation: %d\t\t\t\t|", sendSummation.Load())
				log.Printf("|\trecvSummation: %d\t\t\t\t|", receiveSummation.Load())
				log.Printf("=====================================")
			case "exit":
				RegistrySend(HandleDeregistration)
				return
			default:
				log.Println("Unknown command:", userCommand)
			}
		case registryCommand := <-comChan:
			switch {
			case registryCommand.GetInitiateTask() != nil:
				log.Println("Task Received", registryCommand.GetInitiateTask().Packets)

				// send in a separate go routine, to not block receive
				go func() {
					for range registryCommand.GetInitiateTask().Packets {
						var dest int32
						for {
							dest = allNodes[rand.Int31n(int32(len(allNodes)))]
							if dest != nodeID {
								break
							}
						}
						finger := fingerTable[rand.Int31n(int32(len(fingerTable)))]
						payload := rand.Int31()
						msg := &minichord.MiniChord{
							Message: &minichord.MiniChord_NodeData{
								NodeData: &minichord.NodeData{
									Destination: dest,
									Source:      nodeID,
									Payload:     payload,
									Hops:        0,
									Trace:       []int32{},
								},
							},
						}
						sendTracker.Add(1)
						sendSummation.Add(int64(payload))

						NodeSend(finger.Id, func(conn net.Conn) error {
							minichord.SendMiniChordMessage(conn, msg)
							return nil
						})
						runtime.Gosched()
					}

					RegistrySend(handleTaskFinished)
				}()

			case registryCommand.GetNodeRegistry() != nil:
				fingerTable = make([]Finger, 0)
				allNodes = make([]int32, 0)
				openConnections = make(map[int32]*Conn) // Initialize the map
				for _, node := range registryCommand.GetNodeRegistry().Peers {
					fingerTable = append(fingerTable, Finger{Id: node.Id, Addr: node.Address})
					conn, err := net.Dial("tcp", node.Address)
					if err != nil {
						log.Printf("Error while creating connection with node %s: %s ", node.Address, err)
						continue // Skip adding to map if connection failed
					}
					openConnections[node.Id] = &Conn{
						conn: conn,
						lock: sync.Mutex{},
					}
				}
				for _, node := range registryCommand.GetNodeRegistry().Ids {
					allNodes = append(allNodes, node)
				}
				RegistrySend(HandleRegistryResponse(0))

			case registryCommand.GetNodeData() != nil:
				data := registryCommand.GetNodeData()

				if data.Destination != nodeID {
					next := DetermineNextFinger(data)
					relayTracker.Add(1)
					NodeSend(next, handleForwardNodeData(next, data))
				} else {
					receiveTracker.Add(1)
					receiveSummation.Add(int64(data.Payload))
				}

			// send summary and reset
			case registryCommand.GetRequestTrafficSummary() != nil:
				RegistrySend(HandleSendSummary)
			}
		}
	}
}

func main() {
	log.Println("Starting messenger node")

	regAddress = "localhost:2077"
	if len(os.Args) > 1 {
		regAddress = os.Args[1]
	}

	conn, err := net.Dial("tcp", regAddress)
	if err != nil {
		log.Fatal("Listener failed:", err)
	}
	regConn = conn

	wg.Add(1)
	go Node()
	go inputParser()
	wg.Wait()
}
