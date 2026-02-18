package main

import (
	"bufio"
	"github.com/mkyas/minichord"
	"log"
	"net"
	"os"
	"strings"
)

func inputParser() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
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
				return
			case "exit":
				RegistrySend(HandleDeregistration)
				return
			default:
				log.Println("Unkown command:", userCommand)
			}
		case registryCommand := <-regChan:
			switch {
			case registryCommand.GetInitiateTask() != nil:
				log.Println("Task Received")
			}
			return
		default:
			continue
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
