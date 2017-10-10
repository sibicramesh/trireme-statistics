package configuration

import (
	"github.com/docopt/docopt-go"
)

const (
	usage = `Collector.
Usage: collector
       [--db-user=<user>]
       [--db-pass=<pass>]
       [--db-name=<name>]
       [--db-address=<address>]
       [--ui-user=<user>]
       [--ui-pass=<pass>]
       [--ui-db-address=<address>]
			 [--ui-db-access=<access>]
       [--listen=<address>]
			 collector -h | --help
			 collector -v | --version
Options:
    -h --help                                   Show this screen.
    -v --version                                Show the version.
    --listen=<address>                          Listening address [default: :8080].

DB Options:
		--db-user=<user>          Username of the database [default: aporeto].
		--db-pass=<pass>         Password of the database [default: aporeto].
		--db-name=<name>         Name of the database [default: flowDB].
		--db-address=<address>   Address to connect to DB [default: http://influxdb:8086]

UI Options:
		--ui-user=<user>       Username of the UI to connect with [default: admin].
		--ui-pass=<pass>       Password of the UI to connect with [default: admin].
		--ui-db-address=<address>   Address to connect to UI [default: http://grafana:3000]
		--ui-db-access=<access>   Access to connect to DB [default: proxy]
`
)

// Configuration stuct is used to populate the various fields used by collector
type Configuration struct {
	ListenAddress string

	DBUserName string
	DBPassword string
	DBName     string
	DBAddress  string

	UIUserName string
	UIPassword string
	UIAddress  string
	UIDBAccess string
}

// NewConfiguration will parse arguements and return new Configuration struct
func NewConfiguration() *Configuration {

	arguments, err := docopt.Parse(usage, nil, true, "", false)
	if err != nil {
		panic("Unable to parse usage")
	}

	return &Configuration{

		ListenAddress: arguments["--listen"].(string),

		DBUserName: arguments["--db-user"].(string),
		DBPassword: arguments["--db-pass"].(string),
		DBName:     arguments["--db-name"].(string),
		DBAddress:  arguments["--db-address"].(string),
		UIUserName: arguments["--ui-user"].(string),
		UIPassword: arguments["--ui-pass"].(string),
		UIAddress:  arguments["--ui-db-address"].(string),
		UIDBAccess: arguments["--ui-db-access"].(string),
	}
}
