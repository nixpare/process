//go:build windows && !cgo
package process

import (
	"fmt"
	"syscall"
)

func GetLowerPrivilegeToken() (syscall.Token, error) {
	return token, errors.New("You have to enable CGO")
}