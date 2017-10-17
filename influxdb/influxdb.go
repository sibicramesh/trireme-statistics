package influxdb

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	tcollector "github.com/aporeto-inc/trireme/collector"
	"github.com/influxdata/influxdb/client/v2"
)

// TODO: Remove default credentials
const (
	database = "flowDB"
	username = "aporeto"
	password = "aporeto"
)

//Influxdb inplements a DataAdder interface for influxDB
type Influxdb struct {
	httpClient client.Client

	stopWorker chan struct{}
	worker     *worker
	done       chan bool
}

//DataAdder interface has all the methods required to interact with influxdb api
type DataAdder interface {
	CreateDB(string) error
	AddData(tags map[string]string, fields map[string]interface{}) error
}

// NewDBConnection is used to create a new client and return influxdb handle
func NewDBConnection(user string, pass string, addr string) (*Influxdb, error) {
	zap.L().Debug("Initializing InfluxDBConnection")
	httpClient, err := createHTTPClient(user, pass, addr)
	if err != nil {
		return nil, fmt.Errorf("Error parsing url %s", err)
	}
	_, _, err = httpClient.Ping(time.Second * 0)
	if err != nil {
		return nil, fmt.Errorf("Unable to create InfluxDB http client %s", err)
	}

	dbConnection := &Influxdb{
		httpClient: httpClient,
		stopWorker: make(chan struct{}),
		done:       make(chan bool, 100),
	}

	worker := newWorker(dbConnection.stopWorker, dbConnection)
	dbConnection.worker = worker

	return dbConnection, nil
}

func createHTTPClient(user string, pass string, addr string) (client.Client, error) {

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

// CreateDB is used to create a new databases given name
func (d *Influxdb) CreateDB(dbname string) error {
	zap.L().Info("Creating database", zap.String("db", dbname))

	_, err := d.ExecuteQuery("CREATE DATABASE "+dbname, "")
	if err != nil {
		return err
	}

	return nil
}

// ExecuteQuery is used to execute a query given a database name
func (d *Influxdb) ExecuteQuery(query string, dbname string) (*client.Response, error) {

	q := client.NewQuery(query, dbname, "")
	response, err := d.httpClient.Query(q)
	if err != nil && response.Error() != nil {
		return nil, err
	}

	return response, nil
}

// Start is used to start listening for data
func (d *Influxdb) Start() error {
	zap.L().Info("Starting InfluxDB worker")

	go d.worker.startWorker()

	return nil
}

// Stop is used to stop and return from listen goroutine
func (d *Influxdb) Stop() error {
	zap.L().Info("Stopping InfluxDB worker")

	d.stopWorker <- struct{}{}
	d.httpClient.Close()

	return nil
}

// AddData is used to add data to the batch
func (d *Influxdb) AddData(tags map[string]string, fields map[string]interface{}) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "us",
	})
	if err != nil {
		return fmt.Errorf("Couldn't add data, error creating batchpoint: %s", err)
	}

	if tags["EventName"] == "ContainerStartEvents" || tags["EventName"] == "ContainerStopEvents" {
		pt, err := client.NewPoint("ContainerEvents", tags, fields, time.Now())
		if err != nil {
			return fmt.Errorf("Couldn't add ContainerEvent: %s", err)
		}
		bp.AddPoint(pt)
	} else if tags["EventName"] == "FlowEvents" {
		pt, err := client.NewPoint("FlowEvents", tags, fields, time.Now())
		if err != nil {
			return fmt.Errorf("Couldn't add FlowEvent: %s", err)
		}
		bp.AddPoint(pt)
	}
	if err := d.httpClient.Write(bp); err != nil {
		return fmt.Errorf("Couldn't add data: %s", err)
	}

	if <-d.done {
		return nil
	}

	return nil
}

// CollectFlowEvent implements trireme collector interface
func (d *Influxdb) CollectFlowEvent(record *tcollector.FlowRecord) {
	go d.AddData(map[string]string{
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
	d.done <- true
}

// CollectContainerEvent implements trireme collector interface
func (d *Influxdb) CollectContainerEvent(record *tcollector.ContainerRecord) {
	var eventName string

	switch record.Event {
	case "start", "update", "create":
		eventName = "ContainerStartEvents"

	case "delete":
		eventName = "ContainerStopEvents"
	default:
		zap.L().Error("Unrecognized container event name ")

	}
	go d.AddData(map[string]string{
		"EventName": eventName,
		"EventID":   record.ContextID,
	}, map[string]interface{}{
		"ContextID": record.ContextID,
		"IPAddress": record.IPAddress,
		"Tags":      record.Tags,
		"Event":     record.Event,
	})
	d.done <- true
}
