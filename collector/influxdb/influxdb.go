package influxdb

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/aporeto-inc/trireme-statistics/collector/cache"
	"github.com/aporeto-inc/trireme-statistics/collector/grafana"
	tcollector "github.com/aporeto-inc/trireme/collector"
	"github.com/influxdata/influxdb/client/v2"
)

const (
	database = "flowDB"
	username = "aporeto"
	password = "aporeto"
)

func NewDB() (*Influxdbs, error) {

	httpClient, err := CreateHTTPClient()
	if err != nil {
		return nil, err
	}

	return &Influxdbs{
		httpClient:  httpClient,
		reportFlows: make(chan map[string]interface{}),
		stop:        make(chan bool),
		doneAdding:  make(chan bool),
		tags:        make(chan string),
		cache:       cache.NewCache(),
	}, nil
}

func CreateHTTPClient() (client.Client, error) {
	httpClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://0.0.0.0:8086",
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

func (d *Influxdbs) AddToDB(tags string, fields map[string]interface{}) error {

	if fields != nil {
		d.reportFlows <- fields
		d.tags <- tags
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

	d.grafana, err = grafana.LaunchGrafanaCharts()
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
				d.AddData(bp, <-d.tags, r)
			}(r)
		case <-d.stop:
			return
		default:
		}
	}
}

func (d *Influxdbs) AddData(bp client.BatchPoints, tags string, fields map[string]interface{}) {

	tag := map[string]string{"tag": tags}

	pt, err := client.NewPoint("flows", tag, fields, time.Now())
	if err != nil {
		fmt.Println(err)
	}
	zap.L().Info(pt.String())
	bp.AddPoint(pt)
	d.doneAdding <- true
}

func (d *Influxdbs) CollectFlowEvent(record *tcollector.FlowRecord) {
	cid, _ := d.cache.Get("ContextID")

	d.grafana.AddRows("events", "ContextID", "FlowEvents")
	if record.ContextID == cid {

		d.AddToDB("FlowEvents", map[string]interface{}{
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
	} else {
		fmt.Println("NO PU RUNNING BUT RECEIVED A FLOW")
	}
}

func (d *Influxdbs) CollectContainerEvent(record *tcollector.ContainerRecord) {
	//if record.Event == "Create" {
	d.cache.Add("ContextID", record.ContextID)
	//}

	d.AddToDB("ContainerEvents", map[string]interface{}{
		"ContextID": record.ContextID,
		"IPAddress": record.IPAddress,
		"Tags":      record.Tags,
		"Event":     record.Event,
	})
}
