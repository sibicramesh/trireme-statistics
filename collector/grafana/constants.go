package grafana

type PanelType string

const (
	SingleStat PanelType = "singlestat"
	Graph      PanelType = "graph"
	Diagram    PanelType = "jdbranham-diagram-panel"
)
