package influxdb

import (
	"github.com/influxdata/influxdb/client/v2"
)

type Influxdb interface {
	CreateDB() error
	AddToDB(tags map[string]string, fields map[string]interface{}) error
	AddData(bp client.BatchPoints, value int, tags map[string]interface{})
	Start() error
	Stop() error
}
