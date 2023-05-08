package controllers

import "github.com/kryptomind/BidBox-Trades/middleware"

func (r *Server) initializeRoutes() {
	s := r.Router.PathPrefix("/trades").Subrouter()

	s.HandleFunc("/", middleware.ValidateEmail(r.Home)).Methods("GET")
	s.HandleFunc("/", middleware.MiddlewareJSON(r.StartTrade)).Methods("POST")
	s.HandleFunc("/order", middleware.ValidateEmail(r.PlaceOrder)).Methods("POST")

	s.HandleFunc("/amount", middleware.ValidateEmail(r.UpdateAmount)).Methods("POST")
}
