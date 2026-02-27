package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Listens to the specified port and accepts new connections
func NodeReceive(hostPort string) {
	listener, err := net.Listen("tcp", ":"+hostPort)
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

// Wrapper to send message to the different nodes
// Establishes the connection to the given address
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

// Responsible for parsing all the user input
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
			allFingers = make(map[int32][]int32)
			setupWg.Add(len(nodes))
			for id, addr := range nodes {
				n, err := strconv.Atoi(fields[1])
				if err != nil {
					log.Println("Invalid argument:", fields[1])
				}
				NodeSend(addr, sendFinger(id, uint32(n)))
			}

			// Wait till all the nodes processed the setup
			setupWg.Wait()
			fmt.Println("The registry is now ready to initiate tasks.")
		case "route":
			// Show the already generated finger tables
			for node, finger := range allFingers {
				fmt.Printf("Node %d:", node)
				for _, id := range finger {
					fmt.Printf(" %d", id)
				}
				fmt.Println()
			}
		case "start":
			// Start the task
			startWg.Add(len(nodes))
			n, err := strconv.Atoi(fields[1])
			if err != nil {
				log.Println("Invalid argument:", fields[1])
			}
			for _, addr := range nodes {
				go NodeSend(addr, handleTask(uint32(n)))
			}

			// Wait till all functions have finished their task
			startWg.Wait()

			summaries = make(map[int32]Summary)
			// Start requesting the summary
			for {
				summaryWg.Add(len(nodes))
				for _, addr := range nodes {
					go NodeSend(addr, handleTrafficSummary())
				}
				summaryWg.Wait()

				// Generate the summed up summary
				var send uint32
				var rec uint32
				var rel uint32
				var sendsum int64
				var recsum int64
				for _, summary := range summaries {
					send += summary.sendTracker
					rec += summary.receiveTracker
					rel += summary.relayTracker
					sendsum += summary.sendSummation
					recsum += summary.receiveSummation
				}

				// If all messages have been received we print them
				// Otherwise we gonna wait for a second and try to get the summary again
				if int(rec) == n*len(nodes) {
					log.Println("Id,Sent,Received,Relayed,Total Sent,Total Received")

					for _, summary := range summaries {
						log.Printf("%v,%v,%v,%v,%v,%v\n",
							summary.id,
							summary.sendTracker,
							summary.receiveTracker,
							summary.relayTracker,
							summary.sendSummation,
							summary.receiveSummation,
						)
					}

					log.Println()
					log.Printf("Sum,%v,%v,%v,%v,%v\n",
						send,
						rec,
						rel,
						sendsum,
						recsum,
					)

					break
				}
				time.Sleep(1 * time.Second)
			}
		case "exit":
			wg.Done()
			return
		default:
			log.Println("Unknown command:", fields[0])
		}
	}
}

func main() {
	// Read in the port and startup the listeners
	hostPort := "2077"
	if len(os.Args) > 1 {
		hostPort = os.Args[1]
	}

	log.Println("Starting registry")
	wg.Add(1)
	go NodeReceive(hostPort)
	go InputParser()
	wg.Wait()
}
