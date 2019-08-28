package main

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
