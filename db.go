// Copyright(C) 2020 iDigitalFlame
//
// This program is free software: you can redistribute it and / or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.If not, see <https://www.gnu.org/licenses/>.
//

package mux

import (
	"database/sql"
	"fmt"

	"github.com/PurpleSec/switchproxy"

	// Import MySQL Driver
	_ "github.com/go-sql-driver/mysql"
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
	response *sql.Stmt
	db       *sql.DB
	request  *sql.Stmt
	User     string `json:"username"`
	Database string `json:"name"`
	Password string `json:"password"`
	Host     string `json:"host"`
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
func (d *Database) log(r switchproxy.Result) {
	if r.IsResponse() {
		d.response.Exec(r.UUID, r.URL, r.Path, r.IP, r.Method, r.Status, r.Content)
	} else {
		d.request.Exec(r.UUID, r.URL, r.Path, r.IP, r.Method, r.Headers.Get("SBE-AUTH"), r.Content)
	}
}
