package main

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/aporeto-inc/trireme-statistics/collector/grafana"
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

	for {

	}

}
