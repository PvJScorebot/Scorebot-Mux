package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/iDigitalFlame/scorebot-mux/mux"
)

const (
	version = "v1.0"
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

	mux, err := mux.NewMux(c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	if err := mux.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
