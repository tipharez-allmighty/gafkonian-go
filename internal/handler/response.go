package handler

import (
	"encoding/binary"

	"github.com/gafkonian-go/internal/metadata"
	"github.com/gafkonian-go/internal/utils"
)

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
	for _, apiKey := range a.APIKeys {
		binary.BigEndian.PutUint16(buf16, apiKey.APIKey)
		body = append(body, buf16...)
		binary.BigEndian.PutUint16(buf16, apiKey.MinVersion)
		body = append(body, buf16...)
		binary.BigEndian.PutUint16(buf16, apiKey.MaxVersion)
		body = append(body, buf16...)
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
	CorrelationID     uint32
	TagBufferHeader   uint8
	ThrottleTimeMs    uint32
	Topics            []Topic
	NextCursor        int8
	TagBufferResponse uint8
}

type Topic struct {
	ErrorCode                 uint16
	TopicName                 string
	TopicID                   utils.UUID
	IsInternal                bool
	Partitions                []byte
	TopicAuthorizedOperations uint32
	TagBuffer                 uint8
}

type Partition struct {
	ErrorCode uint16
	metadata.PartitionMetadata
	EligibleLeaderReplicas []uint32
	LastKnownElr           []uint32
	OfflineReplicas        []uint32
}

func (t *topicPartitionsResponse) encode() []byte {
	body := make([]byte, 0)
	buf16 := make([]byte, 2)
	buf32 := make([]byte, 4)
	binary.BigEndian.PutUint32(buf32, t.CorrelationID)
	body = append(body, buf32...)
	binary.BigEndian.PutUint16(buf16, 0)
	body = append(body, buf16...)
	binary.BigEndian.PutUint32(buf32, t.ThrottleTimeMs)
	body = append(body, buf32...)
	body = append(body, uint8(len(t.Topics)+1))
	for _, topic := range t.Topics {
		binary.BigEndian.PutUint16(buf16, topic.ErrorCode)
		body = append(body, buf16...)
		topicNameSize := uint8(len(topic.TopicName) + 1)
		body = append(body, topicNameSize)
		body = append(body, topic.TopicName...)
		body = append(body, topic.TopicID[:]...)
		isInternal := 0
		if topic.IsInternal {
			isInternal = 1
		}
		body = append(body, byte(isInternal))
		body = append(body, topic.Partitions...)
		binary.BigEndian.PutUint32(buf32, topic.TopicAuthorizedOperations)
		body = append(body, buf32...)
		body = append(body, 0)
	}
	body = append(body, byte(t.NextCursor))
	body = append(body, 0)
	size := uint32(len(body))
	encodedResponse := make([]byte, 4)
	binary.BigEndian.PutUint32(encodedResponse, size)
	encodedResponse = append(encodedResponse, body...)
	return encodedResponse
}

func getTopicPartitionsResponse(correlationID uint32) responseEncoder {
	response := &topicPartitionsResponse{
		CorrelationID:   correlationID,
		TagBufferHeader: 0,
		ThrottleTimeMs:  0,
		Topics: []Topic{
			{
				ErrorCode:                 3,
				TopicName:                 "foo",
				TopicID:                   utils.UUID{},
				IsInternal:                false,
				Partitions:                []byte{},
				TopicAuthorizedOperations: 0,
				TagBuffer:                 0,
			},
		},
		NextCursor:        -1,
		TagBufferResponse: 0,
	}
	return response
}
