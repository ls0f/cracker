package cracker

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
)

func GenHMACSHA1(key, raw string) string {
	k := []byte(key)
	mac := hmac.New(sha1.New, k)
	mac.Write([]byte(raw))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func VerifyHMACSHA1(key, raw, sign string) bool {
	return GenHMACSHA1(key, raw) == sign
}
