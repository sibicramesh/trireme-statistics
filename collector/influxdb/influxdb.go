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

// NewDB is used to create a new client and return influxdb handle
func NewDB(user string, pass string, addr string) (*Influxdbs, error) {
	httpClient, err := createHTTPClient(user, pass, addr)
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

func createHTTPClient(user string, pass string, addr string) (client.Client, error) {
	if user == "" && pass == "" || addr == "" {
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
	httpClient, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: user,
		Password: pass,
	})

	if err != nil {
		return nil, err
	}
	return httpClient, nil

}

// CreateAndConnectDB is used by the caller to connect with th ecurrent instance and start collecting
func CreateAndConnectDB(user string, pass string, addr string) *Influxdbs {
	httlpcli, err := NewDB(user, pass, addr)
	if err != nil {
		zap.L().Fatal("Failed to connect", zap.Error(err))
	}

	httlpcli.Start()

	return httlpcli
}

// CreateDB is used to create a new databases given name
func (d *Influxdbs) CreateDB(dbname string) error {
	if dbname == "" {
		q := client.NewQuery("CREATE DATABASE "+database, "", "")
		if response, err := d.httpClient.Query(q); err != nil && response.Error() != nil {

			return err
		}
	} else {
		q := client.NewQuery("CREATE DATABASE "+dbname, "", "")
		if response, err := d.httpClient.Query(q); err != nil && response.Error() != nil {

			return err
		}
	}
	return nil
}

// AddToDB is used to add points to the data as batches
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

// Start is used to start listening for data
func (d *Influxdbs) Start() error {

	// d.grafana, err = grafana.LaunchGrafanaCharts()
	// if err != nil {
	// 	return err
	// }

	go d.listen()

	return nil
}

// Stop is used to stop and return from listen goroutine
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

// AddData is used to add data to the batch
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

// CollectFlowEvent implements trireme collector interface
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

// CollectContainerEvent implements trireme collector interface
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
