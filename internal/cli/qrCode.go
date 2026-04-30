package cli

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// Takes a path to a QR image and returns its payload
func decodeQRCode(path string) (string, error) {
	if path == "" {
		return "", ErrImagePathRequired
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrImageFileDoesNotExist(path)
		}

		return "", ErrReadingFile(err)
	}

	if info.IsDir() {
		return "", ErrImagePathIsDir(path)
	}

	zbarimgPath, err := exec.LookPath("zbarimg")
	if err != nil {
		return "", ErrZbarImgRequired
	}

	cmd := exec.Command(zbarimgPath, "--quiet", "--raw", path)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}

		return "", ErrDecodingQRCode(message)
	}

	payload := strings.TrimSpace(string(output))
	if payload == "" {
		return "", ErrNoQRCodeFound(path)
	}

	return payload, nil
}
