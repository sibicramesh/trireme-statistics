package grafana

import (
	"github.com/aporeto-inc/grafanaclient"
)

type Grafanauis struct {
	session   *grafanaclient.Session
	dashboard *grafanaclient.Dashboard
	row       grafanaclient.Row
}
