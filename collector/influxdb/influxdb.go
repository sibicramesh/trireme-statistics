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
		tags:        make(chan map[string]string),
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

func CreateAndStartDB() *Influxdbs {
	httlpcli, err := NewDB()
	if err != nil {
		zap.L().Fatal("Failed to connect", zap.Error(err))
	}

	err = httlpcli.CreateDB()
	if err != nil {
		fmt.Println(err)
		zap.L().Fatal("Failed to create DB", zap.Error(err))
	}

	httlpcli.Start()

	return httlpcli
}

func (d *Influxdbs) CreateDB() error {
	q := client.NewQuery("CREATE DATABASE "+database, "", "")
	if response, err := d.httpClient.Query(q); err != nil && response.Error() != nil {

		return err
	}

	return nil
}

func (d *Influxdbs) AddToDB(tags map[string]string, fields map[string]interface{}) error {

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

	// d.grafana, err = grafana.LaunchGrafanaCharts()
	// if err != nil {
	// 	return err
	// }

	go d.listen()

	return nil
}

func (d *Influxdbs) Stop() error {
	<-d.stop
	d.httpClient.Close()
	return nil
}

func (d *Influxdbs) listen() {

	for {
		select {
		case r := <-d.reportFlows:
			go func(r map[string]interface{}) {
				d.AddData(<-d.tags, r)
			}(r)
		case <-d.stop:
			return
		default:
		}
	}
}

func (d *Influxdbs) AddData(tags map[string]string, fields map[string]interface{}) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "us",
	})
	if err != nil {

	}
	if tags["EventName"] == "ContainerStartEvents" || tags["EventName"] == "ContainerStopEvents" {
		pt, err := client.NewPoint("ContainerEvents", tags, fields, time.Now())
		if err != nil {
			fmt.Println(err)
		}
		zap.L().Info(pt.String())
		bp.AddPoint(pt)
	} else if tags["EventName"] == "FlowEvents" {
		pt, err := client.NewPoint("FlowEvents", tags, fields, time.Now())
		if err != nil {
			fmt.Println(err)
		}
		zap.L().Info(pt.String())
		bp.AddPoint(pt)
	}
	d.batchPoint = bp
	d.doneAdding <- true
}

func (d *Influxdbs) CollectFlowEvent(record *tcollector.FlowRecord) {
	//	cid, _ := d.cache.Get(record.ContextID)
	//if record.ContextID == cid {
	//d.grafana.CreateGraphs(grafana.Graph, "events", "Action", "FlowEvents")

	d.AddToDB(map[string]string{
		"EventName": "FlowEvents",
		"EventID":   record.ContextID,
	}, map[string]interface{}{
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
	//} else {
	//fmt.Println("NO PU RUNNING BUT RECEIVED A FLOW")
	//}
}

func (d *Influxdbs) CollectContainerEvent(record *tcollector.ContainerRecord) {
	if record.Event == "start" {
		//d.cache.Add(record.ContextID, record.ContextID)
		//d.contextID = record.ContextID
		//d.grafana.AddRows(grafana.Graph, "events", "Action", "FlowEvents")
		d.AddToDB(map[string]string{
			"EventName": "ContainerStartEvents",
			"EventID":   record.ContextID,
		}, map[string]interface{}{
			"ContextID": record.ContextID,
			"IPAddress": record.IPAddress,
			"Tags":      record.Tags,
			"Event":     record.Event,
		})
	}
	if record.Event == "delete" {
		//d.cache.Add(record.ContextID, record.ContextID)
		//d.contextID = record.ContextID
		//d.grafana.AddRows(grafana.Graph, "events", "Action", "FlowEvents")
		d.AddToDB(map[string]string{
			"EventName": "ContainerStopEvents",
			"EventID":   record.ContextID,
		}, map[string]interface{}{
			"ContextID": record.ContextID,
			"IPAddress": record.IPAddress,
			"Tags":      record.Tags,
			"Event":     record.Event,
		})
	}
}
