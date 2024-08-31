package encrypt

import (
	"encoding/base64"
	"fmt"
)

func Base64Encode(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

func Base64Decode(text string) (string, error) {
	result, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", fmt.Errorf("invalid base64 encoding: %w", err)
	}

	return string(result), nil
}
