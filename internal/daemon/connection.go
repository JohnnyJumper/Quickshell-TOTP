package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"totp/internal/ipc"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var request ipc.Request
	if err := json.NewDecoder(conn).Decode(&request); err != nil {
		log.Printf("decode request: %v", err)
		_ = ipc.WriteResponse(conn, ipc.Failure(err))
		return
	}

	response := handleRequest(request)

	if !response.OK {
		log.Printf("[%s] error: %s", request.Type, response.Error)
	}

	if err := ipc.WriteResponse(conn, response); err != nil {
		log.Printf("write response: %v", err)
	}
}

func handleRequest(request ipc.Request) ipc.Response {
	switch request.Type {
	case ipc.StatusRequest:
		return handleStatus(request)
	case ipc.QRCodeImportRequest:
		return handleQRCodeImport(request)
	case ipc.ListAccountsRequest:
		return handleListAccounts(request)
	case ipc.TextImportRequest:
		return handleTextImport(request)
	case ipc.CopyCodeRequest:
		return handleCopyCode(request)

	default:
		return ipc.Failure(fmt.Errorf("Unknown request type: %s", request.Type))
	}
}
