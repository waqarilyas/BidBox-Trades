package controllers

import "github.com/kryptomind/BidBox-Trades/middleware"

func (r *Server) initializeRoutes() {
	s := r.Router.PathPrefix("/trades").Subrouter()

	s.HandleFunc("/", middleware.MiddlewareJSON(r.Home)).Methods("GET")

	/*//accounts routes
	s.HandleFunc("/user", middleware.MiddlewareJSON(r.GetUserBalanceByExchange)).Methods("GET")
	s.HandleFunc("/connected", middleware.MiddlewareJSON(r.GetUserConnectedAccounts)).Methods("GET")

	//trades
	s.HandleFunc("/active_trades", middleware.MiddlewareJSON(r.GetOpenTrades)).Methods("GET")
	s.HandleFunc("/history", middleware.MiddlewareJSON(r.GetClosedTrades)).Methods("GET")
	*/
}
