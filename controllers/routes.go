package controllers

import "github.com/kryptomind/BidBox-Trades/middleware"

func (r *Server) initializeRoutes() {
	s := r.Router.PathPrefix("/trades").Subrouter()

	s.HandleFunc("/", middleware.ValidateEmail(r.Home)).Methods("GET")
	s.HandleFunc("/order", middleware.ValidateEmail(r.PlaceOrder)).Methods("POST")

	/*//accounts routes
	s.HandleFunc("/user", middleware.MiddlewareJSON(r.GetUserBalanceByExchange)).Methods("GET")
	s.HandleFunc("/connected", middleware.MiddlewareJSON(r.GetUserConnectedAccounts)).Methods("GET")

	//trades
	s.HandleFunc("/active_trades", middleware.MiddlewareJSON(r.GetOpenTrades)).Methods("GET")
	s.HandleFunc("/history", middleware.MiddlewareJSON(r.GetClosedTrades)).Methods("GET")
	*/
}
