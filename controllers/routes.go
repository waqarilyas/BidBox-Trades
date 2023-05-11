package controllers

import "github.com/kryptomind/BidBox-Trades/middleware"

func (r *Server) initializeRoutes() {
	s := r.Router.PathPrefix("/trades").Subrouter()

	s.HandleFunc("/", middleware.MiddlewareJSON(r.StartTrade))
	s.HandleFunc("/order", middleware.ValidateEmail(r.PlaceOrder)).Methods("POST")
	s.HandleFunc("/ws", wsEndpoint)
	s.HandleFunc("/amount", middleware.ValidateEmail(r.UpdateAmount)).Methods("POST")
}
