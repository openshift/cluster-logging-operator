package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

func GetInt64(value int64) *int64 {
	i := value
	return &i
}
func GetInt32(value int32) *int32 {
	i := value
	return &i
}

func ToJson(obj interface{}) (string, error) {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

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

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandStringBytes(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("Can't generate random strings of length: %d", length)
	}

	randString := make([]byte, length)
	_, err := rand.Read(randString)
	if err != nil {
		return "", fmt.Errorf("Failed to generate random string: %v", err)
	}

	for i, b := range randString {
		randString[i] = letters[b%byte(len(letters))]
	}
	return string(randString), nil
}
