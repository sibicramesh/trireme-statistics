package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/rs/cors"

	"github.com/aporeto-inc/trireme-statistics/configuration"
	"github.com/aporeto-inc/trireme-statistics/grafana"
	"github.com/aporeto-inc/trireme-statistics/graph/server"
	"github.com/aporeto-inc/trireme-statistics/influxdb"
)

func banner() {
	fmt.Printf(`
Trireme-Stats
`)
}

func main() {
	banner()

	cfg, err := configuration.LoadConfiguration()
	if err != nil {
		log.Fatal("Error parsing configuration", err)
	}

	httlpcli, err := influxdb.NewDBConnection(cfg.DBUserName, cfg.DBPassword, cfg.DBAddress)
	if err != nil {
		zap.L().Fatal("Error: Initiating Connection to DB", zap.Error(err))
	}

	err = httlpcli.CreateDB(cfg.DBName)
	if err != nil {
		zap.L().Fatal("Error: Creating Database", zap.Error(err))
	}

	time.Sleep(time.Second * 5)
	graphanasession, err := grafana.NewUISession(cfg.UIUserName, cfg.UIPassword, cfg.UIAddress)
	if err != nil {
		zap.L().Fatal("Error: Initiating Connection to Grafana Server", zap.Error(err))
	}

	err = graphanasession.CreateDataSource("Events", cfg.DBName, cfg.DBUserName, cfg.DBPassword, cfg.DBAddress, cfg.UIDBAccess)
	if err != nil {
		zap.L().Fatal("Error: Creating Datasource", zap.Error(err))
	}

	graphanasession.CreateDashboard("StatisticBoard")
	graphanasession.AddRows(grafana.SingleStat, "events", "Action", "FlowEvents")
	graphanasession.AddRows(grafana.SingleStat, "events", "IPAddress", "ContainerEvents")

	mux := http.NewServeMux()
	mux.HandleFunc("/get", server.GetData(httlpcli))
	mux.HandleFunc("/graph", server.GetGraph)

	handler := cors.Default().Handler(mux)

	fmt.Println("Server Listening at", cfg.ListenAddress)

	err = http.ListenAndServe(cfg.ListenAddress, handler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
