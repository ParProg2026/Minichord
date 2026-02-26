package main

import (
	"bufio"
	"log"
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

func RegistryReceive(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error:", err)
			continue
		}

		msg, err := minichord.ReceiveMiniChordMessage(conn)
		if err != nil {
			log.Println("Read failed:", err)
			return
		}
		regChan <- msg
	}
}

func Node() {
	defer wg.Done()
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	go RegistryReceive(listener)

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

			case registryCommand.GetNodeRegistry() != nil:
				log.Println("Node Registry Received, my id is", nodeID)
				for _, node := range registryCommand.GetNodeRegistry().Peers {
					log.Println("Node:", node.Address, "ID:", node.Id)
				}

				// TODO each messaging node should initiate connections to the nodes that comprise its finger table.
				// TODO Every messaging node must report to the registry on the status of setting up connections to nodes that are part of its finger table
				// don't know what to report, but hey here you can report something:
				RegistrySend(HandleRegistryResponse(0))
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
