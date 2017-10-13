package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/rs/cors"

	"github.com/aporeto-inc/trireme-statistics/collector/configuration"
	"github.com/aporeto-inc/trireme-statistics/collector/grafana"
	"github.com/aporeto-inc/trireme-statistics/collector/graph/server"
	"github.com/aporeto-inc/trireme-statistics/collector/influxdb"
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
		zap.L().Fatal("Failed to connect to db", zap.Error(err))
	}

	err = httlpcli.CreateDB(cfg.DBName)
	if err != nil {
		zap.L().Fatal("Failed to create DB", zap.Error(err))
	}

	time.Sleep(time.Second * 10)
	graphanasession, err := grafana.NewUI(cfg.UIUserName, cfg.UIPassword, cfg.UIAddress)
	if err != nil {
		zap.L().Fatal("Failed to connect to ui", zap.Error(err))
	}

	err = graphanasession.CreateDataSource("Events", cfg.DBName, cfg.DBUserName, cfg.DBPassword, cfg.DBAddress, cfg.UIDBAccess)
	if err != nil {
		fmt.Println(err)
		zap.L().Fatal("Failed to create datasource", zap.Error(err))
	}

	graphanasession.CreateDashboard("StatisticBoard")
	graphanasession.AddRows(grafana.SingleStat, "events", "Action", "FlowEvents")
	graphanasession.AddRows(grafana.SingleStat, "events", "IPAddress", "ContainerEvents")

	zap.L().Info("Trireme-Statistics Started...")

	mux := http.NewServeMux()
	mux.HandleFunc("/get", server.GetData)
	mux.HandleFunc("/graph", server.GetGraph)
	//mux.Handle("/graph/", http.StripPrefix("/graph/", http.HandlerFunc(server.GetGraph)))

	handler := cors.Default().Handler(mux)
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, handler))

}
