package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"totp/internal/daemon/importer"
	"totp/internal/ipc"
	"totp/internal/store"
	"totp/internal/totp"
)

func handleStatus(_ ipc.Request) ipc.Response {
	return ipc.Success("daemon is alive")
}

func handleQRCodeImport(request ipc.Request) ipc.Response {
	if request.Payload == "" {
		return ipc.Failure(fmt.Errorf("QR code import payload is required"))
	}

	importedAccounts, err := importer.ParsePayload(request.Payload)
	if err != nil {
		return ipc.Failure(err)
	}

	fileStore := store.NewFileStore("")

	existingAccounts, err := fileStore.Load()
	if err != nil {
		return ipc.Failure(err)
	}

	mergedAccounts := mergeAccounts(
		existingAccounts,
		toStoreAccounts(importedAccounts),
	)

	if err := fileStore.Save(mergedAccounts); err != nil {
		return ipc.Failure(err)
	}

	return ipc.Success(fmt.Sprintf("imported %d account(s)", len(importedAccounts)))
}

func handleListAccounts(_ ipc.Request) ipc.Response {
	fileStore := store.NewFileStore("")

	accounts, err := fileStore.Load()
	if err != nil {
		return ipc.Failure(err)
	}

	return ipc.AccountsListed(toIPCAccounts(accounts))
}

func toIPCAccounts(accounts []store.Account) []ipc.Account {
	result := make([]ipc.Account, 0, len(accounts))

	for _, account := range accounts {
		result = append(result, ipc.Account{
			ID:        account.ID,
			Issuer:    account.Issuer,
			Label:     account.Label,
			Algorithm: account.Algorithm,
			Digits:    account.Digits,
			Period:    account.Period,
		})
	}

	return result
}

func toStoreAccounts(accounts []importer.Account) []store.Account {
	result := make([]store.Account, 0, len(accounts))

	for _, account := range accounts {
		result = append(result, store.Account{
			ID:        account.ID,
			Issuer:    account.Issuer,
			Label:     account.Label,
			Secret:    account.Secret,
			Algorithm: account.Algorithm,
			Digits:    account.Digits,
			Period:    account.Period,
		})
	}

	return result
}

func handleTextImport(request ipc.Request) ipc.Response {
	if request.Payload == "" {
		return ipc.Failure(fmt.Errorf("text import payload is required"))
	}

	importedAccounts, err := importer.ParsePayload(request.Payload)
	if err != nil {
		return ipc.Failure(err)
	}

	fileStore := store.NewFileStore("")

	existingAccounts, err := fileStore.Load()
	if err != nil {
		return ipc.Failure(err)
	}

	mergedAccounts := mergeAccounts(existingAccounts, toStoreAccounts(importedAccounts))

	if err := fileStore.Save(mergedAccounts); err != nil {
		return ipc.Failure(err)
	}

	return ipc.Success(fmt.Sprintf("imported %d account(s)", len(importedAccounts)))
}

func handleCopyCode(request ipc.Request) ipc.Response {
	if request.AccountID == "" {
		return ipc.Failure(fmt.Errorf("account ID is required"))
	}

	fileStore := store.NewFileStore("")

	accounts, err := fileStore.Load()
	if err != nil {
		return ipc.Failure(err)
	}

	var account *store.Account
	for i := range accounts {
		if accounts[i].ID == request.AccountID {
			account = &accounts[i]
			break
		}
	}
	if account == nil {
		return ipc.Failure(fmt.Errorf("account not found: %s", request.AccountID))
	}

	code, err := totp.Generate(account.Secret, account.Algorithm, account.Digits, account.Period, time.Now())
	if err != nil {
		return ipc.Failure(fmt.Errorf("generate TOTP code: %w", err))
	}

	if err := clipboardCopy(code.Value); err != nil {
		return ipc.Failure(fmt.Errorf("clipboard copy: %w", err))
	}

	return ipc.Success(fmt.Sprintf("copied code for %s (%ds remaining)", account.Issuer, code.Remaining))
}

func clipboardCopy(text string) error {
	var cmd *exec.Cmd
	switch {
	case os.Getenv("WAYLAND_DISPLAY") != "":
		cmd = exec.Command("wl-copy")
	case os.Getenv("DISPLAY") != "":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	default:
		cmd = exec.Command("wl-copy")
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func mergeAccounts(existing []store.Account, imported []store.Account) []store.Account {
	accountsByID := make(map[string]store.Account, len(existing)+len(imported))

	for _, account := range existing {
		accountsByID[account.ID] = account
	}

	for _, account := range imported {
		accountsByID[account.ID] = account
	}

	result := make([]store.Account, 0, len(accountsByID))
	for _, account := range accountsByID {
		result = append(result, account)
	}

	sort.Slice(result, func(i, j int) bool {
		left := result[i]
		right := result[j]

		if left.Issuer != right.Issuer {
			return left.Issuer < right.Issuer
		}

		if left.Label != right.Label {
			return left.Label < right.Label
		}

		return left.ID < right.ID
	})

	return result
}
