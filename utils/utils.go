package utils

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func GetCurrentDir() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", errors.Cause(err)
	}
	return dir, nil
}
