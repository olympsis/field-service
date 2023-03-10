package main

import (
	"context"
	"net/http"
	field "olympsis-services/field/service"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	// logger
	l := logrus.New()

	// mux router
	r := mux.NewRouter()

	// field service
	service := field.NewFieldService(l, r)

	// connecting to database
	res, err := service.ConnectToDatabase()

	// quit on err
	if !res || err != nil {
		os.Exit(1)
	}

	r.Handle("/v1/healthz", service.Healthz()).Methods("GET")

	// handlers for http requests
	// get fields
	r.Handle("/v1/fields", service.GetFields()).Methods("GET")

	// get a field
	r.Handle("/v1/fields/{id}", service.GetAField()).Methods("GET")

	// create field
	r.Handle("/v1/fields", service.InsertAField()).Methods("POST")

	// update a field
	r.Handle("/v1/fields/{id}", service.UpdateAField()).Methods("PUT")

	// delete a field
	r.Handle("/v1/fields/{id}", service.DeleteAField()).Methods("DELETE")

	port := os.Getenv("PORT")

	// server config
	s := &http.Server{
		Addr:         `:` + port, // pull from env
		Handler:      r,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// start server
	go func() {
		l.Info(`Starting Field Service at...` + port)
		err := s.ListenAndServe()

		if err != nil {
			l.Info("Error Starting Server: ", err)
			os.Exit(1)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs

	l.Printf("Recieved Termination(%s), graceful shutdown \n", sig)

	tc, c := context.WithTimeout(context.Background(), 30*time.Second)

	defer c()

	s.Shutdown(tc)

}
