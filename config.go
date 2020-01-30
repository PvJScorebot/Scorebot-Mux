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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

const (
	// DefaultListen is the default listening address for the Mux.
	DefaultListen = "0.0.0.0:8080"
	// DefaultTimeout is the default Mux connection timeout.
	DefaultTimeout = 5
)

var (
	// ErrInvalidConfig is returned by the 'Start' function when the config is missing or
	// is not valid during startup.
	ErrInvalidConfig = errors.New("configuration is invalid or missing")
	// ErrInvalidTimeout is returned if the specified timeout value is less than zero.
	ErrInvalidTimeout = errors.New("timeout must be grater or equal to than zero")
)

// Config is a struct that contains the configuration options needed to
// start a Mux struct.
type Config struct {
	Key      string        `json:"key,omitempty"`
	Cert     string        `json:"cert,omitempty"`
	Listen   string        `json:"listen,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
	Proxies  []*Secondary  `json:"proxies,omitempty"`
	Scorebot string        `json:"scorebot"`
	Database *Database     `json:"db"`
}

// Secondary is a struct that holds secondary Proxy info.
type Secondary struct {
	URL     string            `json:"url"`
	Ignore  bool              `json:"ignore,omitempty"`
	Rewrite map[string]string `json:"rewrite,omitempty"`
}

// Defaults returns a JSON string representation of the default config.
// Used for creating and understanding the config file structure.
func Defaults() string {
	c := &Config{
		Key:     "",
		Cert:    "",
		Listen:  DefaultListen,
		Timeout: DefaultTimeout,
		Proxies: []*Secondary{
			&Secondary{
				URL:    "http://proxy1",
				Ignore: false,
				Rewrite: map[string]string{
					"/url1": "/url2",
				},
			},
		},
		Scorebot: "http://scorebot",
		Database: &Database{
			Host:     "tcp(mysql:3306)",
			User:     "muxer",
			Database: "muxdb",
			Password: "password",
		},
	}
	b, _ := json.MarshalIndent(c, "", "    ")
	return string(b)
}

// New creates and returns a new Mux instance from the specified config.
func New(c *Config) (*Mux, error) {
	if c.Timeout < 0 {
		return nil, ErrInvalidTimeout
	}
	if c.Database == nil {
		return nil, fmt.Errorf("database: %w", ErrInvalidConfig)
	}
	if len(c.Scorebot) == 0 {
		return nil, fmt.Errorf("scorebot: %w", ErrInvalidConfig)
	}
	if len(c.Listen) == 0 {
		c.Listen = DefaultListen
	}
	c.Timeout *= time.Second
	return &Mux{config: c}, nil
}

// Load attempts to create a load a config file from the specified path.
func Load(s string) (*Config, error) {
	f, err := os.Stat(s)
	if err != nil {
		return nil, fmt.Errorf("cannot load file \"%s\": %w", s, err)
	}
	if f.IsDir() {
		return nil, fmt.Errorf("cannot load \"%s\": path is not a file", s)
	}
	var c *Config
	b, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, fmt.Errorf("unable to read file \"%s\": %w", s, err)
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("cannot read file \"%s\" into JSON: %w", s, err)
	}
	return c, nil
}
