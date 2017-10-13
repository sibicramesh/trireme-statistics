package influxdb

import (
	"fmt"
	"time"

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

	worker     *worker
	stopWorker chan struct{}
}

//DataAdder interface has all the methods required to interact with influxdb api
type DataAdder interface {
	CreateDB() error
	AddData(tags map[string]string, fields map[string]interface{}) error
}

// NewNewDBConnectionDB is used to create a new client and return influxdb handle
func NewNewDBConnectionDB(user string, pass string, addr string) (*Influxdb, error) {
	httpClient, err := createHTTPClient(user, pass, addr)
	if err != nil {
		return nil, fmt.Errorf("Unable to create InfluxDB http client %s", err)
	}

	return &Influxdb{
		httpClient: httpClient,
	}, nil
}

func createHTTPClient(user string, pass string, addr string) (client.Client, error) {

	// TODO: Remove this.
	if user == "" && pass == "" || addr == "" {
		addr = "http://influxdb:8086"
		user = username
		pass = password
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

// CreateDB is used to create a new databases given name
func (d *Influxdb) CreateDB(dbname string) error {
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

// Start is used to start listening for data
func (d *Influxdb) Start() error {
	go d.worker.startWorker()

	return nil
}

// Stop is used to stop and return from listen goroutine
func (d *Influxdb) Stop() error {
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
		return fmt.Errorf("Couldn't add data: %s", err)
	}

	if tags["EventName"] == "ContainerStartEvents" || tags["EventName"] == "ContainerStopEvents" {
		pt, err := client.NewPoint("ContainerEvents", tags, fields, time.Now())
		if err != nil {
			fmt.Println(err)
		}
		bp.AddPoint(pt)
	} else if tags["EventName"] == "FlowEvents" {
		pt, err := client.NewPoint("FlowEvents", tags, fields, time.Now())
		if err != nil {
			fmt.Println(err)
		}
		bp.AddPoint(pt)
	}
	if err := d.httpClient.Write(bp); err != nil {
		return fmt.Errorf("Couldn't add data: %s", err)
	}
	return nil
}

// CollectFlowEvent implements trireme collector interface
func (d *Influxdb) CollectFlowEvent(record *tcollector.FlowRecord) {
	d.worker.addEvent(
		&workerEvent{
			event:      flowEvent,
			flowRecord: record,
		},
	)
}

// CollectContainerEvent implements trireme collector interface
func (d *Influxdb) CollectContainerEvent(record *tcollector.ContainerRecord) {
	d.worker.addEvent(
		&workerEvent{
			event:           containerEvent,
			containerRecord: record,
		},
	)
}
