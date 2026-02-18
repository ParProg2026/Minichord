package main

import (
	"github.com/mkyas/minichord"
	"log"
	"math/rand/v2"
	"net"
	"os"
)

func handleTask(conn net.Conn, n uint32) error {
	msg := &minichord.MiniChord{
		Message: &minichord.MiniChord_InitiateTask{
			InitiateTask: &minichord.InitiateTask{
				Packets: n,
			},
		},
	}
	return minichord.SendMiniChordMessage(conn, msg)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	msg, err := minichord.ReceiveMiniChordMessage(conn)
	if err != nil {
		log.Println("Read failed:", err)
		return
	}

	messageLock.Lock()
	switch {
	case msg.GetRegistration() != nil:
		reg := msg.GetRegistration()
		handleRegistrationResponse(conn, reg)

	case msg.GetDeregistration() != nil:
		dereg := msg.GetDeregistration()
		handleDeregistrationResponse(conn, dereg)

	default:
		log.Printf("Unexpected Message type: %T\n", msg.GetMessage())
	}
	messageLock.Unlock()
}

func generateId(usedIDs IDs) int32 {
	if len(usedIDs) >= int(MAX_ID) {
		log.Println("Exceeded Maximum allowed IDs")
		log.Println("Terminating Registry")
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

func handleRegistrationResponse(conn net.Conn, reg *minichord.Registration) {
	log.Println("Detected node:", reg.Address)
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
	minichord.SendMiniChordMessage(conn, resp)
}

func handleDeregistrationResponse(conn net.Conn, dereg *minichord.Deregistration) {
	if dereg != nil {
		nodeId := nodes[dereg.Address]
		log.Println("Node:", nodeId, "Requests dergistration")
		resp := &minichord.MiniChord{
			Message: &minichord.MiniChord_DeregistrationResponse{
				DeregistrationResponse: &minichord.DeregistrationResponse{
					Result: nodeId,
					Info:   "Node removed from registry",
				},
			},
		}
		minichord.SendMiniChordMessage(conn, resp)
		delete(nodes, dereg.Address)
		delete(usedIDs, nodeId)
	}
}
