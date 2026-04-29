package handler

import (
	"encoding/binary"
)

type APIKey uint16

const (
	apiVersion      = 18
	topicPartitions = 75
)

var avaliableAPI = []apiVersionKey{
	{
		APIKey:     apiVersion,
		MinVersion: 0,
		MaxVersion: 4,
		TagBuffer:  0,
	},
	{
		APIKey:     topicPartitions,
		MinVersion: 0,
		MaxVersion: 4,
		TagBuffer:  0,
	},
}

type baseResponse struct {
	CorrelationID uint32
	ErrorCode     uint16
}

// Error Response hadnling
type errorResponse struct {
	baseResponse
}

type responseEncoder interface {
	encode() []byte
}

func (e *errorResponse) encode() []byte {
	buf := make([]byte, 10)
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(buf)-4))
	binary.BigEndian.PutUint32(buf[4:8], e.CorrelationID)
	binary.BigEndian.PutUint16(buf[8:10], e.ErrorCode)
	return buf
}

func getErrorResponse(errorCode uint16, correlationID uint32) responseEncoder {
	response := &errorResponse{
		baseResponse: baseResponse{
			CorrelationID: correlationID,
			ErrorCode:     errorCode,
		},
	}
	return response
}

// API version response handling
type apiVersionKey struct {
	APIKey     uint16
	MinVersion uint16
	MaxVersion uint16
	TagBuffer  uint8
}

type apiVersionResponse struct {
	baseResponse
	APIKeysArrayLength uint8
	APIKeys            []apiVersionKey
	ThrottleTimeMs     uint32
	TagBuffer          uint8
}

func (a *apiVersionResponse) encode() []byte {
	body := make([]byte, 0)

	buf16 := make([]byte, 2)
	buf32 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf32, a.CorrelationID)
	body = append(body, buf32...)
	binary.BigEndian.PutUint16(buf16, 0)
	body = append(body, buf16...)

	body = append(body, uint8(len(a.APIKeys)+1))

	apiKeyBuf := make([]byte, 2)
	for _, apiKey := range a.APIKeys {
		binary.BigEndian.PutUint16(apiKeyBuf, apiKey.APIKey)
		body = append(body, apiKeyBuf...)
		binary.BigEndian.PutUint16(apiKeyBuf, apiKey.MinVersion)
		body = append(body, apiKeyBuf...)
		binary.BigEndian.PutUint16(apiKeyBuf, apiKey.MaxVersion)
		body = append(body, apiKeyBuf...)
		body = append(body, apiKey.TagBuffer)
	}
	binary.BigEndian.PutUint32(buf32, a.ThrottleTimeMs)
	body = append(body, buf32...)
	body = append(body, a.TagBuffer)
	size := uint32(len(body))
	encodedResponse := make([]byte, 4)
	binary.BigEndian.PutUint32(encodedResponse, size)
	encodedResponse = append(encodedResponse, body...)
	return encodedResponse
}

func getAPIVersionResponse(correlationID uint32) responseEncoder {
	response := &apiVersionResponse{
		baseResponse: baseResponse{
			CorrelationID: correlationID,
			ErrorCode:     0,
		},
		APIKeysArrayLength: uint8(len(avaliableAPI)),
		APIKeys:            avaliableAPI,
		ThrottleTimeMs:     0,
		TagBuffer:          0,
	}
	return response
}

// Topic Version Response handling
type topicPartitionsResponse struct {
	baseResponse
}

func (t *topicPartitionsResponse) encode() []byte {
	buf := make([]byte, 10)
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(buf)-4))
	binary.BigEndian.PutUint32(buf[4:8], t.CorrelationID)
	binary.BigEndian.PutUint16(buf[8:10], t.ErrorCode)
	return buf
}

func getTopicPartitionsResponse(correlationID uint32) responseEncoder {
	response := &topicPartitionsResponse{
		baseResponse: baseResponse{
			CorrelationID: correlationID,
			ErrorCode:     0,
		},
	}
	return response
}
