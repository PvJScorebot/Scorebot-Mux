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

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	mux "github.com/iDigitalFlame/scorebot-mux"
)

const (
	version = "v2.0"
)

func main() {
	ConfigFile := flag.String("c", "", "Muxer Config File Path.")
	ConfigDefault := flag.Bool("d", false, "Print Default Config and Exit.")

	Listen := flag.String("bind", mux.DefaultListen, "Address and Port to Listen on.")

	Key := flag.String("key", "", "Path to TLS Key File.")
	Cert := flag.String("cert", "", "Path to TLS Certificate File.")

	Timeout := flag.Int("timeout", int(mux.DefaultTimeout), "Muxer Request Timeout. (in seconds)")

	Scorebot := flag.String("sbe", "", "Scorebot Core Address or URL.")

	Database := flag.String("db", "", "Muxer Database Hostname or Address.")
	DatabaseName := flag.String("db-name", "", "Muxer Database Name.")
	DatabaseUser := flag.String("db-user", "", "Muxer Database Username.")
	DatabasePassword := flag.String("db-password", "", "Muxer Database Password.")

	Proxy := flag.String("proxy", "", "URL to secondary proxy to use.")

	flag.Usage = func() {
		fmt.Printf(
			"Scorebot Muxer %s\n2019 iDigitalFlame, The Scorebot Project, CTF Factory\n\nUsage:\n",
			version,
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *ConfigDefault {
		fmt.Printf("%s\n", mux.Defaults())
		os.Exit(0)
	}

	var c *mux.Config
	if len(*ConfigFile) > 0 {
		var err error
		c, err = mux.Load(*ConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
	} else {
		if len(*Scorebot) == 0 || len(*Listen) == 0 || *Timeout < 0 || len(*Database) == 0 || len(*DatabaseName) == 0 || len(*DatabaseUser) == 0 {
			flag.Usage()
			os.Exit(2)
		}
		c = &mux.Config{
			Key:     *Key,
			Cert:    *Cert,
			Listen:  *Listen,
			Timeout: time.Duration(*Timeout) * time.Second,
			Proxies: []*mux.Secondary{
				&mux.Secondary{
					URL:    *Proxy,
					Ignore: false,
				},
			},
			Scorebot: *Scorebot,
			Database: &mux.Database{
				Host:     *Database,
				User:     *DatabaseUser,
				Database: *DatabaseName,
				Password: *DatabasePassword,
			},
		}
	}

	m, err := mux.New(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
	if err := m.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
