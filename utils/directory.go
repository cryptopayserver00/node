package utils

import (
	"errors"
	"os"
)

func PathExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		if f.IsDir() {
			return true, nil
		}
		return false, errors.New("file same name exists")
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
