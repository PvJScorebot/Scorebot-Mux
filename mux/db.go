package mux

import (
	"database/sql"
	"fmt"
	"net/http"

	// Import MySQL Driver
	_ "github.com/go-sql-driver/mysql"
)

const (
	sqlCreateRequests = "" +
		"CREATE TABLE IF NOT EXISTS requests (" +
		"ID BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT, " +
		"Time DATETIME NOT NULL, " +
		"URL VARCHAR(256) NOT NULL, " +
		"Path VARCHAR(256) NOT NULL, " +
		"IP VARCHAR(128) NOT NULL, " +
		"Method VARCHAR(8) NOT NULL, " +
		"Token VARCHAR(256) NOT NULL, " +
		"Data VARBINARY(32768) NULL)"
	sqlCreateResponse = "" +
		"CREATE TABLE IF NOT EXISTS responses (" +
		"ID BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT, " +
		"Time DATETIME NOT NULL, " +
		"URL VARCHAR(256) NOT NULL, " +
		"Path VARCHAR(256) NOT NULL, " +
		"IP VARCHAR(128) NOT NULL, " +
		"Method VARCHAR(8) NOT NULL, " +
		"Result SMALLINT NOT NULL, " +
		"Data VARBINARY(32768) NULL)"
	sqlInsertRequest = "" +
		"INSERT INTO requests (" +
		"Time, URL, Path, IP, Method, Token, Data" +
		") VALUES (NOW(), ?, ?, ?, ?, ?, ?)"
	sqlInsertResponse = "" +
		"INSERT INTO responses (" +
		"Time, URL, Path, IP, Method, Result, Data" +
		") VALUES (NOW(), ?, ?, ?, ?, ?, ?)"
)

// Database represents the SQL Backend data provider.
type Database struct {
	Host     string `json:"host"`
	User     string `json:"username"`
	Database string `json:"name"`
	Password string `json:"password"`

	db       *sql.DB
	request  *sql.Stmt
	response *sql.Stmt
}

func (d *Database) init() error {
	var err error
	if d.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s", d.User, d.Password, d.Host, d.Database)); err != nil {
		return fmt.Errorf("could not open mysql database: %w", err)
	}
	if s, err := d.db.Prepare(sqlCreateRequests); err == nil {
		defer s.Close()
		if _, err := s.Exec(); err != nil {
			return fmt.Errorf("could not create requests table: %w", err)
		}
	} else {
		return fmt.Errorf("could no prepare SQL statement: %w", err)
	}
	if s, err := d.db.Prepare(sqlCreateResponse); err == nil {
		defer s.Close()
		if _, err := s.Exec(); err != nil {
			return fmt.Errorf("could not create responses table: %w", err)
		}
	} else {
		return fmt.Errorf("could no prepare SQL statement: %w", err)
	}
	if d.request, err = d.db.Prepare(sqlInsertRequest); err != nil {
		return fmt.Errorf("could not prepare SQL statement: %w", err)
	}
	if d.response, err = d.db.Prepare(sqlInsertResponse); err != nil {
		return fmt.Errorf("could not prepare SQL statement: %w", err)
	}
	return nil
}
func (d *Database) close() error {
	d.request.Close()
	d.response.Close()
	return d.db.Close()
}
func (d *Database) saveRequest(url, path, ip, method string, header http.Header, data []byte) {
	token := header.Get("SBE-AUTH")
	d.request.Exec(url, path, ip, method, token, data)
}
func (d *Database) saveResponse(url, path, ip, method string, status int, header http.Header, data []byte) {
	d.response.Exec(url, path, ip, method, status, data)
}
