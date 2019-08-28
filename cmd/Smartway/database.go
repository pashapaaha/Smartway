package main

const dbUrl = "host=localhost user=postgres password=12345 dbname=employeesDB sslmode=disable"

var startQuery = `
CREATE SEQUENCE IF NOT EXISTS employees_seq;
CREATE SEQUENCE IF NOT EXISTS passport_seq;
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
