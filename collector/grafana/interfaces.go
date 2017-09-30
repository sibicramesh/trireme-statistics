package grafana

import (
	"github.com/adejoux/grafanaclient"
)

type Grafanaui interface {
	CreateDataSource() error
	ListDataSources() error
	CreateDashboard(dbr string)
	AddCharts(title string, fields string) grafanaclient.Panel
	AddRows(rowname string, paneltitle string, events string)
	GetDashboard(name string) error
	UploadToDashboard()
}

type Grafanauis struct {
	session   *grafanaclient.Session
	dashboard *grafanaclient.Dashboard
	row       grafanaclient.Row
}
