package utils

import (
	"crypto/sha256"
	"fmt"
	"os"
)

func StringHash(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func BytesHash(b []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func FileHash(name string) (string, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return BytesHash(data), nil
}
