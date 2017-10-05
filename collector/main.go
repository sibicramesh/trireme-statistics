package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/aporeto-inc/trireme-statistics/collector/grafana"
	"github.com/aporeto-inc/trireme-statistics/collector/graph/utils"
)

func banner() {
	fmt.Printf(`
Trireme-Stats
`)
}

func main() {
	banner()

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

	http.HandleFunc("/get", utils.GetData)

	http.Handle("/graph/", http.StripPrefix("/graph/", http.FileServer(http.Dir("graph"))))

	log.Fatal(http.ListenAndServe(":8080", nil))

}
