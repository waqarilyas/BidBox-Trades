package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //postgres database driver

	log "github.com/sirupsen/logrus"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
}

func (server *Server) Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName string) {

	var err error
	DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
	server.DB, err = gorm.Open(Dbdriver, DBURL)
	if err != nil {
		log.Info("Cannot connect to the %s database", Dbdriver)
		log.Fatal("This is the error:", err)
	} else {
		log.Info("Connected to the %s database", Dbdriver)
	}

	// server.DB.Debug().AutoMigrate(&models.Key{}) //database migration
	server.Router = mux.NewRouter()
	server.initializeRoutes()
}

func (server *Server) Run(addr string) {
	log.Info("Listening on port 8080")
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
