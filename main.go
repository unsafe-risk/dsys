package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/lemon-mint/godotenv"
)

func main() {
	godotenv.Load()
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	fmt.Println("Listening on port", port)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}
