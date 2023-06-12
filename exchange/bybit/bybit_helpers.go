package bybit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"github.com/kryptomind/BidBox-Trades/shared"
	// "github.com/kryptomin/shared"
)

func GenerateBybitSignature(apiKey, apiSecret string, recvWindow, timestamp int64, queryString string) string {
	dataToSign := fmt.Sprintf("%d%s%d%s", timestamp, apiKey, recvWindow, queryString)
	hmacKey := []byte(apiSecret)
	hmacHash := hmac.New(sha256.New, hmacKey)
	hmacHash.Write([]byte(dataToSign))
	signature := hex.EncodeToString(hmacHash.Sum(nil))
	return signature
}

func GenerateBybitV2Signature(queryString, apiSecret string) string {
	hmacHash := hmac.New(sha256.New, []byte(apiSecret))
	hmacHash.Write([]byte(queryString))
	signature := hmacHash.Sum(nil)

	return hex.EncodeToString(signature)
}

func TransformContractAccountBalance(accountData AccountBalanceResponse) shared.AccountData {
	data := accountData.Result.List[0]

	totalAvailable := 0.0
	totalUnrealizedPL := 0.0
	totalEquity := 0.0

	for _, coinData := range data.Coin {
		available, err := strconv.ParseFloat(coinData.WalletBalance, 64)
		if err != nil {
			fmt.Println("Error parsing available balance:", err)
		}
		totalAvailable += available
		equity, err := strconv.ParseFloat(coinData.Equity, 64)
		if err != nil {
			fmt.Println("Error parsing equity:", err)
		}
		totalEquity += equity
		unrealizedPL, err := strconv.ParseFloat(coinData.UnrealisedPnl, 64)
		if err != nil {
			fmt.Println("Error parsing unrealized PL:", err)
		}
		totalUnrealizedPL += unrealizedPL

	}

	return shared.AccountData{
		Available:     totalAvailable,
		MarginBalance: totalEquity,
		UnrealizedPL:  totalUnrealizedPL,
	}

}

func TransformUnifiedAccountBalance(accountData AccountBalanceResponse) shared.AccountData {
	data := accountData.Result.List[0]

	available, err := strconv.ParseFloat(data.TotalAvailableBalance, 64)
	if err != nil {
		fmt.Println("Error parsing available balance:", err)
	}

	equity, err := strconv.ParseFloat(data.TotalEquity, 64)
	if err != nil {
		fmt.Println("Error parsing equity:", err)
	}

	unrealizedPL, err := strconv.ParseFloat(data.TotalPerpUPL, 64)
	if err != nil {
		fmt.Println("Error parsing unrealized PL:", err)
	}

	return shared.AccountData{
		Available:     available,
		MarginBalance: equity,
		UnrealizedPL:  unrealizedPL,
	}

}

func TransformAccountPositionsResponse(positionData PositionsResponse) []shared.PositionsData {
	positionsList := positionData.Result.List

	var formattedPositions []shared.PositionsData

	for _, position := range positionsList {

		positionSide := "long"
		holdMode := "single_hold"

		if position.Side == "short" {
			positionSide = "short"
		}

		if position.PositionIdx == 1 {
			holdMode = "double_hold"
		}

		pos := shared.PositionsData{
			MarginCoin:       position.Symbol,
			Symbol:           position.Symbol,
			HoldSide:         positionSide,
			Margin:           position.PositionMM,
			Available:        position.PositionValue,
			Total:            position.Size,
			MarginMode:       "fixed",
			HoldMode:         holdMode,
			LiquidationPrice: position.LiqPrice,
			MarketPrice:      position.MarkPrice,
			CreationTime:     position.CreatedTime,
			UnrealizedPL:     position.UnrealisedPnl,
			Leverage:         position.Leverage,
		}

		formattedPositions = append(formattedPositions, pos)
	}

	return formattedPositions
}

func HasRequiredPermissions(permissions []string) bool {
	requiredStrings := []string{"Order", "Position", "ExchangeHistory"}

	for _, str := range requiredStrings {
		found := false
		for _, permission := range permissions {
			if permission == str {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
