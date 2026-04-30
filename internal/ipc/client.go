package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"
)

type Client struct {
	socketPath string
	timeout    time.Duration
}

func New(socketPath string) Client {
	return Client{
		socketPath: socketPath,
		timeout:    2 * time.Second,
	}
}

func NewDefaultClient() Client {
	return New(DefaultSocketPath())
}

func (client Client) Send(request Request) (Response, error) {
	conn, err := net.Dial("unix", client.socketPath)
	if err != nil {
		return Response{}, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(client.timeout)
	_ = conn.SetDeadline(deadline)

	if err := json.NewEncoder(conn).Encode(request); err != nil {
		return Response{}, fmt.Errorf("failed to send request: %w", err)
	}

	var response Response
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return Response{}, fmt.Errorf("failed to receive response: %w", err)
	}

	if !response.OK {
		return response, errors.New(response.Error)
	}

	return response, nil
}
