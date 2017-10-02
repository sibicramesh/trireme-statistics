package grafana

import (
	"github.com/sibicramesh/grafanaclient"
)

type Grafanaui interface {
	CreateDataSource(name string) error
	ListDataSources() error
	CreateDashboard(dbr string)
	GetDatasource(name string) (*grafanaclient.DataSource, error)
	AddCharts(panel PanelType, title string, fields string) grafanaclient.Panel
	AddRows(panel PanelType, rowname string, paneltitle string, events string)
	GetDashboard(name string) error
	UploadToDashboard()
}
