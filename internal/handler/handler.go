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
	customerr "github.com/gafkonian-go/internal/custom_error"
	"github.com/gafkonian-go/internal/utils"
)

func HandleConnection(conn net.Conn, cfg *config.Config) {
	defer utils.CloseResource(conn)
	if err := conn.SetDeadline(time.Now().Add(time.Duration(cfg.TimeoutSeconds) * time.Second)); err != nil {
		fmt.Println("Error while setting the deadline:", err.Error())
		return
	}
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
		var response responseEncoder
		if err != nil {
			fmt.Println("Error parsing header:", err.Error())
			if targetErr, ok := errors.AsType[*customerr.ProtocolError](err); ok {
				response = getErrorResponse(uint16(targetErr.Code), header.CorrelationID)
			}
		} else {
			switch apiKey := header.RequestAPIKey; apiKey {
			case apiVersion:
				fmt.Println("Received API version request...")
				response = getAPIVersionResponse(header.CorrelationID)
			case topicPartitions:
				fmt.Println("Recieved Describe Topic Partitions request...")
				body, err := ParseDescribeTopicBody(payload)
				if err != nil {
					if targetErr, ok := errors.AsType[*customerr.ProtocolError](err); ok {
						response = getErrorResponse(uint16(targetErr.Code), header.CorrelationID)
					}
				} else {
					response = getTopicPartitionsResponse(header.CorrelationID, body)
				}
			case produce:
				fmt.Println("Recieved Produce Record request...")
				body, err := ParseProduceBody(payload)
				if err != nil {
					if targetErr, ok := errors.AsType[*customerr.ProtocolError](err); ok {
						response = getErrorResponse(uint16(targetErr.Code), header.CorrelationID)
					}
				} else {
					response, err = getProduceResponse(header.CorrelationID, body, cfg.PartitionLog)
					if err != nil {
						if targetErr, ok := errors.AsType[*customerr.ProtocolError](err); ok {
							response = getErrorResponse(uint16(targetErr.Code), header.CorrelationID)
						}
					}
				}
			case fetch:
				fmt.Println("Recieved Fetch Record request...")
				body, err := ParseFetchBody(payload)
				if err != nil {
					if targetErr, ok := errors.AsType[*customerr.ProtocolError](err); ok {
						response = getErrorResponse(uint16(targetErr.Code), header.CorrelationID)
					}
				} else {
					response, err = getFetchResponse(header.CorrelationID, body, cfg.PartitionLog)
					if err != nil {
						fmt.Println(err.Error())
						if targetErr, ok := errors.AsType[*customerr.ProtocolError](err); ok {
							response = getErrorResponse(uint16(targetErr.Code), header.CorrelationID)
						}
					}
				}
			}
		}
		_, err = conn.Write(response.encode())
		if err != nil {
			fmt.Println("Error writing a response:", err.Error())
		} else {
			fmt.Println("Response sent! Back to waiting for next request...")
		}
	}
}
