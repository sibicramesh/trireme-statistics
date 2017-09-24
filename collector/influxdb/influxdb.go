package influxdb

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	tcollector "github.com/aporeto-inc/trireme/collector"
	"github.com/influxdata/influxdb/client/v2"
)

const (
	database = "flowDB"
	username = "aporeto"
	password = "aporeto"
)

func NewDB() (Influxdb, error) {

	httpClient, err := CreateHTTPClient()
	if err != nil {
		return nil, err
	}

	return &Influxdbs{
		httpClient:  httpClient,
		reportFlows: make(chan map[string]interface{}),
		stop:        make(chan bool),
		doneAdding:  make(chan bool),
		count:       make(chan int),
	}, nil
}

func CreateHTTPClient() (client.Client, error) {
	httpClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://influxdb:8086",
		Username: username,
		Password: password,
	})

	if err != nil {
		return nil, err
	}

	return httpClient, nil
}

func (d *Influxdbs) CreateDB() error {
	q := client.NewQuery("CREATE DATABASE "+database, "", "")
	if response, err := d.httpClient.Query(q); err != nil && response.Error() != nil {

		return err
	}

	return nil
}

func (d *Influxdbs) AddToDB(value int, tags map[string]interface{}) error {

	if tags != nil {
		d.reportFlows <- tags
		d.count <- value
		if <-d.doneAdding {
			err := d.httpClient.Write(d.batchPoint)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Influxdbs) Start() error {

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "us",
	})
	if err != nil {
		return err
	}

	d.batchPoint = bp
	go d.listen(d.batchPoint)

	return nil
}

func (d *Influxdbs) Stop() error {
	<-d.stop
	d.httpClient.Close()
	return nil
}

func (d *Influxdbs) listen(bp client.BatchPoints) {

	for {
		select {
		case r := <-d.reportFlows:
			go func(r map[string]interface{}) {
				d.AddData(bp, <-d.count, r)
			}(r)
		case <-d.stop:
			return
		default:
		}
	}
}

func (d *Influxdbs) AddData(bp client.BatchPoints, value int, fields map[string]interface{}) {

	tag := map[string]string{"counter": "flowstats"}

	pt, err := client.NewPoint("flows", tag, fields, time.Now())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pt)
	zap.L().Info("hi")
	bp.AddPoint(pt)
	d.doneAdding <- true

}

func (d *Influxdbs) CollectFlowEvent(record *tcollector.FlowRecord) {
	d.AddToDB(record.Count, map[string]interface{}{
		"ContextID":       record.ContextID,
		"Counter":         record.Count,
		"SourceID":        record.Source.ID,
		"SourceIP":        record.Source.IP,
		"SourcePort":      record.Source.Port,
		"SourceType":      record.Source.Type,
		"DestinationID":   record.Destination.ID,
		"DestinationIP":   record.Destination.IP,
		"DestinationPort": record.Destination.Port,
		"DestinationType": record.Destination.Type,
		"Action":          record.Action,
		"DropReason":      record.DropReason,
		"PolicyID":        record.PolicyID,
	})
}

func (d *Influxdbs) CollectContainerEvent(record *tcollector.ContainerRecord) {
	d.AddToDB(1, map[string]interface{}{
		"ContextID": record.ContextID,
		"IPAddress": record.IPAddress,
		"Tags":      record.Tags,
		"Event":     record.Event,
	})
}
