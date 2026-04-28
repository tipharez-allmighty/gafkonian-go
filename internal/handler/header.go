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

func (h *RequestHeader) Validate() error {
	if h.RequestAPIKey != 18 {
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
