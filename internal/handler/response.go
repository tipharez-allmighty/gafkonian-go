package handler

import (
	"encoding/binary"
)

type baseResponse struct {
	MessageSize   uint32
	CorrelationID uint32
	ErrorCode     uint16
}
type errorResponse struct {
	baseResponse
}

type responseEncoder interface {
	encode() []byte
}

func (e *errorResponse) encode() []byte {
	buf := make([]byte, 10)
	binary.BigEndian.PutUint32(buf[0:4], e.MessageSize)
	binary.BigEndian.PutUint32(buf[4:8], e.CorrelationID)
	binary.BigEndian.PutUint16(buf[8:10], e.ErrorCode)
	return buf
}

func getErrorResponse(errorCode uint16, correlationID uint32) responseEncoder {
	response := &errorResponse{
		baseResponse: baseResponse{
			MessageSize:   6,
			CorrelationID: correlationID,
			ErrorCode:     errorCode,
		},
	}
	return response
}

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
	buf := make([]byte, 10, 23)
	binary.BigEndian.PutUint32(buf[0:4], a.MessageSize)
	binary.BigEndian.PutUint32(buf[4:8], a.CorrelationID)
	binary.BigEndian.PutUint16(buf[8:10], uint16(0))
	buf = append(buf, uint8(len(a.APIKeys)+1))

	apiKeyBuf := make([]byte, 2)
	for _, apiKey := range a.APIKeys {
		binary.BigEndian.PutUint16(apiKeyBuf, apiKey.APIKey)
		buf = append(buf, apiKeyBuf...)
		binary.BigEndian.PutUint16(apiKeyBuf, apiKey.MinVersion)
		buf = append(buf, apiKeyBuf...)
		binary.BigEndian.PutUint16(apiKeyBuf, apiKey.MaxVersion)
		buf = append(buf, apiKeyBuf...)
		buf = append(buf, apiKey.TagBuffer)
	}
	throttleTimeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(throttleTimeBuf, a.ThrottleTimeMs)
	buf = append(buf, throttleTimeBuf...)
	buf = append(buf, a.TagBuffer)
	return buf
}

func getAPIVersionResponse(correlationID uint32) responseEncoder {
	response := &apiVersionResponse{
		baseResponse: baseResponse{
			MessageSize:   19,
			CorrelationID: correlationID,
			ErrorCode:     0,
		},
		APIKeysArrayLength: 1,
		APIKeys: []apiVersionKey{
			{
				APIKey:     18,
				MinVersion: 0,
				MaxVersion: 4,
				TagBuffer:  0,
			},
		},
		ThrottleTimeMs: 0,
		TagBuffer:      0,
	}
	return response
}
