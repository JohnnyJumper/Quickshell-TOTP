package cli

import (
	"fmt"
	"totp/internal/daemon"
	"totp/internal/ipc"
)

type command string

const (
	DaemonCommand       command = "daemon"
	StatusCommand       command = "status"
	ListCommand         command = "list"
	ImportTextCommand   command = "import-text"
	ImportImageCommand  command = "import-image"
	CopyCommand         command = "copy"
	HelpCommand         command = "help"
	HelpShortcutCommand command = "-h"
)

func Run(args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}

	switch cmd := command(args[0]); cmd {
	case DaemonCommand:
		return runDaemon()

	case StatusCommand:
		return runStatus()

	case ImportImageCommand:
		if len(args) != 2 {
			return fmt.Errorf("usage: totp import-image <path>")
		}

		return runQRCodeImport(args[1])

	case ListCommand:
		return runList()

	case ImportTextCommand:
		if len(args) != 2 {
			return fmt.Errorf("usage: totp import-text <otpauth://totp/...>")
		}
		return runTextImport(args[1])

	case CopyCommand:
		if len(args) != 2 {
			return fmt.Errorf("usage: totp copy <account-id>")
		}
		return runCopyCode(args[1])

	case HelpCommand, HelpShortcutCommand:
		printUsage()
		return nil

	default:
		return fmt.Errorf("Unknown command: %s", cmd)
	}
}

func runList() error {
	client := ipc.NewDefaultClient()

	response, err := client.Send(ipc.Request{
		Type: ipc.ListAccountsRequest,
	})
	if err != nil {
		return err
	}

	if len(response.Accounts) == 0 {
		fmt.Println("No accounts imported.")
		return nil
	}

	for _, account := range response.Accounts {
		fmt.Printf("%s\t%s\t%s\n", account.ID, account.Issuer, account.Label)
	}

	return nil
}

func runQRCodeImport(path string) error {
	payload, err := decodeQRCode(path)
	if err != nil {
		return fmt.Errorf("Couldn't parse image:\n\t%w", err)
	}

	client := ipc.NewDefaultClient()
	response, err := client.Send(ipc.Request{
		Type:    ipc.QRCodeImportRequest,
		Payload: payload,
	})

	if err != nil {
		return err
	}

	fmt.Println(response.Message)
	for _, account := range response.Accounts {
		fmt.Printf("%s\t%s\t%s\t%s\t%d\t%d\n",
			account.ID,
			account.Issuer,
			account.Label,
			account.Algorithm,
			account.Digits,
			account.Period,
		)
	}
	return nil

}

func runTextImport(payload string) error {
	client := ipc.NewDefaultClient()
	response, err := client.Send(ipc.Request{
		Type:    ipc.TextImportRequest,
		Payload: payload,
	})
	if err != nil {
		return err
	}
	fmt.Println(response.Message)
	return nil
}

func runCopyCode(accountID string) error {
	client := ipc.NewDefaultClient()
	response, err := client.Send(ipc.Request{
		Type:      ipc.CopyCodeRequest,
		AccountID: accountID,
	})
	if err != nil {
		return err
	}
	fmt.Println(response.Message)
	return nil
}

func runStatus() error {
	client := ipc.NewDefaultClient()

	response, err := client.Send(ipc.Request{
		Type: ipc.StatusRequest,
	})
	if err != nil {
		return err
	}

	fmt.Println(response.Message)
	return nil
}

func runDaemon() error {
	if err := daemon.Run(); err != nil {
		return fmt.Errorf("Failed to launch daemon: %w", err)
	}
	return nil
}

func printUsage() {
	fmt.Println("Hello!")
}
