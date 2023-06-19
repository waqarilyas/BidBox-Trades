package binance


type BinanceLimitResponse struct {
	SymbolStruct                      []SymbolStruct    `json:"symbols"`
}
type SymbolStruct struct {
	Symbol                 string   `json:"symbol"`
	Pair                   string   `json:"pair"`
	ContractType           string   `json:"contractType"`
	DeliveryDate           int64    `json:"deliveryDate"`
	OnboardDate            int64    `json:"onboardDate"`
	Status                 string   `json:"status"`
	MaintMarginPercent     string   `json:"maintMarginPercent"`
	RequiredMarginPercent  string   `json:"requiredMarginPercent"`
	BaseAsset              string   `json:"baseAsset"`
	QuoteAsset             string   `json:"quoteAsset"`
	MarginAsset            string   `json:"marginAsset"`
	PricePrecision         int      `json:"pricePrecision"`
	QuantityPrecision      int      `json:"quantityPrecision"`
	BaseAssetPrecision     int      `json:"baseAssetPrecision"`
	QuotePrecision         int      `json:"quotePrecision"`
	UnderlyingType         string   `json:"underlyingType"`
	UnderlyingSubType      []string `json:"underlyingSubType"`
	SettlePlan             int      `json:"settlePlan"`
	TriggerProtect         string   `json:"triggerProtect"`
	LiquidationFee         string   `json:"liquidationFee"`
	MarketTakeBound        string   `json:"marketTakeBound"`
	MaxMoveOrderLimit      int      `json:"maxMoveOrderLimit"`
	Filters                []Filter `json:"filters"`
	OrderTypes             []string `json:"orderTypes"`
	TimeInForce            []string `json:"timeInForce"`
}

type Filter struct {
	MinPrice          string `json:"minPrice,omitempty"`
	MaxPrice          string `json:"maxPrice,omitempty"`
	FilterType        string `json:"filterType"`
	TickSize          string `json:"tickSize,omitempty"`
	StepSize          string `json:"stepSize,omitempty"`
	// MaxQty            string `json:"maxQty,omitempty"`
	// MinQty            string `json:"minQty,omitempty"`
	// Limit             int    `json:"limit,omitempty"`
	// Notional          string `json:"notional,omitempty"`
	// MultiplierDown    string `json:"multiplierDown,omitempty"`
	// MultiplierUp      string `json:"multiplierUp,omitempty"`
	// MultiplierDecimal string `json:"multiplierDecimal,omitempty"`
}

type AccountsResponse struct {
	FeeTier                     int        `json:"feeTier"`
	CanTrade                    bool       `json:"canTrade"`
	CanDeposit                  bool       `json:"canDeposit"`
	CanWithdraw                 bool       `json:"canWithdraw"`
	UpdateTime                  int        `json:"updateTime"`
	MultiAssetsMargin           bool       `json:"multiAssetsMargin"`
	TotalInitialMargin          string     `json:"totalInitialMargin"`
	TotalMaintMargin            string     `json:"totalMaintMargin"`
	TotalWalletBalance          string     `json:"totalWalletBalance"`
	TotalUnrealizedProfit       string     `json:"totalUnrealizedProfit"`
	TotalMarginBalance          string     `json:"totalMarginBalance"`
	TotalPositionInitialMargin  string     `json:"totalPositionInitialMargin"`
	TotalOpenOrderInitialMargin string     `json:"totalOpenOrderInitialMargin"`
	TotalCrossWalletBalance     string     `json:"totalCrossWalletBalance"`
	TotalCrossUnPnl             string     `json:"totalCrossUnPnl"`
	AvailableBalance            string     `json:"availableBalance"`
	MaxWithdrawAmount           string     `json:"maxWithdrawAmount"`
	Assets                      []Asset    `json:"assets"`
	Positions                   []Position `json:"positions"`
}

type Asset struct {
	Asset                  string `json:"asset"`
	WalletBalance          string `json:"walletBalance"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	MarginBalance          string `json:"marginBalance"`
	MaintMargin            string `json:"maintMargin"`
	InitialMargin          string `json:"initialMargin"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	MaxWithdrawAmount      string `json:"maxWithdrawAmount"`
	CrossWalletBalance     string `json:"crossWalletBalance"`
	CrossUnPnl             string `json:"crossUnPnl"`
	AvailableBalance       string `json:"availableBalance"`
	MarginAvailable        bool   `json:"marginAvailable"`
	UpdateTime             int    `json:"updateTime"`
}

type Position struct {
	Symbol                 string `json:"symbol"`
	InitialMargin          string `json:"initialMargin"`
	MaintMargin            string `json:"maintMargin"`
	UnrealizedProfit       string `json:"unrealizedProfit"`
	PositionInitialMargin  string `json:"positionInitialMargin"`
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"`
	Leverage               string `json:"leverage"`
	Isolated               bool   `json:"isolated"`
	EntryPrice             string `json:"entryPrice"`
	MaxNotional            string `json:"maxNotional"`
	PositionSide           string `json:"positionSide"`
	PositionAmt            string `json:"positionAmt"`
	Notional               string `json:"notional"`
	IsolatedWallet         string `json:"isolatedWallet"`
	UpdateTime             int    `json:"updateTime"`
	BidNotional            string `json:"bidNotional"`
	AskNotional            string `json:"askNotional"`
	LiquidationPrice       string `json:"liquidationPrice"`
	MarkPrice              string `json:"markPrice"`
	MarginType             string `json:"marginType"`
}


type BinanceErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
