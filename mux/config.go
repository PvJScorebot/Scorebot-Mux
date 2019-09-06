package mux

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/iDigitalFlame/switchproxy/proxy"
	"golang.org/x/xerrors"
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

// Mux is a struct that represents a Muxer that can split and log
// traffic between two or more endpoints
type Mux struct {
	proxy  *proxy.Proxy
	config *Config
}

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

// Start is a method that starts the Mux process and inits the
// database and proxy switch.
func (m *Mux) Start() error {
	if m.config == nil {
		return ErrInvalidConfig
	}
	if m.config.Database == nil {
		return ErrInvalidConfig
	}
	if err := m.config.Database.init(); err != nil {
		return xerrors.Errorf("unable to init database: %w", err)
	}
	m.proxy = proxy.NewProxyEx(m.config.Listen, m.config.Cert, m.config.Key)
	for i := range m.config.Proxies {
		s, err := proxy.NewSwitch(m.config.Proxies[i].URL, m.config.Timeout)
		if err != nil {
			return xerrors.Errorf("unable to configure secondary switch: %w", err)
		}
		for k, v := range m.config.Proxies[i].Rewrite {
			s.Rewrite(k, v)
		}
		if !m.config.Proxies[i].Ignore {
			s.Pre = m.config.Database.saveRequest
			s.Post = m.config.Database.saveResponse
		}
		m.proxy.AddSecondary(s)
	}
	p, err := proxy.NewSwitch(m.config.Scorebot, m.config.Timeout)
	if err != nil {
		return xerrors.Errorf("unable to configure primary switch: %w", err)
	}
	p.Pre = m.config.Database.saveRequest
	p.Post = m.config.Database.saveResponse
	m.proxy.Primary(p)
	err = m.proxy.Start()
	m.config.Database.close()
	return err
}

// NewMux creates and returns a new Mux instance from the specified config.
func NewMux(c *Config) (*Mux, error) {
	if c.Timeout < 0 {
		return nil, ErrInvalidTimeout
	}
	if c.Database == nil {
		return nil, ErrInvalidConfig
	}
	if len(c.Listen) == 0 {
		c.Listen = DefaultListen
	}
	c.Timeout *= time.Second
	return &Mux{config: c}, nil
}

// Load attempts to create a load a config file from the specified path 's'.
func Load(s string) (*Config, error) {
	f, err := os.Stat(s)
	if err != nil {
		return nil, xerrors.Errorf("cannot load file \"%s\": %w", s, err)
	}
	if f.IsDir() {
		return nil, xerrors.Errorf("cannot load \"%s\": path is not a file", s)
	}
	var c *Config
	b, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, xerrors.Errorf("unable to read file \"%s\": %w", s, err)
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, xerrors.Errorf("cannot read file \"%s\" into JSON: %w", s, err)
	}
	return c, nil
}
