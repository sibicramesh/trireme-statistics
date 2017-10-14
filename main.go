package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

	cfg := configuration.NewConfiguration()

	httlpcli, err := influxdb.NewDBConnection(cfg.DBUserName, cfg.DBPassword, cfg.DBAddress)
	if err != nil {
		log.Fatal("Failed to initialize connection to database", err)
	}

	err = httlpcli.CreateDB(cfg.DBName)
	if err != nil {
		log.Fatal("Failed to create database", err)
	}

	time.Sleep(time.Second * 5)
	graphanasession, err := grafana.NewUISession(cfg.UIUserName, cfg.UIPassword, cfg.UIAddress)
	if err != nil {
		log.Fatal("Failed to initialize a session to grafana", err)
	}

	err = graphanasession.CreateDataSource("Events", cfg.DBName, cfg.DBUserName, cfg.DBPassword, cfg.DBAddress, cfg.UIDBAccess)
	if err != nil {
		log.Fatal("Failed to create datasource", err)
	}

	graphanasession.CreateDashboard("StatisticBoard")
	graphanasession.AddRows(grafana.SingleStat, "events", "Action", "FlowEvents")
	graphanasession.AddRows(grafana.SingleStat, "events", "IPAddress", "ContainerEvents")

	mux := http.NewServeMux()
	mux.HandleFunc("/get", server.GetData)
	mux.HandleFunc("/graph", server.GetGraph)

	handler := cors.Default().Handler(mux)

	log.Println("Server Listening at", cfg.ListenAddress)

	err = http.ListenAndServe(cfg.ListenAddress, handler)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
