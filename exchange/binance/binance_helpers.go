package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"

	"strings"
)

func GenerateBinanceSignature(params map[string]string, secretKey string) string {

	var queryString string
	for key, value := range params {
		queryString += key + "=" + url.QueryEscape(value) + "&"
	}

	queryString = strings.TrimSuffix(queryString, "&")
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(queryString))
	signature := hex.EncodeToString(mac.Sum(nil))

	return signature
}
