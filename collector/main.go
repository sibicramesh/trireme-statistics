package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/aporeto-inc/trireme-statistics/collector/grafana"
	"github.com/aporeto-inc/trireme-statistics/collector/graph/utils"
	"github.com/aporeto-inc/trireme-statistics/collector/influxdb"
)

func banner() {
	fmt.Printf(`
Trireme-Stats
`)
}

func main() {
	banner()

	httlpcli, err := influxdb.NewDB()
	if err != nil {
		zap.L().Fatal("Failed to connect", zap.Error(err))
	}

	err = httlpcli.CreateDB()
	if err != nil {
		fmt.Println(err)
		zap.L().Fatal("Failed to create DB", zap.Error(err))
	}
	time.Sleep(time.Second * 10)
	graphanasession, err := grafana.NewUI()
	if err != nil {
		zap.L().Fatal("Failed to connect", zap.Error(err))
	}

	err = graphanasession.CreateDataSource("Events")
	if err != nil {
		fmt.Println(err)
		zap.L().Fatal("Failed to create datasource", zap.Error(err))
	}

	graphanasession.CreateDashboard("StatisticBoard")
	graphanasession.AddRows(grafana.SingleStat, "events", "Action", "FlowEvents")
	graphanasession.AddRows(grafana.SingleStat, "events", "IPAddress", "ContainerEvents")

	zap.L().Info("Database created and ready to be consumed")

	router := mux.NewRouter()

	router.HandleFunc("/get", utils.GetData).Methods("GET")
	router.HandleFunc("/put", utils.PostData).Methods("POST")
	router.Handle("/graph", http.FileServer(http.Dir("./graph")))

	log.Fatal(http.ListenAndServe(":8080", router))

}
