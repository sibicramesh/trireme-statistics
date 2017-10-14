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
	ID    string `json:"id"`
	Group int    `json:"group"`
	Name  string `json:"name"`
}

// Links which holds the links between pu's
type Links struct {
	Source int    `json:"source"`
	Target int    `json:"target"`
	Value  int    `json:"value"`
	Action string `json:"action"`
}

// InfluxData is th estruct that holds the data returned from influxdb api
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
	var res InfluxData
	body, err := getContainerEvents()
	if err != nil {
		//TODO
	}
	json.Unmarshal(body, &res)
	jso := transform(res)
	json.NewEncoder(w).Encode(jso)
}

// GetGraph is used to parse html with custom address to request for json
func GetGraph(w http.ResponseWriter, r *http.Request) {

	t, err := template.New("graph").Parse(js)
	if err != nil {
		fmt.Println(err)
	}
	data := struct {
		Address string
	}{
		Address: r.URL.Query().Get("address"),
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "text/html")
}

func getContainerEvents() ([]byte, error) {
	return getByURI("http://influxdb:8086/query?db=flowDB&&q=SELECT%20*%20FROM%20ContainerEvents")
}

func getFlowEvents() ([]byte, error) {
	return getByURI("http://influxdb:8086/query?db=flowDB&&q=SELECT%20*%20FROM%20FlowEvents")
}

func getByURI(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func deleteContainerEvents(id []string) []Nodes {
	//TODO: better var names
	var node Nodes
	var nodea []Nodes
	for i := 0; i < len(id); i++ {
		_, err := http.Get("http://influxdb:8086/query?db=flowDB&q=DELETE%20FROM%20%22ContainerEvents%22%20WHERE%20(%22EventID%22%20=%20%27" + id[i] + "%27)")
		if err != nil {
			// TODO: Handle error better
			fmt.Println(err)
		}
	}
	var res InfluxData
	body, err := getContainerEvents()
	if err != nil {
		//TODO handle error
	}

	json.Unmarshal(body, &res)
	for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
		node.ID = res.Results[0].Series[0].Values[j][1].(string)
		name := getName(res.Results[0].Series[0].Values[j][6].(string))
		node.Name = name
		nodea = append(nodea, node)
	}
	return nodea
}

// TODO: Add Multiline comment that explains exactly how is this algo. working
func transform(res InfluxData) GraphData {
	var nodea []Nodes
	var linka []Links

	var node Nodes
	var id []string
	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == "ContainerEvents" {
			for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
				if res.Results[0].Series[0].Values[j][2].(string) != "delete" {
					if res.Results[0].Series[0].Values[j][2].(string) == "start" {
						node.ID = res.Results[0].Series[0].Values[j][1].(string)
						name := getName(res.Results[0].Series[0].Values[j][6].(string))
						node.Name = name
						nodea = append(nodea, node)
					}
				} else {
					id = append(id, res.Results[0].Series[0].Values[j][1].(string))
					nodea = deleteContainerEvents(id)
				}
			}
		}
	}
	linka = generateLinks(nodea)

	// TODO: What is JSO, Why no JSON ?
	jso := GraphData{Nodes: nodea, Links: linka}

	return jso
}

func generateLinks(nodea []Nodes) []Links {
	body, err := getFlowEvents()
	if err != nil {
		//TODO
	}
	var res InfluxData

	json.Unmarshal(body, &res)

	// TODO: Better var names than linka link
	var linka []Links
	var link Links
	var isSrc, isDst bool
	var k int
	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == "FlowEvents" {
			for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
				for i := 0; i < len(nodea); i++ {
					if nodea[i].ID == res.Results[0].Series[0].Values[j][4] {
						link.Target = i
						isSrc = true
					} else if nodea[i].ID == res.Results[0].Series[0].Values[j][12] {
						link.Source = i
						isDst = true
					}
				}
				if isSrc && isDst {
					link.Value = k + 1
					link.Action = res.Results[0].Series[0].Values[j][1].(string)
					if link.Action == "reject" {
						link.Action = checkIfAccept(res.Results[0].Series[0].Values[j][12].(string))
					}
					linka = append(linka, link)
					isSrc = false
					isDst = false
					k++
				}
			}
		}
	}
	if len(linka) == 0 {
		link.Source = 0
		link.Target = 0
		link.Value = 2
		linka = append(linka, link)

	}

	return linka
}

func getName(tag string) string {
	eachTag := strings.Split(tag, " ")
	name := strings.SplitAfter(eachTag[0], "=")
	return name[1]
}

//TODO: Returning a string? Is there no better way to do this ? Define types or return booleans
func checkIfAccept(id string) string {
	body, err := getFlowEvents()
	if err != nil {
		//TODO add error handling
	}
	var res InfluxData

	json.Unmarshal(body, &res)

	for i := 0; i < len(res.Results[0].Series[0].Values); i++ {
		if id == res.Results[0].Series[0].Values[i][12].(string) {
			if res.Results[0].Series[0].Values[i][12].(string) == "accept" {
				return "nowaccepted"
			}
		}
	}
	return "reject"
}
