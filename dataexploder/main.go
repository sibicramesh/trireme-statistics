package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/aporeto-inc/trireme-statistics/influxdb"
	"github.com/aporeto-inc/trireme-statistics/models"
	"github.com/aporeto-inc/trireme/collector"
	"github.com/aporeto-inc/trireme/policy"
)

var wg sync.WaitGroup

func explode() {
	defer wg.Done()
	var flowModel models.FlowModel
	var contModel models.ContainerModel
	var source collector.EndPoint
	var destination collector.EndPoint
	samplesize := 1000
	counter := 0
	httpCli, _ := influxdb.NewDBConnection("aporeto", "aporeto", "http://influxdb:8086")
	httpCli.Start()
	for i := 0; i < samplesize; i++ {

		flowModel.FlowRecord.ContextID = "1ascasd7t"
		flowModel.FlowRecord.Count = counter
		flowModel.Counter = counter

		source.ID = "srcID"
		source.IP = "192.168.0.1"
		source.Port = 1234 + uint16(i)
		source.Type = collector.Address

		flowModel.FlowRecord.Source = &source

		destination.ID = "dstID"
		destination.IP = "192.1688.2.2"
		destination.Port = 880
		destination.Type = collector.Address

		flowModel.FlowRecord.Destination = &destination

		var tags policy.TagStore
		tags.Tags = []string{"server"}
		flowModel.FlowRecord.Tags = &tags
		var actype policy.ActionType
		actype.Accepted()
		actype.ActionString()
		flowModel.FlowRecord.Action = actype
		flowModel.FlowRecord.DropReason = "None"
		flowModel.FlowRecord.PolicyID = "sampleID"

		httpCli.CollectFlowEvent(&flowModel.FlowRecord)
		var policy policy.TagStore
		policy.Tags = []string{"@sys=name @app=web"}
		contModel.ContainerRecord.ContextID = "1ascasd7t"
		contModel.ContainerRecord.Event = "start"
		contModel.ContainerRecord.Tags = &policy
		httpCli.CollectContainerEvent(&contModel.ContainerRecord)
		counter++

	}
	wg.Wait()
	httpCli.Stop()
}

func main() {
	wg.Add(1)
	time.Sleep(time.Second * 5)
	go explode()
	wg.Wait()
	fmt.Println("Done main")
}
