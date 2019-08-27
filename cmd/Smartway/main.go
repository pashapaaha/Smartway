package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

const (
	path  = "/employee/"
	dbUrl = "host=localhost user=postgres password=12345 dbname=employeesDB sslmode=disable"
)

var startQuery = `
CREATE TABLE IF NOT EXISTS passport
(
    id   BIGINT PRIMARY KEY,
    type varchar(255),
    number varchar(255)
);

CREATE TABLE IF NOT EXISTS employees
(
    id   BIGINT PRIMARY KEY,
    name VARCHAR(255),
    surname VARCHAR(255),
    phone VARCHAR(255),
    company_id BIGINT,
    passport_id BIGINT,
    FOREIGN KEY (passport_id) REFERENCES passport(id)
);`

var employees = make([]Employee, 0)
var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalln(err)
	}

	Init()
	handler := CreateHandler()

	server := http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 0,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Fatal(server.ListenAndServe())

}

func Init() {
	db.Exec(startQuery)
	employee1 := Employee{
		Id:        1,
		Name:      "John",
		Surname:   "Crammer",
		Phone:     "555-01-00",
		CompanyId: 1,
		Passport: Passport{
			Type:   "USA-passport",
			Number: "12",
		},
	}
	employee2 := Employee{
		Id:        2,
		Name:      "Ivan",
		Surname:   "Ivanov",
		Phone:     "01010101",
		CompanyId: 2,
		Passport: Passport{
			Type:   "USA-passport",
			Number: "12",
		},
	}
	employee3 := Employee{
		Id:        3,
		Name:      "Petr",
		Surname:   "Petrov",
		Phone:     "1111111",
		CompanyId: 2,
		Passport: Passport{
			Type:   "USA-passport",
			Number: "12",
		},
	}
	employees = append(employees, employee1, employee2, employee3)
}

func CreateHandler() (handler *http.ServeMux) {
	handler = http.NewServeMux()
	handler.HandleFunc(path, handleFunc)
	return
}

func handleFunc(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	switch request.Method {
	case http.MethodGet:
		getEmployeesByCompanyHandler(writer, request)
	case http.MethodPost:
		addEmployeeHandler(writer, request)
	case http.MethodPut:
		updateEmployeeHandler(writer, request)
	case http.MethodDelete:
		deleteEmployeeHandler(writer, request)
	default:

	}
}

func getEmployeesByCompanyHandler(writer http.ResponseWriter, request *http.Request) {
	companyId, err := getIdFromURL(request.URL.Path)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	employeesOfCompany := make([]Employee, 0)
	rows, err := db.Query("select e.id, e.name, e.surname, e.phone, e.company_id, p.type, p.number from employees e inner join passport p on e.passport_id = p.id where e.company_id = $1;", companyId)
	if err != nil {

	}
	defer rows.Close()
	for rows.Next() {
		empl := Employee{
			Passport: Passport{},
		}
		err = rows.Scan(&empl.Id, &empl.Name, &empl.Surname, &empl.Phone, &empl.CompanyId, &empl.Passport.Type, &empl.Passport.Number)
		if err != nil {
			continue
		}
		employeesOfCompany = append(employeesOfCompany, empl)
	}

	employeesJSON, err1 := json.Marshal(employeesOfCompany)
	if err1 != nil {
		writer.WriteHeader(http.StatusBadRequest)
	} else {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(employeesJSON)
	}
}

func addEmployeeHandler(writer http.ResponseWriter, request *http.Request) {
	var employee Employee
	err := json.NewDecoder(request.Body).Decode(&employee)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	newId := maxEmployeeId() + 1
	employee.Id = newId
	employees = append(employees, employee)
	writer.WriteHeader(http.StatusOK)
	outPut, _ := json.Marshal(newId)
	_, _ = writer.Write(outPut)

}

func maxEmployeeId() int {
	max := employees[0].Id
	for _, empl := range employees {
		if empl.Id > max {
			max = empl.Id
		}
	}
	return max
}

func updateEmployeeHandler(writer http.ResponseWriter, request *http.Request) {
	id, err := getIdFromURL(request.URL.Path)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	var employee Employee
	err = json.NewDecoder(request.Body).Decode(&employee)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	var idIsFound = false
	for index, temp := range employees {
		if temp.Id == id {
			employees[index] = employee
			employees[index].Id = id
			idIsFound = true
			return
		}
	}
	if !idIsFound {
		writer.WriteHeader(http.StatusNotFound)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}

func deleteEmployeeHandler(writer http.ResponseWriter, request *http.Request) {
	id, err := getIdFromURL(request.URL.Path)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	for index, temp := range employees {
		if temp.Id == id {
			employees = append(employees[:index], employees[index+1:]...)
			return
		}
	}
}

func getIdFromURL(url string) (int, error) {
	param := strings.Replace(url, path, "", 1)
	return strconv.Atoi(param)
}
