package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func main() {

	handler := CreateHandler()

	serv := http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 0,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Fatal(serv.ListenAndServe())

}

func CreateHandler() (handler *http.ServeMux) {
	handler = http.NewServeMux()
	handler.HandleFunc("/test", handleFunc)
	return
}

func handleFunc(writer http.ResponseWriter, request *http.Request) {

	writer.Header().Set("Content-Type", "application/json")

	employee := Employee{
		Id:        12,
		Name:      "John",
		Surname:   "Crammer",
		Phone:     "555-01-00",
		CompanyId: 1,
		Passport: Passport{
			Type:   "USA-passport",
			Number: "12",
		},
	}
	emplJSON, _ := json.Marshal(employee)

	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(emplJSON)
}

type Employee struct {
	Id        int
	Name      string
	Surname   string
	Phone     string
	CompanyId int
	Passport  Passport
}

type Passport struct {
	Type   string
	Number string
}
