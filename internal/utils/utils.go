// Package utils contains general-purpose helper functions that are not tied to specific business or protocol logic.
package utils

import (
	"encoding/binary"
	"fmt"
	"io"
)

type UUID [16]byte

func (u UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		u[:4],
		u[4:6],
		u[6:8],
		u[8:10],
		u[10:16],
	)
}

func CloseResource(r io.Closer) {
	if err := r.Close(); err != nil {
		fmt.Println("Error closing resource:", err.Error())
	}
}

func AppendBuf16(buf *[]byte, value uint16, result *[]byte) {
	binary.BigEndian.PutUint16((*buf), value)
	*result = append(*result, *buf...)
}

func AppendBuf32(buf *[]byte, value uint32, result *[]byte) {
	binary.BigEndian.PutUint32(*buf, value)
	*result = append(*result, *buf...)
}
