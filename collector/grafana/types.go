package grafana

import (
	"github.com/sibicramesh/grafanaclient"
)

type Grafanauis struct {
	session   *grafanaclient.Session
	dashboard *grafanaclient.Dashboard
	row       grafanaclient.Row
}
