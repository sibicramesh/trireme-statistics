package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

// GraphData is the struct that holds the json format required for graph to generate nodes and link
type GraphData struct {
	Nodes []Nodes `json:"nodes"`
	Links []Links `json:"links"`
}

// Nodes which holds pu information
type Nodes struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IPAddress string `json:"ipaddress"`
}

// Links which holds the links between pu's
type Links struct {
	Source int    `json:"source"`
	Target int    `json:"target"`
	Action string `json:"action"`
}

// InfluxData is the struct that holds the data returned from influxdb api
type InfluxData struct {
	Results []struct {
		StatementID int `json:"statement_id"`
		Series      []struct {
			Name    string          `json:"name"`
			Columns []string        `json:"columns"`
			Values  [][]interface{} `json:"values"`
		} `json:"series"`
	} `json:"results"`
}

// GetData is called by the client which generates json with a logic that defines the nodes and links
func GetData(w http.ResponseWriter, r *http.Request) {

	body, res, err := getContainerEvents()
	if err != nil {
		http.Error(w, err.Error(), 0)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		http.Error(w, err.Error(), 1)
	}

	jsonData, err := transform(res)
	if err != nil {
		http.Error(w, err.Error(), 2)
	}

	err = json.NewEncoder(w).Encode(jsonData)
	if err != nil {
		http.Error(w, err.Error(), 3)
	}
}

// GetGraph is used to parse html with custom address to request for json
func GetGraph(w http.ResponseWriter, r *http.Request) {

	htmlData, err := template.New("graph").Parse(js)
	if err != nil {
		http.Error(w, err.Error(), 0)
	}

	data := struct {
		Address string
	}{
		Address: r.URL.Query().Get("address"),
	}

	err = htmlData.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), 1)
	}

	w.Header().Set("Content-Type", "text/html")
}

func getContainerEvents() ([]byte, *InfluxData, error) {

	body, res, err := getByURI("http://influxdb:8086/query?db=flowDB&&q=SELECT%20*%20FROM%20ContainerEvents")
	if err != nil {
		return nil, nil, fmt.Errorf("Error: Resource Unavailabe %s", err)
	}

	return body, res, nil
}

func getFlowEvents() ([]byte, *InfluxData, error) {

	body, res, err := getByURI("http://influxdb:8086/query?db=flowDB&&q=SELECT%20*%20FROM%20FlowEvents")
	if err != nil {
		return nil, nil, fmt.Errorf("Error: Resource Unavailabe %s", err)
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, nil, fmt.Errorf("Error: Malformed JSON Data %s", err)
	}

	return body, res, nil
}

func getByURI(uri string) ([]byte, *InfluxData, error) {
	var res InfluxData

	resp, err := http.Get(uri)
	if err != nil {
		return nil, nil, fmt.Errorf("Error: Resource Unavailabe %s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("Error: Reading from Resoponse %s", err)
	}

	return body, &res, nil
}

func deleteContainerEvents(id []string) ([]Nodes, error) {
	var node Nodes
	var nodes []Nodes

	for i := 0; i < len(id); i++ {
		_, err := http.Get("http://influxdb:8086/query?db=flowDB&q=DELETE%20FROM%20%22ContainerEvents%22%20WHERE%20(%22EventID%22%20=%20%27" + id[i] + "%27)")
		if err != nil {
			return nil, fmt.Errorf("Error: Reading from Resoponse %s", err)
		}
	}

	body, res, _ := getContainerEvents()

	err := json.Unmarshal(body, &res)
	if err != nil {
		return nil, fmt.Errorf("Error: Malformed JSON Data %s", err)
	}

	for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
		node.ID = res.Results[0].Series[0].Values[j][1].(string)
		if res.Results[0].Series[0].Values[j][6].(string) != "" {
			name := getName(res.Results[0].Series[0].Values[j][6].(string))
			node.Name = name
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// transform will convert the JSON response from influxdb to nodes and links to generate graph
// nodes struct will have nodeid, nodeipaddress and nodename
// links struct will have source, target and action
// the nodes are extracted from the influx data and stored in the array of structure
// then later this array is sent to the link generator which process the links between the nodes
// the link generator basically generates the link by comparing the nodeip with the flows src and dst ip's
func transform(res *InfluxData) (*GraphData, error) {
	var nodes []Nodes
	var links []Links
	var node Nodes
	var err error
	var id []string
	var startEvents = []string{"start", "update", "create"}

	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == "ContainerEvents" {
			for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
				if res.Results[0].Series[0].Values[j][2].(string) != "delete" {
					for k := 0; k < len(startEvents); k++ {
						if res.Results[0].Series[0].Values[j][2].(string) == startEvents[k] {
							node.ID = res.Results[0].Series[0].Values[j][1].(string)
							node.IPAddress = res.Results[0].Series[0].Values[j][5].(string)
							if res.Results[0].Series[0].Values[j][6].(string) != "" {
								name := getName(res.Results[0].Series[0].Values[j][6].(string))
								node.Name = name
							}
							nodes = append(nodes, node)
						}
					}
				} else {
					id = append(id, res.Results[0].Series[0].Values[j][1].(string))
					nodes, err = deleteContainerEvents(id)
					if err != nil {
						return nil, fmt.Errorf("Error: Reading from Resoponse %s", err)
					}
				}
			}
		}
	}

	links, err = generateLinks(nodes)
	if err != nil {
		return nil, fmt.Errorf("Error: Generating Links %s", err)
	}

	jsonData := GraphData{Nodes: nodes, Links: links}

	return &jsonData, nil
}

func generateLinks(nodea []Nodes) ([]Links, error) {

	_, res, err := getFlowEvents()
	if err != nil {
		return nil, fmt.Errorf("Error: Flow Events not received %s", err)
	}

	var links []Links
	var link Links
	var isSrc, isDst bool

	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == "FlowEvents" {
			for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
				for i := 0; i < len(nodea); i++ {
					if nodea[i].IPAddress == res.Results[0].Series[0].Values[j][5] {
						link.Target = i
						isSrc = true
					} else if nodea[i].IPAddress == res.Results[0].Series[0].Values[j][13] {
						link.Source = i
						isDst = true
					}
				}
				if isSrc && isDst {
					link.Action = res.Results[0].Series[0].Values[j][1].(string)
					links = append(links, link)
					isSrc = false
					isDst = false
				}
			}
		}
	}

	if len(links) == 0 {
		link.Source = 0
		link.Target = 0
		links = append(links, link)
	}

	return links, nil
}

func getName(tag string) string {
	var name string

	if strings.Contains(tag, "@usr:io.kubernetes.pod.name") {
		eachTag := strings.Split(tag, " ")
		for i := 0; i < len(eachTag); i++ {
			podName := strings.SplitAfter(eachTag[i], "=")
			for j := 0; j < len(podName); j++ {
				if podName[j] == "@usr:io.kubernetes.pod.name=" {
					name = podName[1]
				}
			}
		}
	} else {
		eachTag := strings.Split(tag, " ")
		containerName := strings.SplitAfter(eachTag[0], "=")
		if containerName != nil {
			name = containerName[1]
		} else {
			name = "unknown"
		}
	}

	return name
}
