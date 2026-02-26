package main

import (
	"log"
	"math/rand/v2"
	"net"
	"os"
	"slices"

	"github.com/mkyas/minichord"
)

// handleTask command a specific node to begin routing random packets.
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

func sendFinger(conn net.Conn, p int32, nr uint32) error {
	ids := make([]int32, 0, len(nodes))
	for v, _ := range nodes {
		ids = append(ids, v)
	}
	slices.Sort(ids)

	fingers := make([]*minichord.Deregistration, 0, nr)
	for i := range nr {
		a := p + 1<<i%MAX_ID
		m := ids[0]
		if m == p {
			m = ids[1]
		}
		for _, id := range ids {
			if id != p && id >= a {
				m = id
				break
			}
		}

		fingers = append(fingers, &minichord.Deregistration{
			Address: nodes[m],
			Id:      m,
		})
	}

	msg := &minichord.MiniChord{
		Message: &minichord.MiniChord_NodeRegistry{
			NodeRegistry: &minichord.NodeRegistry{
				NR:    nr,
				Peers: fingers,
				NoIds: uint32(len(nodes)),
				Ids:   ids,
			},
		},
	}
	return minichord.SendMiniChordMessage(conn, msg)
}

// handleConnection processes inbound TCP connections to the registry.
// It unmarshals the payload and routes it to the specific processing logic.
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

func generateId() int32 {
	if len(nodes) >= int(MAX_ID) {
		log.Println("Exceeded Maximum allowed IDs")
		log.Println("Terminating Registry")
		os.Exit(1)
	}
	for {
		newId := rand.Int32N(MAX_ID + 1)
		if _, exists := nodes[newId]; !exists {
			return newId
		}
	}
}

// handleRegistrationResponse assigns a new ID and registers the messenger node.
func handleRegistrationResponse(conn net.Conn, reg *minichord.Registration) {
	log.Println("Detected node:", reg.Address)
	newId := generateId()
	nodes[newId] = reg.Address
	resp := &minichord.MiniChord{
		Message: &minichord.MiniChord_RegistrationResponse{
			RegistrationResponse: &minichord.RegistrationResponse{
				Result: newId,
				Info:   "Node added to the registry",
			},
		},
	}
	err := minichord.SendMiniChordMessage(conn, resp)
	if err != nil {
		delete(nodes, newId)
		return
	}
}

// handleDeregistrationResponse removes a previously registered node from the overlay.
func handleDeregistrationResponse(conn net.Conn, dereg *minichord.Deregistration) {
	if dereg != nil {
		log.Println("Node:", dereg.Id, "Requests deregistration")
		resp := &minichord.MiniChord{
			Message: &minichord.MiniChord_DeregistrationResponse{
				DeregistrationResponse: &minichord.DeregistrationResponse{
					Result: dereg.Id,
					Info:   "Node removed from registry",
				},
			},
		}
		minichord.SendMiniChordMessage(conn, resp)
		delete(nodes, dereg.Id)
	}
}
