package daemon

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/term"

	"totp/internal/ipc"
	"totp/internal/store"
)

func Run() error {
	if err := store.InitMasterKey(func(prompt string) (string, error) {
		fmt.Print(prompt)
		passphrase, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		return string(passphrase), err
	}); err != nil {
		return fmt.Errorf("initialize master key: %w", err)
	}

	socketPath := ipc.DefaultSocketPath()

	if err := prepareSocket(socketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("listen on socket: %w", err)
	}
	defer listener.Close()
	defer os.Remove(socketPath)

	if err := os.Chmod(socketPath, 0600); err != nil {
		return fmt.Errorf("set socket permissions: %w", err)
	}

	log.Printf("totp daemon running at %s", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go handleConnection(conn)
	}
}

func prepareSocket(socketPath string) error {
	socketDir := filepath.Dir(socketPath)

	if err := os.MkdirAll(socketDir, 0700); err != nil {
		return fmt.Errorf("create socket directory: %w", err)
	}

	if _, err := os.Stat(socketPath); err == nil {
		if daemonIsRunning(socketPath) {
			return fmt.Errorf("daemon already running at %s", socketPath)
		}

		if err := os.Remove(socketPath); err != nil {
			return fmt.Errorf("remove stale socket: %w", err)
		}
	}

	return nil
}

func daemonIsRunning(socketPath string) bool {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return false
	}

	_ = conn.Close()
	return true
}
