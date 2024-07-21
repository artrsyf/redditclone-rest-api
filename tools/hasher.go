package tools

import (
	"crypto/sha1"
	"encoding/base64"
)

func GetSHA1Hash(args ...string) string {
	resultString := ""
	for _, arg := range args {
		resultString += arg
	}

	hasher := sha1.New()
	hasher.Write([]byte(resultString))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return hash
}
