package trade

import "testing"

func TestTrade(t *testing.T) {
	trade_cron := TradeCron{}
	trade_cron.Run()
}
