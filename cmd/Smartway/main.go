package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const path = "/employee/"

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalln(err)
	}

	_, _ = db.Exec(startQuery)
	handler := http.NewServeMux()
	handler.HandleFunc(path, handleFunc)

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

func handleFunc(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	switch request.Method {
	case http.MethodGet:
		getEmployeesByCompanyListener(writer, request)
	case http.MethodPost:
		addEmployeeListener(writer, request)
	case http.MethodPut:
		updateEmployeeListener(writer, request)
	case http.MethodDelete:
		deleteEmployeeListener(writer, request)
	default:

	}
}

func getEmployeesByCompanyListener(writer http.ResponseWriter, request *http.Request) {
	companyId, err := getIdFromURL(request.URL.Path)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	employeesOfCompany := make([]Employee, 0)
	rows, err := db.Query("select e.id, e.name, e.surname, e.phone, e.company_id, p.type, p.number from employees e inner join passport p on e.passport_id = p.id where e.company_id = $1;", companyId)
	defer rows.Close()
	if err != nil {
		return
	}
	for rows.Next() {
		empl := Employee{
			Passport: Passport{},
		}
		err = rows.Scan(&empl.Id,
			&empl.Name,
			&empl.Surname,
			&empl.Phone,
			&empl.CompanyId,
			&empl.Passport.Type,
			&empl.Passport.Number)
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

func addEmployeeListener(writer http.ResponseWriter, request *http.Request) {
	var employee Employee
	err := json.NewDecoder(request.Body).Decode(&employee)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	var passportId, employeeId int
	_ = db.QueryRow("INSERT INTO passport (type, number) VALUES ($1, $2) returning id;",
		employee.Passport.Type,
		employee.Passport.Number).Scan(&passportId)

	_ = db.QueryRow("INSERT INTO employees (name, surname, phone, company_id, passport_id) VALUES ($1, $2, $3, $4, $5) returning id;",
		employee.Name,
		employee.Surname,
		employee.Phone,
		employee.CompanyId,
		passportId).Scan(&employeeId)

	writer.WriteHeader(http.StatusOK)
	outPut, _ := json.Marshal(fmt.Sprintf("Id of new employee is %d", employeeId))
	_, _ = writer.Write(outPut)

}

func updateEmployeeListener(writer http.ResponseWriter, request *http.Request) {
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
	var count int
	_ = db.QueryRow("SELECT COUNT(*) FROM employees WHERE id = $1", id).Scan(&count)
	if count == 0 {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	// TODO: реализовать изменение структуры passport
	if employee.Name != "" {
		_, _ = db.Exec("UPDATE employees SET name=$1 WHERE id = $2", employee.Name, id)
	}
	if employee.Surname != "" {
		_, _ = db.Exec("UPDATE employees SET surname=$1 WHERE id = $2", employee.Surname, id)
	}
	if employee.Phone != "" {
		_, _ = db.Exec("UPDATE employees SET phone=$1 WHERE id = $2", employee.Phone, id)
	}
	if employee.CompanyId != 0 {
		_, _ = db.Exec("UPDATE employees SET company_id=$1 WHERE id = $2", employee.CompanyId, id)
	}

	writer.WriteHeader(http.StatusOK)
}

func deleteEmployeeListener(writer http.ResponseWriter, request *http.Request) {
	id, err := getIdFromURL(request.URL.Path)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	_, _ = db.Exec("DELETE FROM employees WHERE id = $1;", id)

}

func getIdFromURL(url string) (int, error) {
	param := strings.Replace(url, path, "", 1)
	return strconv.Atoi(param)
}
