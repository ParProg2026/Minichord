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

func NodeReceieve(hostPort string) {
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

func NodeSend(addr string, fn func(conn net.Conn, n uint32) error, n uint32) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}
	defer conn.Close()

	if err := fn(conn, n); err != nil {
		log.Fatal("Operation failed:", err)
	}
}

func NodeSend2(addr string, fn func(conn net.Conn, p int32, nr uint32) error, p int32, nr uint32) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal("Dial failed:", err)
	}
	defer conn.Close()

	if err := fn(conn, p, nr); err != nil {
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
				go NodeSend2(addr, sendFinger, id, uint32(n))
			}
		case "route":
			return
		case "start":
			for _, addr := range nodes {
				n, err := strconv.Atoi(fields[1])
				if err != nil {
					log.Println("Invalid argument:", fields[1])
				}
				go NodeSend(addr, handleTask, uint32(n))
			}
		case "exit":
			wg.Done()
			return
		default:
			log.Println("Unkown command:", fields[0])
		}
	}
}

func main() {
	hostPort := "localhost:2077" // TODO: Replace with flag
	log.Println("Starting message node")
	wg.Add(1)
	go NodeReceieve(hostPort)
	go InputParser()
	wg.Wait()
}
