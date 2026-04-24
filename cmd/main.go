package main

import (
	"fmt"
	"net"
	"os"

	"github.com/gafkonian-go/internal/config"
	"github.com/gafkonian-go/internal/handler"
	"github.com/gafkonian-go/internal/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to initialize config: ", err.Error())
		return
	}
	address := fmt.Sprintf("%v:%v", "0.0.0.0", cfg.Port)
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Failed to bind to port :%v. Error: %v", cfg.Port, err.Error())
		os.Exit(1)
	}
	fmt.Printf("Server starting on :%v...", cfg.Port)
	defer utils.CloseResource(l)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go handler.HandleConnection(conn, cfg)
	}
}
