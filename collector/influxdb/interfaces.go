package influxdb

import (
	"github.com/aporeto-inc/trireme-statistics/collector/cache"
	"github.com/aporeto-inc/trireme-statistics/collector/grafana"
	"github.com/influxdata/influxdb/client/v2"
)

type Influxdb interface {
	CreateDB() error
	AddToDB(value int, tags map[string]interface{}) error
	AddData(bp client.BatchPoints, value int, tags map[string]interface{})
	Start() error
	Stop() error
}

type Influxdbs struct {
	httpClient  client.Client
	batchPoint  client.BatchPoints
	tags        chan string
	reportFlows chan (map[string]interface{})
	stop        chan bool
	doneAdding  chan bool
	cache       cache.Cache
	grafana     grafana.Grafanaui
}
