package grafana

import (
	"github.com/sibicramesh/grafanaclient"
)

type Grafanaui interface {
	CreateDataSource() error
	ListDataSources() error
	CreateDashboard(dbr string)
	AddCharts(panel PanelType, title string, fields string) grafanaclient.Panel
	AddRows(panel PanelType, rowname string, paneltitle string, events string)
	GetDashboard(name string) error
	UploadToDashboard()
}
