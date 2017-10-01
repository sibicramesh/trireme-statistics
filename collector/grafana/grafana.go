package grafana

import (
	"fmt"
	"time"

	"github.com/adejoux/grafanaclient"
)

const (
	url      = "http://0.0.0.0:3000"
	username = "admin"
	password = "admin"
)

func NewUI() (Grafanaui, error) {
	session, err := CreateSession()
	return &Grafanauis{
		session: session,
	}, err
}

func CreateSession() (*grafanaclient.Session, error) {
	session := grafanaclient.NewSession(username, password, url)
	err := session.DoLogon()
	if err != nil {
		return nil, err
	}
	return session, nil
}

func LaunchGrafanaCharts() (Grafanaui, error) {
	session, err := NewUI()
	if err != nil {
		return nil, err
	}
	// err = session.CreateDataSource()
	// if err != nil {
	// 	return nil, err
	// }
	session.CreateDashboard("")

	return session, nil
}

func (g *Grafanauis) CreateDataSource() error {
	ds := grafanaclient.DataSource{Name: "Events",
		Type:     "influxdb",
		Access:   "direct",
		URL:      "http://0.0.0.0:8086",
		User:     "aporeto",
		Password: "aporeto",
		Database: "flowDB",
	}

	err := g.session.CreateDataSource(ds)
	if err != nil {
		return err
	}
	return nil
}

func (g *Grafanauis) ListDataSources() error {
	dss, err := g.session.GetDataSourceList()
	if err != nil {
		return err
	}

	for _, ds := range dss {
		fmt.Printf("name: %s type: %s url: %s\n", ds.Name, ds.Type, ds.URL)
	}

	return nil
}

func (g *Grafanauis) GetDashboard(name string) error {
	dr, err := g.session.GetDashboard(name)
	if err != nil {
		return err
	}
	fmt.Println(dr)
	return nil
}

func (g *Grafanauis) CreateDashboard(dbr string) {
	dashboard := grafanaclient.Dashboard{Editable: true}
	g.dashboard = &dashboard
	if dbr == "" {
		dashboard.Title = "Dependency"
	} else {
		dashboard.Title = dbr
	}
}

func (g *Grafanauis) AddRows(rowname string, fields string, events string) {

	graphRow := grafanaclient.NewRow()
	graphRow.Title = rowname
	//graphRow.Collapse = true // it will be collapsed by default
	g.row = graphRow
	newpanel := g.AddCharts(events, fields)

	g.AddPanels(newpanel)

	g.dashboard.AddRow(g.row)

	g.UploadToDashboard()
}

func (g *Grafanauis) AddPanels(newpanel grafanaclient.Panel) {

	g.row.AddPanel(newpanel)

}

func (g *Grafanauis) UploadToDashboard() {

	g.dashboard.SetTimeFrame(time.Now().Add(-5*time.Minute), time.Now().Add(10*time.Minute))

	g.session.UploadDashboard(*g.dashboard, true)
}

func (g *Grafanauis) AddCharts(paneltitle string, fields string) grafanaclient.Panel {

	// NewPanel will create a graph panel by default
	graphPanel := grafanaclient.NewPanel()

	// set panel title
	graphPanel.Title = paneltitle

	// let's specify the datasource
	graphPanel.DataSource = "Events"

	// change panel span from default 12 to 6
	graphPanel.Span = 12

	// stack lines with a filling of 1
	graphPanel.Stack = true
	graphPanel.Fill = 1

	// define a target
	target := grafanaclient.NewTarget()

	//specify the measurement to use
	target.Measurement = "flows"
	var selectd grafanaclient.Select
	selectd.Type = "field"
	selectd.Params = []string{fields}

	var selectcount grafanaclient.Select
	selectcount.Type = "count"

	var selects grafanaclient.Selects
	selects = append(selects, selectd)
	selects = append(selects, selectcount)

	// var selectcd grafanaclient.Selects
	// selectcd = append(selectcd, selectcount)

	var selectarr []grafanaclient.Selects
	selectarr = append(selectarr, selects)
	//	selectarr = append(selectarr, selectcd)
	target.Select = selectarr
	target.Alias = fields

	// Adding everything
	graphPanel.AddTarget(target)

	return graphPanel

}
