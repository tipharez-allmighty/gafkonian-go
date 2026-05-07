package handler

import (
	"encoding/binary"

	customerr "github.com/gafkonian-go/internal/custom_error"
)

type RequestHeader struct {
	RequestAPIKey     uint16
	RequestAPIVersion uint16
	CorrelationID     uint32
}

const (
	apiVersion      uint16 = 18
	topicPartitions uint16 = 75
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
}
var APIKeyMap = make(map[uint16]*apiVersionKey)

func InitAvaliableAPIKeys() {
	for i := range avaliableAPI {
		APIKeyMap[avaliableAPI[i].APIKey] = &avaliableAPI[i]
	}
}

func (h *RequestHeader) Validate() error {
	if _, ok := APIKeyMap[h.RequestAPIKey]; !ok {
		return customerr.RaiseError(customerr.UnsupportedAPIKeyError, h.RequestAPIKey)
	}
	if h.RequestAPIVersion > 4 {
		return customerr.RaiseError(customerr.UnsupportedAPIVersionError, h.RequestAPIVersion)
	}
	return nil
}

func ParseHeader(data []byte) (*RequestHeader, error) {
	if len(data) < 8 {
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
