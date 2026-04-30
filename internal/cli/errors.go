package cli

import (
	"errors"
	"fmt"
)

var (
	ErrUsage             = errors.New("Incorrect usage use -h or --help for help")
	ErrImagePathRequired = errors.New("Image path is required.")
	ErrZbarImgRequired   = errors.New("zbarimg not found; please install it on your machine and ensure it is accessible via $PATH")
)

func ErrImageFileDoesNotExist(path string) error {
	return fmt.Errorf("Image file does not exists: %s", path)
}

func ErrReadingFile(err error) error {
	return fmt.Errorf("Error reading file: %w", err)
}

func ErrImagePathIsDir(path string) error {
	return fmt.Errorf("Provided path to a directory and not image: %s", path)
}

func ErrDecodingQRCode(message string) error {
	return fmt.Errorf("Decoding QR image with zbarimg errored: %s", message)
}

func ErrNoQRCodeFound(path string) error {
	message := fmt.Sprintf("No QR code found in image: %s", path)
	return ErrDecodingQRCode(message)
}
