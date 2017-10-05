package influxdb

import (
	"github.com/aporeto-inc/trireme-statistics/collector/grafana"
	"github.com/influxdata/influxdb/client/v2"
)

type Influxdbs struct {
	httpClient  client.Client
	batchPoint  client.BatchPoints
	tags        chan (map[string]string)
	reportFlows chan (map[string]interface{})
	stop        chan bool
	doneAdding  chan bool
	grafana     grafana.Grafanaui
	contextID   string
}
