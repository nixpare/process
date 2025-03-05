//go:build windows && cgo
package process

/*
#include <windows.h>
#include <winsafer.h>

void* getLowerPrivilegeToken() {
	SAFER_LEVEL_HANDLE safer;
	HANDLE token;

	if (!SaferCreateLevel(SAFER_SCOPEID_USER, SAFER_LEVELID_NORMALUSER,
						SAFER_LEVEL_OPEN, &safer, NULL)) {
		return NULL;
	}

	if (!SaferComputeTokenFromLevel(safer, NULL, &token,
						SAFER_TOKEN_NULL_IF_EQUAL, NULL)) {
		return NULL;
	}

	// Set medium integrity
	TOKEN_MANDATORY_LABEL TIL;
	unsigned char medium[12] = {
		0x01, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x10,
		0x00, 0x20, 0x00, 0x00
	};
	TIL.Label.Sid = (PSID)medium;
	TIL.Label.Attributes = SE_GROUP_INTEGRITY;

	if (!SetTokenInformation(token, TokenIntegrityLevel, &TIL,
							sizeof(TOKEN_MANDATORY_LABEL))) {
		return NULL;
	}

	return token;
}
*/
import "C"
import (
	"fmt"
	"syscall"
)

func GetLowerPrivilegeToken() (syscall.Token, error) {
	token := syscall.Token(C.getLowerPrivilegeToken())
	if token == 0 {
		return 0, fmt.Errorf("GetLowerPrivilegeToken error: %w", syscall.GetLastError())
	}

	return token, nil
}