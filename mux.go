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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iDigitalFlame/switchproxy"
)

// Mux is a struct that represents a Muxer that can split and log
// traffic between two or more endpoints
type Mux struct {
	err    error
	proxy  *switchproxy.Proxy
	config *Config
}

// Start is a method that starts the Mux process and inits the
// database and proxy switch.
func (m *Mux) Start() error {
	if m.config == nil || m.config.Database == nil {
		return ErrInvalidConfig
	}
	if err := m.config.Database.init(); err != nil {
		return fmt.Errorf("unable to init database: %w", err)
	}
	defer m.config.Database.close()
	m.proxy = switchproxy.New(
		m.config.Listen,
		switchproxy.Timeout(m.config.Timeout),
		switchproxy.TLS(m.config.Cert, m.config.Key),
	)
	for i := range m.config.Proxies {
		s, err := switchproxy.NewSwitchTimeout(m.config.Proxies[i].URL, m.config.Timeout)
		if err != nil {
			return fmt.Errorf("unable to configure secondary switch: %w", err)
		}
		for k, v := range m.config.Proxies[i].Rewrite {
			s.Rewrite(k, v)
		}
		if !m.config.Proxies[i].Ignore {
			s.Pre = m.config.Database.log
			s.Post = m.config.Database.log
		}
		m.proxy.AddSecondary(s)
	}
	p, err := switchproxy.NewSwitchTimeout(m.config.Scorebot, m.config.Timeout)
	if err != nil {
		return fmt.Errorf("unable to configure primary switch: %w", err)
	}
	p.Pre = m.config.Database.log
	p.Post = m.config.Database.log
	m.proxy.Primary(p)
	w := make(chan os.Signal, 1)
	signal.Notify(w, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go m.start(w)
	<-w
	signal.Stop(w)
	m.proxy.Stop()
	return m.err
}
func (m *Mux) start(c chan os.Signal) {
	defer func() { recover() }()
	m.err = m.proxy.Start()
	if len(c) == 0 {
		c <- syscall.SIGINT
	}
}
