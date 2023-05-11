package tests

import (
	"testing"

	"github.com/kryptomind/BidBox-Trades/utils"
)

func TestGetSize(t *testing.T) {
	coins := []string{"SBTCSUSDT_SUMCBL", "SETHSUSDT_SUMCBL", "SEOSSUSDT_SUMCBL", "SXRPSUSDT_SUMCBL"}

	for _, v := range coins {
		size, err := utils.GetSize(v, 13.33)
		if err != nil {
			t.Error(err)
		}
		t.Log(size)
	}

}
