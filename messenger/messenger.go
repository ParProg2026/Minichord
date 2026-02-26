package main

import (
	"bufio"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"

	"github.com/mkyas/minichord"
)

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

func RegistrySend(fn func(conn net.Conn) error) {
	conn, err := net.Dial("tcp", ":"+port)
	if err != nil {
		log.Fatal("Listener failed:", err)
	}
	defer conn.Close()

	if err := fn(conn); err != nil {
		log.Fatal("Operation failed:", err)
	}
}

func NodeSend(id int32, fn func(conn net.Conn) error) {
	conn := openConnections[id]

	if err := fn(conn); err != nil {
		log.Fatal("Operation failed:", err)
	}
}

func DetermineNextFinger(data *minichord.NodeData) int32 {
	dest := data.Destination
	if dest < nodeID {
		dest += MAX_ID
	}

	biggest := fingerTable[0].Id
	for _, finger := range fingerTable {
		id := finger.Id
		if id < nodeID {
			id += MAX_ID
		}

		if finger.Id < dest {
			biggest = finger.Id
		}
	}
	return biggest % MAX_ID
}

func MessageConnReceive(conn net.Conn) {
	for {
		msg, err := minichord.ReceiveMiniChordMessage(conn)
		if err != nil {
			return
		}
		regChan <- msg
	}
}

func MessageReceive(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error:", err)
			continue
		}

		go MessageConnReceive(conn)
	}
}

func Node() {
	defer wg.Done()
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	go MessageReceive(listener)

	nodeAddr = listener.Addr().String()
	RegistrySend(HandleRegistration)
	log.Println("Node listening on:", nodeAddr)
	for {
		select {
		case userCommand := <-userChan:
			switch userCommand {
			case "print":
			case "exit":
				RegistrySend(HandleDeregistration)
				return
			default:
				log.Println("Unknown command:", userCommand)
			}
		case registryCommand := <-regChan:
			switch {
			case registryCommand.GetInitiateTask() != nil:
				log.Println("Task Received", registryCommand.GetInitiateTask().Packets)

				for range registryCommand.GetInitiateTask().Packets {
					dest := allNodes[rand.Int31n(int32(len(allNodes)))]
					finger := fingerTable[rand.Int31n(int32(len(fingerTable)))]

					msg := &minichord.MiniChord{
						Message: &minichord.MiniChord_NodeData{
							NodeData: &minichord.NodeData{
								Destination: dest,
								Source:      nodeID,
								Payload:     rand.Int31(),
								Hops:        0,
								Trace:       []int32{nodeID},
							},
						},
					}

					NodeSend(finger.Id, func(conn net.Conn) error {
						minichord.SendMiniChordMessage(conn, msg)
						return nil
					})

					if err != nil {
						log.Printf("Error while closin connection: %s", err)
					}
				}

			case registryCommand.GetNodeRegistry() != nil:
				fingerTable = make([]Finger, 0)
				allNodes = make([]int32, 0)
				openConnections = make(map[int32]net.Conn) // Initialize the map
				for _, node := range registryCommand.GetNodeRegistry().Peers {
					fingerTable = append(fingerTable, Finger{Id: node.Id, Addr: node.Address})
					conn, err := net.Dial("tcp", node.Address)
					if err != nil {
						log.Printf("Error while creating connection with node %s: %s ", node.Address, err)
						continue // Skip adding to map if connection failed
					}
					openConnections[node.Id] = conn
				}
				for _, node := range registryCommand.GetNodeRegistry().Ids {
					allNodes = append(allNodes, node)
				}
				RegistrySend(HandleRegistryResponse(0))

			case registryCommand.GetNodeData() != nil:
				data := registryCommand.GetNodeData()
				receiveTracker.Add(1)
				receiveSummation.Add(int64(data.Payload))

				if data.Destination != nodeID {
					sendTracker.Add(1)
					sendSummation.Add(int64(data.Payload))

					next := DetermineNextFinger(data)
					NodeSend(next, handleForwardNodeData(next, data))
				}
			}
		}
	}
}

func main() {
	log.Println("Starting registry")
	wg.Add(1)
	go Node()
	go inputParser()
	wg.Wait()
}
