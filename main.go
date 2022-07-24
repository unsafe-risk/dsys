package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lemon-mint/godotenv"
	"github.com/unsafe-risk/dsys/multicast"
)

func main() {
	godotenv.Load()
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	ip := ln.Addr().(*net.TCPAddr).IP.String()
	fmt.Println("Listening on port", port)

	mdis := multicast.New(multicast.New_AddrInfo(uint16(port), ip))
	err = mdis.Start()
	if err != nil {
		panic(err)
	}
	defer mdis.Stop()
	time.AfterFunc(time.Minute, func() {
		mdis.Stop()
	})

	go func() {
		for addr := range mdis.C {
			fmt.Println("Peer:", addr)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}
