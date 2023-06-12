package bybit

import "time"

type APIKeyInfoResponse struct {
	RetCode          int          `json:"ret_code"`
	RetMsg           string       `json:"ret_msg"`
	Result           []APIKeyInfo `json:"result"`
	ExtCode          string       `json:"ext_code"`
	ExtInfo          string       `json:"ext_info"`
	TimeNow          string       `json:"time_now"`
	RateLimitStatus  int          `json:"rate_limit_status"`
	RateLimit        int          `json:"rate_limit"`
	RateLimitResetMs int64        `json:"rate_limit_reset_ms"`
}

type APIKeyInfo struct {
	Note          string    `json:"note"`
	APIKey        string    `json:"api_key"`
	ReadOnly      bool      `json:"read_only"`
	Permissions   []string  `json:"permissions"`
	IPs           []string  `json:"ips"`
	Type          string    `json:"type"`
	ExpiredAt     time.Time `json:"expired_at"`
	CreatedAt     time.Time `json:"created_at"`
	UserID        int       `json:"user_id"`
	InviterID     int       `json:"inviter_id"`
	VIPLevel      string    `json:"vip_level"`
	MktMakerLevel string    `json:"mkt_maker_level"`
	AffiliateID   int       `json:"affiliate_id"`
}

type AccountInfo struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		MarginMode          string `json:"marginMode"`
		UpdatedTime         string `json:"updatedTime"`
		UnifiedMarginStatus int    `json:"unifiedMarginStatus"`
		DcpStatus           string `json:"dcpStatus"`
		TimeWindow          int    `json:"timeWindow"`
		SmpGroup            int    `json:"smpGroup"`
	} `json:"result"`
}

type ErrorResponse struct {
	RetCode    int                    `json:"retCode"`
	RetMsg     string                 `json:"retMsg"`
	Result     map[string]interface{} `json:"result"`
	RetExtInfo map[string]interface{} `json:"retExtInfo"`
	Time       int64                  `json:"time"`
}

type AccountBalanceSuccessResponse struct {
	RetCode    int                    `json:"retCode"`
	RetMsg     string                 `json:"retMsg"`
	Result     interface{}            `json:"result"`
	RetExtInfo map[string]interface{} `json:"retExtInfo"`
	Time       int64                  `json:"time"`
}

type AccountBalanceResponse struct {
	RetCode    int                    `json:"retCode"`
	RetMsg     string                 `json:"retMsg"`
	Result     ResultData             `json:"result"`
	RetExtInfo map[string]interface{} `json:"retExtInfo"`
	Time       int64
}

type ResultData struct {
	List []AccountData `json:"list"`
}

type AccountData struct {
	TotalEquity            string     `json:"totalEquity"`
	AccountIMRate          string     `json:"accountIMRate"`
	TotalMarginBalance     string     `json:"totalMarginBalance"`
	TotalInitialMargin     string     `json:"totalInitialMargin"`
	AccountType            string     `json:"accountType"`
	TotalAvailableBalance  string     `json:"totalAvailableBalance"`
	AccountMMRate          string     `json:"accountMMRate"`
	TotalPerpUPL           string     `json:"totalPerpUPL"`
	TotalWalletBalance     string     `json:"totalWalletBalance"`
	AccountLTV             string     `json:"accountLTV"`
	TotalMaintenanceMargin string     `json:"totalMaintenanceMargin"`
	Coin                   []CoinData `json:"coin"`
}

type CoinData struct {
	AvailableToBorrow   string `json:"availableToBorrow"`
	Bonus               string `json:"bonus"`
	AccruedInterest     string `json:"accruedInterest"`
	AvailableToWithdraw string `json:"availableToWithdraw"`
	TotalOrderIM        string `json:"totalOrderIM"`
	Equity              string `json:"equity"`
	TotalPositionMM     string `json:"totalPositionMM"`
	USDValue            string `json:"usdValue"`
	UnrealisedPnl       string `json:"unrealisedPnl"`
	BorrowAmount        string `json:"borrowAmount"`
	TotalPositionIM     string `json:"totalPositionIM"`
	WalletBalance       string `json:"walletBalance"`
	CumRealisedPnl      string `json:"cumRealisedPnl"`
	Coin                string `json:"coin"`
}

type Position struct {
	Symbol           string `json:"symbol"`
	Leverage         string `json:"leverage"`
	AutoAddMargin    int    `json:"autoAddMargin"`
	AvgPrice         string `json:"avgPrice"`
	LiqPrice         string `json:"liqPrice"`
	RiskLimitValue   string `json:"riskLimitValue"`
	TakeProfit       string `json:"takeProfit"`
	PositionValue    string `json:"positionValue"`
	TpslMode         string `json:"tpslMode"`
	RiskID           int    `json:"riskId"`
	TrailingStop     string `json:"trailingStop"`
	UnrealisedPnl    string `json:"unrealisedPnl"`
	MarkPrice        string `json:"markPrice"`
	AdlRankIndicator int    `json:"adlRankIndicator"`
	CumRealisedPnl   string `json:"cumRealisedPnl"`
	PositionMM       string `json:"positionMM"`
	CreatedTime      string `json:"createdTime"`
	PositionIdx      int    `json:"positionIdx"`
	PositionIM       string `json:"positionIM"`
	UpdatedTime      string `json:"updatedTime"`
	Side             string `json:"side"`
	BustPrice        string `json:"bustPrice"`
	PositionBalance  string `json:"positionBalance"`
	Size             string `json:"size"`
	PositionStatus   string `json:"positionStatus"`
	StopLoss         string `json:"stopLoss"`
	TradeMode        int    `json:"tradeMode"`
}

type PositionsResponse struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
	Result  struct {
		NextPageCursor string     `json:"nextPageCursor"`
		Category       string     `json:"category"`
		List           []Position `json:"list"`
	} `json:"result"`
	RetExtInfo struct{} `json:"retExtInfo"`
	Time       int64    `json:"time"`
}
