package mux

import (
	"database/sql"
	"fmt"

	// Import MySQL Driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/iDigitalFlame/switchproxy/proxy"
)

const (
	sqlCreateRequests = "" +
		"CREATE TABLE IF NOT EXISTS requests (" +
		"ID BIGINT NOT NULL PRIMARY KEY AUTO_INCREMENT, " +
		"UUID CHAR(36) NOT NULL, " +
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
		"UUID CHAR(36) NOT NULL, " +
		"Time DATETIME NOT NULL, " +
		"URL VARCHAR(256) NOT NULL, " +
		"Path VARCHAR(256) NOT NULL, " +
		"IP VARCHAR(128) NOT NULL, " +
		"Method VARCHAR(8) NOT NULL, " +
		"Result SMALLINT NOT NULL, " +
		"Data VARBINARY(32768) NULL)"
	sqlInsertRequest = "" +
		"INSERT INTO requests (" +
		"UUID, Time, URL, Path, IP, Method, Token, Data" +
		") VALUES (?, NOW(), ?, ?, ?, ?, ?, ?)"
	sqlInsertResponse = "" +
		"INSERT INTO responses (" +
		"UUID, Time, URL, Path, IP, Method, Result, Data" +
		") VALUES (?, NOW(), ?, ?, ?, ?, ?, ?)"
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
	s, err := d.db.Prepare(sqlCreateRequests)
	if err != nil {
		return fmt.Errorf("could not prepare SQL statement: %w", err)
	}
	_, err = s.Exec()
	if s.Close(); err != nil {
		return fmt.Errorf("could not create requests table: %w", err)
	}
	s, err = d.db.Prepare(sqlCreateResponse)
	if err != nil {
		return fmt.Errorf("could not prepare SQL statement: %w", err)
	}
	_, err = s.Exec()
	if s.Close(); err != nil {
		return fmt.Errorf("could not create responses table: %w", err)
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
	if d.request != nil {
		d.request.Close()
	}
	if d.response != nil {
		d.response.Close()
	}
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}
func (d *Database) log(r proxy.Result) {
	if r.IsResponse() {
		d.response.Exec(r.UUID, r.URL, r.Path, r.IP, r.Method, r.Status, r.Content)
	} else {
		d.request.Exec(r.UUID, r.URL, r.Path, r.IP, r.Method, r.Headers.Get("SBE-AUTH"), r.Content)
	}
}
