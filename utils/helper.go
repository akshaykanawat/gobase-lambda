package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
	"unsafe"

	"gobase-lambda/log"
)

var IST, _ = time.LoadLocation("Asia/Kolkata")

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func GetStringValue(data map[string]interface{}, key string) (string, bool) {
	logger := log.GetDefaultLogger()
	valIn, ok := data[key]
	if ok {
		logger.Info(fmt.Sprintf("key found - %v, value - %v", key, valIn), nil)
		return valIn.(string), ok
	}
	logger.Info(fmt.Sprintf("key not found - %v", key), nil)
	return "", ok
}

func GetMapValue(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	valIn, ok := data[key]
	logger := log.GetDefaultLogger()
	if ok {
		logger.Info(fmt.Sprintf("key found - %v", key), valIn)
		return valIn.(map[string]interface{}), ok
	}
	logger.Info(fmt.Sprintf("key not found - %v", key), nil)
	return nil, ok
}

func GetString(obj interface{}) (*string, error) {
	blob, err := json.Marshal(obj)
	logger := log.GetDefaultLogger()
	if err != nil {
		logger.Error("Error marshalling object", err)
		logger.Error("Marshall input object", obj)
		return nil, err
	}
	body := string(blob)
	return &body, nil
}

func GetHash(val string) string {
	uriHashByte := sha256.Sum256([]byte(val))
	return hex.EncodeToString([]byte(uriHashByte[:]))
}

func GetRandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
