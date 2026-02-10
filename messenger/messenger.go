package main

import (
	"PA2/imports"
	"fmt"
	"net"
	"os"
)

func Register(port string, conn net.Conn) *minichord.RegistrationResponse {
	msg := &minichord.MiniChord{
		Message: &minichord.MiniChord_Registration{
			Registration: &minichord.Registration{
				Address: "localhost:" + port,
			},
		},
	}

	err := minichord.SendMiniChordMessage(conn, msg)
	if err != nil {
		panic(err)
	}

	resp, err := minichord.ReceiveMiniChordMessage(conn)
	if err != nil {
		panic(err)
	}

	regResponse := resp.GetRegistrationResponse()
	if regResponse != nil {
		fmt.Println("Registration succesfully. ID:", regResponse.Result, "Info:", regResponse.Info)
	} else {
		fmt.Println("Registration response failed")
		fmt.Println("Terminating Process")
		os.Exit(1)
	}
	return regResponse
}

func Deregister(port string, conn net.Conn, regResponse *minichord.RegistrationResponse) {
	msg := &minichord.MiniChord{
		Message: &minichord.MiniChord_Deregistration{
			Deregistration: &minichord.Deregistration{
				Address: "localhost:" + port,
				Id:      regResponse.Result,
			},
		},
	}
	err := minichord.SendMiniChordMessage(conn, msg)

	if err != nil {
		panic(err)
	}

	resp, err := minichord.ReceiveMiniChordMessage(conn)
	if err != nil {
		panic(err)
	}

	deregResponse := resp.GetRegistrationResponse()
	if deregResponse != nil {
		fmt.Println("Deregistration succesfully. ID:", deregResponse.Result, "Info:", deregResponse.Info)
		fmt.Println("Terminating Process")
		os.Exit(0)
	} else {
		fmt.Println("deregistration response failed")
		fmt.Println("Terminating Process")
		os.Exit(1)
	}
}

func Messenger(port string) {
	conn, err := net.Dial("tcp", ":"+port)
	if err != nil {
		fmt.Println("Listener failed:", err)
		return
	}
	defer conn.Close()

	regResponse := Register(port, conn)
	defer Deregister(port, conn, regResponse)
}

func main() {
	port := "2077" // TODO: replace with flag
	fmt.Println("Starting registry")
	Messenger(port)
}
