package models

import "github.com/aporeto-inc/trireme/collector"

type FlowModel struct {
	Counter    int
	FlowRecord collector.FlowRecord
}

type ContainerModel struct {
	Counter         int
	ContainerRecord collector.ContainerRecord
}
