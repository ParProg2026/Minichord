package main

import (
	"log"
	"net"
	"os"

	"github.com/mkyas/minichord"
)

func HandleRegistration(conn net.Conn) error {
	msg := &minichord.MiniChord{
		Message: &minichord.MiniChord_Registration{
			Registration: &minichord.Registration{
				Address: nodeAddr,
			},
		},
	}

	err := minichord.SendMiniChordMessage(conn, msg)
	if err != nil {
		return err
	}

	resp, err := minichord.ReceiveMiniChordMessage(conn)
	if err != nil {
		return err
	}

	regResponse = resp.GetRegistrationResponse()
	if regResponse != nil {
		nodeID = regResponse.Result
		log.Println("Registration succesfully. ID:", regResponse.Result, "Info:", regResponse.Info)
	} else {
		log.Println("Registration response failed")
		log.Println("Terminating Process")
		os.Exit(1)
	}
	return nil
}

func HandleDeregistration(conn net.Conn) error {
	msg := &minichord.MiniChord{
		Message: &minichord.MiniChord_Deregistration{
			Deregistration: &minichord.Deregistration{
				Address: nodeAddr,
				Id:      regResponse.Result,
			},
		},
	}
	err := minichord.SendMiniChordMessage(conn, msg)
	if err != nil {
		return err
	}

	resp, err := minichord.ReceiveMiniChordMessage(conn)
	if err != nil {
		return err
	}

	deregResponse := resp.GetDeregistrationResponse()
	if deregResponse != nil {
		log.Println("Deregistration succesfully. ID:", deregResponse.Result, "Info:", deregResponse.Info)
	} else {
		log.Println("deregistration response failed")
		return err
	}
	return nil
}

func HandleRegistryResponse(s int32) func(net.Conn) error {
	return func(conn net.Conn) error {
		msg := &minichord.MiniChord{
			Message: &minichord.MiniChord_NodeRegistryResponse{
				NodeRegistryResponse: &minichord.NodeRegistryResponse{
					Result: s,
					Info:   "Registered them",
				},
			},
		}

		err := minichord.SendMiniChordMessage(conn, msg)
		if err != nil {
			return err
		}
		return nil
	}
}
