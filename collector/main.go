package main

import (
	"fmt"

	"go.uber.org/zap"

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

	err = httlpcli.Start()
	if err != nil {
		fmt.Println(err)
		zap.L().Fatal("Failed to create Batch point", zap.Error(err))
	}

	httlpcli.AddToDB(1, map[string]interface{}{
		"Flow": "newFlow"})

	for {

	}
}
