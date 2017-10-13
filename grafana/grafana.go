package grafana

import (
	"fmt"
	"time"

	"github.com/aporeto-inc/grafanaclient"
)

// NewUI is used to create a new session and return grafana handle
func NewUI(user string, pass string, addr string) (Grafanaui, error) {
	if user == "" && pass == "" || addr == "" {
		session, err := createSession(username, password, url)
		if err != nil {
			return nil, err
		}
		return &Grafanauis{
			session: session,
		}, nil
	}
	session, err := createSession(user, pass, addr)
	if err != nil {
		return nil, err
	}
	return &Grafanauis{
		session: session,
	}, nil

}

func createSession(user string, pass string, addr string) (*grafanaclient.Session, error) {

	session := grafanaclient.NewSession(user, pass, addr)
	err := session.DoLogon()
	if err != nil {
		return nil, err
	}
	return session, nil
}

// func LaunchGrafanaCharts() (Grafanaui, error) {
// 	session, err := NewUI()
// 	if err != nil {
// 		return nil, err
// 	}
// 	ds, _ := session.GetDatasource("Dependency")
// 	if ds.Name != "Dependency" {
// 		err = session.CreateDataSource("Dependency")
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	session.CreateDashboard("")
//
// 	return session, nil
// }

// CreateDataSource is used to create a new datasource based on users arguements
func (g *Grafanauis) CreateDataSource(name string, dbname string, dbuname string, dbpass string, dburl string, access string) error {
	dsn, _ := g.GetDatasource(name)
	if dsn.Name != name {
		if dbname == "" {
			ds := grafanaclient.DataSource{Name: name,
				Type:     "influxdb",
				Access:   access,
				URL:      dburl,
				User:     dbuname,
				Password: dbpass,
				Database: database,
			}

			err := g.session.CreateDataSource(ds)
			if err != nil {
				return err
			}
		} else {
			ds := grafanaclient.DataSource{Name: name,
				Type:     "influxdb",
				Access:   access,
				URL:      dburl,
				User:     dbuname,
				Password: dbpass,
				Database: database,
			}

			err := g.session.CreateDataSource(ds)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ListDataSources is used to list all available sources
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

// GetDatasource is used to find and return a specific datasource, error otherwise
func (g *Grafanauis) GetDatasource(name string) (*grafanaclient.DataSource, error) {
	ds, err := g.session.GetDataSource(name)
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

// GetDashboard is used to find and return a specific dashboard, error otherwise
func (g *Grafanauis) GetDashboard(name string) (grafanaclient.DashboardResult, error) {
	dr, err := g.session.GetDashboard(name)
	if err != nil {
		return grafanaclient.DashboardResult{}, err
	}
	return dr, nil
}

// CreateDashboard is used to create a new dashboard
func (g *Grafanauis) CreateDashboard(dbr string) {
	dashboard := grafanaclient.Dashboard{Editable: true}
	if dbr == "" {
		dashboard.Title = "DependencyBoard"
	} else {
		dashboard.Title = dbr
	}
	g.dashboard = &dashboard
}

// AddRows is used to add a new row in the dashboard
func (g *Grafanauis) AddRows(panel PanelType, rowname string, fields string, events string) {

	graphRow := grafanaclient.NewRow()
	graphRow.Title = rowname
	//graphRow.Collapse = true // it will be collapsed by default
	g.row = graphRow

	newpanel := g.AddCharts(panel, events, fields)

	g.AddPanels(newpanel)

	g.dashboard.AddRow(g.row)

	g.UploadToDashboard()
}

// AddPanels is a wrapper around addpanel to add a new panel to a row or create just a row
func (g *Grafanauis) AddPanels(newpanel grafanaclient.Panel) {

	g.row.AddPanel(newpanel)

}

// UploadToDashboard is used to push all created rows into the dashboard
func (g *Grafanauis) UploadToDashboard() {

	g.dashboard.SetTimeFrame(time.Now().Add(-5*time.Minute), time.Now().Add(10*time.Minute))

	g.session.UploadDashboard(*g.dashboard, true)
}

// AddCharts is used to add different charts into rows
func (g *Grafanauis) AddCharts(paneltype PanelType, paneltitle string, fields string) grafanaclient.Panel {

	// NewPanel will create a graph panel by default
	graphPanel := grafanaclient.NewPanel()

	// set panel title
	graphPanel.Title = paneltitle
	if paneltype == "singlestat" {
		graphPanel.Type = "singlestat"
		graphPanel.DataSource = "Events"
		// } else if paneltype == "graph" {
		// 	graphPanel.Type = "graph"
		// 	graphPanel.DataSource = "Dependency"
		// } else if paneltype == "jdbranham-diagram-panel" {
		// 	graphPanel.Type = "jdbranham-diagram-panel"
		// 	legend := grafanaclient.NewLegend()
		// 	legend.Gradient = []string{""}
		// 	graphPanel.Legend = legend
		// 	graphPanel.DataSource = "Dependency"
	}
	// let's specify the datasource

	graphPanel.ValueName = "total"

	// change panel span from default 12 to 6
	graphPanel.Span = 12

	// stack lines with a filling of 1
	graphPanel.Stack = true
	graphPanel.Fill = 1
	// define a target
	target := grafanaclient.NewTarget()

	//specify the measurement to use
	if paneltitle == "FlowEvents" {
		target.Measurement = "FlowEvents"
	} else if paneltitle == "ContainerEvents" {
		target.Measurement = "ContainerEvents"
	}
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

// CreateGraphs is used to create new graphs without adding any rows
func (g *Grafanauis) CreateGraphs(panel PanelType, rowname string, fields string, events string) {
	newpanel := g.AddCharts(panel, events, fields)

	g.AddPanels(newpanel)

	g.UploadToDashboard()
}
