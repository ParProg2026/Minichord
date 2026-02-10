package main

import (
	"PA2/imports"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
)

const MAX_ID int32 = 1023

type IDs map[int32]struct{}
type Nodes map[string]int32

func generateId(usedIDs IDs) int32 {
	if len(usedIDs) >= int(MAX_ID) {
		fmt.Println("Exceeded Maximum allowed IDs")
		fmt.Println("Terminating Registry")
		os.Exit(1)
	}
	for {
		newId := rand.Int32N(MAX_ID + 1)
		if _, exists := usedIDs[newId]; !exists {
			usedIDs[newId] = struct{}{}
			return newId
		}
	}
}

func Register(hostPort string) {
	listener, _ := net.Listen("tcp", hostPort)
	conn, err := listener.Accept()
	usedIDs := make(IDs)
	nodes := make(Nodes)
	for {
		if err != nil {
			fmt.Println("listener failed:", err)
			return
		}
		defer conn.Close()

		msg, err := minichord.ReceiveMiniChordMessage(conn)
		if err != nil {
			fmt.Println("Read failed:", err)
			return
		}

		reg := msg.GetRegistration()
		if reg != nil {
			println("Detected node:", reg.Address)
			newId := generateId(usedIDs)
			nodes[reg.Address] = newId
			resp := &minichord.MiniChord{
				Message: &minichord.MiniChord_RegistrationResponse{
					RegistrationResponse: &minichord.RegistrationResponse{
						Result: newId,
						Info:   "Node added to the registry",
					},
				},
			}
			_ = minichord.SendMiniChordMessage(conn, resp)
		}

		dereg := msg.GetDeregistration()
		if dereg != nil {
			nodeId := nodes[dereg.Address]
			println("Node:", nodeId, "Requests dergistration")
			resp := &minichord.MiniChord{
				Message: &minichord.MiniChord_DeregistrationResponse{
					DeregistrationResponse: &minichord.DeregistrationResponse{
						Result: nodeId,
						Info:   "Node removed from registry",
					},
				},
			}
			_ = minichord.SendMiniChordMessage(conn, resp)
			delete(nodes, dereg.Address)
			delete(usedIDs, nodeId)
		}
	}
}

func main() {
	hostPort := "localhost:2077" // TODO: Replace with flag
	fmt.Println("Starting message node")
	Register(hostPort)
}
