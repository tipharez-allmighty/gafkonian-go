// Package utils contains general-purpose helper functions that are not tied to specific business or protocol logic.
package utils

import (
	"fmt"
	"io"
)

func CloseResource(r io.Closer) {
	if err := r.Close(); err != nil {
		fmt.Println("Error closing resource:", err.Error())
	}
}
