package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/gafkonian-go/internal/config"
)

func CloseResource(r io.Closer) {
	if err := r.Close(); err != nil {
		fmt.Println("Error closing resource:", err.Error())
	}
}

type RequestHeader struct {
	MessageSize       int32
	RequestAPIKey     int16
	RequestAPIVersion int16
	CorrelationID     int32
}

func (h *RequestHeader) Validate() error {
	if h.RequestAPIKey != 18 {
		return fmt.Errorf("unsupported API key %v", h.RequestAPIKey)
	}
	if h.RequestAPIVersion < 0 || h.RequestAPIVersion > 4 {
		return fmt.Errorf("unsupported API version %v", h.RequestAPIVersion)
	}
	return nil
}

func ParseHeader(msgSize int32, data []byte) (*RequestHeader, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("insufficient data for header: %v < 8", len(data))
	}
	header := &RequestHeader{
		MessageSize:       msgSize,
		RequestAPIKey:     int16(binary.BigEndian.Uint16(data[0:2])),
		RequestAPIVersion: int16(binary.BigEndian.Uint16(data[2:4])),
		CorrelationID:     int32(binary.BigEndian.Uint32(data[4:8])),
	}
	if err := header.Validate(); err != nil {
		return nil, err
	}
	return header, nil
}

func handleConnection(conn net.Conn, cfg *config.Config) {
	if err := conn.SetDeadline(time.Now().Add(time.Duration(cfg.TimeoutSeconds) * time.Second)); err != nil {
		fmt.Println("Error while setting the deadline:", err.Error())
		return
	}
	defer CloseResource(conn)
	for {
		sizeBuf := make([]byte, 4)
		_, err := io.ReadFull(conn, sizeBuf)
		if err != nil {
			return
		}
		msgSize := binary.BigEndian.Uint32(sizeBuf)
		payload := make([]byte, msgSize)
		_, err = io.ReadFull(conn, payload)
		if err != nil {
			fmt.Println("Error reading payload:", err.Error())
			return
		}
		header, err := ParseHeader(int32(msgSize), payload)
		if err != nil {
			fmt.Println("Error parsing header:", err.Error())
			return
		}
		response := make([]byte, 8)
		binary.BigEndian.PutUint32(response[0:4], 4)
		binary.BigEndian.PutUint32(response[4:8], uint32(header.CorrelationID))
		_, err = conn.Write(response)
		if err != nil {
			fmt.Println("Error writing a response:", err.Error())
		} else {
			fmt.Println("Response sent! Back to waiting for next request...")
		}

	}
}

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
	defer CloseResource(l)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go handleConnection(conn, cfg)
	}
}
