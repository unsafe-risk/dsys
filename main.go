package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/lemon-mint/godotenv"
)

func main() {
	godotenv.Load()

	var err error
	bootstrapJSON, err := os.ReadFile("./bootstrap.json")
	if err != nil {
		panic(err)
	}
	var bootstrapAddresses []string
	err = json.Unmarshal(bootstrapJSON, &bootstrapAddresses)
	if err != nil {
		panic(err)
	}
	var ln net.Listener
	for _, bootstrapAddress := range bootstrapAddresses {
		ln, err = net.Listen("tcp", bootstrapAddress)
		if err != nil {
			continue
		}
		break
	}
	if ln == nil {
		panic(err)
	}
	defer ln.Close()
	fmt.Println("Listening on", ln.Addr())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}
