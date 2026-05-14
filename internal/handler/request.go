package handler

import (
	"encoding/binary"

	customerr "github.com/gafkonian-go/internal/custom_error"
	"github.com/gafkonian-go/internal/metadata"
)

type RequestHeader struct {
	RequestAPIKey     uint16
	RequestAPIVersion uint16
	CorrelationID     uint32
}

const (
	apiVersion      uint16 = 18
	topicPartitions uint16 = 75
	fetch           uint16 = 16
	produce         uint16 = 0
)

const (
	headerSize int = 8
	bodyOffset int = 20
)

type apiVersionKey struct {
	APIKey     uint16
	MinVersion uint16
	MaxVersion uint16
	TagBuffer  uint8
}

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
	{
		APIKey:     fetch,
		MinVersion: 0,
		MaxVersion: 16,
		TagBuffer:  0,
	},
	{
		APIKey:     produce,
		MinVersion: 0,
		MaxVersion: 11,
		TagBuffer:  0,
	},
}
var APIKeyMap = make(map[uint16]*apiVersionKey)

func InitAvaliableAPIKeys() {
	for i := range avaliableAPI {
		APIKeyMap[avaliableAPI[i].APIKey] = &avaliableAPI[i]
	}
}

func (h *RequestHeader) Validate() error {
	apiVersionKey, ok := APIKeyMap[h.RequestAPIKey]
	if !ok {
		return customerr.RaiseError(customerr.UnsupportedAPIKeyError, h.RequestAPIKey)
	}
	if h.RequestAPIVersion > apiVersionKey.MaxVersion {
		return customerr.RaiseError(customerr.UnsupportedAPIVersionError, h.RequestAPIVersion)
	}
	return nil
}

func ParseHeader(data []byte) (*RequestHeader, error) {
	if len(data) < headerSize {
		return nil, customerr.RaiseError(customerr.InsufficientHeaderError, len(data))
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

type TopicRequest struct {
	TopicName string
	TagBuffer uint8
}
type DescribeTopicRequestBody struct {
	Topics    []TopicRequest
	TagBuffer uint8
}

func ParseDescribeTopicBody(data []byte) (*DescribeTopicRequestBody, error) {
	body := data[bodyOffset:]

	arraylength, n := binary.Uvarint(body)
	if n <= 0 {
		return nil, customerr.RaiseError(customerr.InsufficientBodyError)
	}
	offset := n
	arraySize := int(arraylength - 1)
	if arraySize < 1 {
		return nil, customerr.RaiseError(customerr.EmptyTopicsBodyError)
	}
	topicRequests := make([]TopicRequest, 0, arraySize)

	for range arraySize {
		strLenPlusOne, n := binary.Uvarint(body[offset:])
		offset += n
		strLen := int(strLenPlusOne - 1)

		topicName := string(body[offset : offset+strLen])
		offset += strLen
		offset++

		topicRequests = append(topicRequests, TopicRequest{
			TopicName: topicName,
		})
	}
	return &DescribeTopicRequestBody{
		Topics: topicRequests,
	}, nil
}

type ProduceBody struct {
	TopicName   string
	PartitionID uint32
	Records     []metadata.Record
}

func ParseProduceBody(data []byte) (*ProduceBody, error) {
	if len(data) < bodyOffset {
		return nil, customerr.RaiseError(customerr.InsufficientBodyError)
	}
	body := data[bodyOffset:]
	offset := 0

	topicLen := int(binary.BigEndian.Uint32(body[offset : offset+4]))
	offset += 4
	topicName := string(body[offset : offset+topicLen])
	offset += topicLen

	partitionID := binary.BigEndian.Uint32(body[offset : offset+4])
	offset += 4

	recordCount := int(binary.BigEndian.Uint32(body[offset : offset+4]))
	offset += 4

	records := make([]metadata.Record, 0, recordCount)

	for range recordCount {
		off := binary.BigEndian.Uint64(body[offset : offset+8])
		offset += 8

		ts := binary.BigEndian.Uint64(body[offset : offset+8])
		offset += 8

		keyLen := int(binary.BigEndian.Uint32(body[offset : offset+4]))
		offset += 4
		key := body[offset : offset+keyLen]
		offset += keyLen

		valLen := int(binary.BigEndian.Uint32(body[offset : offset+4]))
		offset += 4
		value := body[offset : offset+valLen]
		offset += valLen

		records = append(records, metadata.Record{
			Offset:    off,
			Timestamp: ts,
			Key:       key,
			Value:     value,
		})
	}

	return &ProduceBody{
		TopicName:   topicName,
		PartitionID: partitionID,
		Records:     records,
	}, nil
}
