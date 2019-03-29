package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

func LookupEnvWithDefault(envName, defaultValue string) string {
	if value, ok := os.LookupEnv(envName); ok {
		return value
	}
	return defaultValue
}

func RandStringBase64(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("Can't generate random strings of length: %d", length)
	}

	randString := make([]byte, length)
	_, err := rand.Read(randString)

	if err != nil {
		return "", fmt.Errorf("Failed to generate random string: %v", err)
	}

	randStringBase64 := base64.StdEncoding.EncodeToString(randString)

	return randStringBase64, nil
}
