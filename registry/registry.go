package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func NodeReceive(hostPort string) {
	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		log.Fatal("listener failed:", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection failed:", err)
			return
		}
		go handleConnection(conn)
	}
}

func NodeSend(addr string, fn func(conn net.Conn) error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}
	defer conn.Close()

	if err := fn(conn); err != nil {
		log.Fatal("Operation failed:", err)
	}
}

func InputParser() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "list":
			for id, addr := range nodes {
				fmt.Printf("hostname:port : %s | Id : %d\n", addr, id)
			}
		case "setup":
			for id, addr := range nodes {
				n, err := strconv.Atoi(fields[1])
				if err != nil {
					log.Println("Invalid argument:", fields[1])
				}
				go NodeSend(addr, sendFinger(id, uint32(n)))
			}
		case "route":
			return
		case "start":
			startWg.Add(len(nodes))
			for _, addr := range nodes {
				n, err := strconv.Atoi(fields[1])
				if err != nil {
					log.Println("Invalid argument:", fields[1])
				}
				go NodeSend(addr, handleTask(uint32(n)))
			}
			startWg.Wait()

			summaries = make([]Summary, 0)
			summaryWg.Add(len(nodes))
			for _, addr := range nodes {
				go NodeSend(addr, handleTrafficSummary())
			}
			summaryWg.Wait()

			log.Println("Id,Sent,Received,Relayed,Total Sent,Total Received")

			var send uint32
			var rec uint32
			var sendsum int64
			var recsum int64
			for _, summary := range summaries {
				log.Printf("%v,%v,%v,%v,%v,%v\n",
					summary.id,
					summary.sendTracker,
					summary.receiveTracker,
					summary.relayTracker,
					summary.sendSummation,
					summary.receiveSummation,
				)
				send += summary.sendTracker
				rec += summary.receiveTracker
				sendsum += summary.sendSummation
				recsum += summary.receiveSummation
			}

			log.Println()
			log.Printf("%v,%v,%v,%v\n",
				send,
				rec,
				sendsum,
				recsum,
			)
		case "exit":
			wg.Done()
			return
		default:
			log.Println("Unknown command:", fields[0])
		}
	}
}

func main() {
	hostPort := "localhost:2077" // TODO: Replace with flag
	log.Println("Starting message node")
	wg.Add(1)
	go NodeReceive(hostPort)
	go InputParser()
	wg.Wait()
}
