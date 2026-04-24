// Package handler implements the protocol request handlers and connection management logic.
package handler

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gafkonian-go/internal/config"
	exc "github.com/gafkonian-go/internal/exceptions"
	"github.com/gafkonian-go/internal/utils"
)

type RequestHeader struct {
	RequestAPIKey     uint16
	RequestAPIVersion uint16
	CorrelationID     uint32
}

func (h *RequestHeader) Validate() error {
	if h.RequestAPIKey != 18 {
		return exc.RaiseError(exc.UnsupportedAPIKeyError, h.RequestAPIKey)
	}
	if h.RequestAPIVersion > 4 {
		return exc.RaiseError(exc.UnsupportedAPIVersionError, h.RequestAPIVersion)
	}
	return nil
}

func ParseHeader(data []byte) (*RequestHeader, error) {
	if len(data) < 8 {
		return nil, exc.RaiseError(exc.InsufficientHeaderError, len(data))
	}
	header := &RequestHeader{
		RequestAPIKey:     binary.BigEndian.Uint16(data[0:2]),
		RequestAPIVersion: binary.BigEndian.Uint16(data[2:4]),
		CorrelationID:     binary.BigEndian.Uint32(data[4:8]),
	}
	if err := header.Validate(); err != nil {
		return header, err
	}
	return header, nil
}

func HandleConnection(conn net.Conn, cfg *config.Config) {
	if err := conn.SetDeadline(time.Now().Add(time.Duration(cfg.TimeoutSeconds) * time.Second)); err != nil {
		fmt.Println("Error while setting the deadline:", err.Error())
		return
	}
	defer utils.CloseResource(conn)
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
		header, err := ParseHeader(payload)
		var errorCode int16
		if err != nil {
			fmt.Println("Error parsing header:", err.Error())
			if targetErr, ok := errors.AsType[*exc.ProtocolError](err); ok {
				errorCode = targetErr.Code
			} else {
				return
			}
		}
		responseSize := 8
		if errorCode != 0 {
			responseSize += 2
		}
		response := make([]byte, responseSize)
		binary.BigEndian.PutUint32(response[0:4], uint32(len(response)-4))
		binary.BigEndian.PutUint32(response[4:8], header.CorrelationID)
		if responseSize == 10 {
			binary.BigEndian.PutUint16(response[8:10], uint16(errorCode))
		}
		_, err = conn.Write(response)
		if err != nil {
			fmt.Println("Error writing a response:", err.Error())
		} else {
			fmt.Println("Response sent! Back to waiting for next request...")
		}
	}
}
