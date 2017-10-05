package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//var js []JSONData

type GraphData struct {
	Nodes []Nodes `json:"nodes"`
	Links []Links `json:"links"`
}

type Nodes struct {
	ID    string `json:"id"`
	Group int    `json:"group"`
}

type Links struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

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

func GetData(w http.ResponseWriter, r *http.Request) {

	body, res := GetContainerEvents()

	json.Unmarshal(body, &res)

	jso := Transform(res)

	json.NewEncoder(w).Encode(jso)
}

func GetContainerEvents() ([]byte, InfluxData) {
	var res InfluxData
	resp, err := http.Get("http://influxdb:8086/query?db=flowDB&&q=SELECT%20*%20FROM%20ContainerEvents")
	if err != nil {
		fmt.Print(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	return body, res
}

func GetFlowEvents() ([]byte, InfluxData) {
	var res InfluxData
	resp, err := http.Get("http://influxdb:8086/query?db=flowDB&&q=SELECT%20*%20FROM%20FlowEvents")
	if err != nil {
		fmt.Print(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(body, &res)
	return body, res
}

func DeleteContainerEvents(id []string) []Nodes {
	var node Nodes
	var nodea []Nodes
	for i := 0; i < len(id); i++ {
		_, err := http.Get("http://influxdb:8086/query?db=flowDB&q=DELETE%20FROM%20%22ContainerEvents%22%20WHERE%20(%22EventID%22%20=%20%27" + id[i] + "%27)")
		if err != nil {
			fmt.Println(err)
		}
	}
	body, res := GetContainerEvents()

	json.Unmarshal(body, &res)
	for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
		node.ID = res.Results[0].Series[0].Values[j][1].(string)
		nodea = append(nodea, node)
	}
	return nodea
}

func Transform(res InfluxData) GraphData {
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
						nodea = append(nodea, node)
					}
				} else {
					id = append(id, res.Results[0].Series[0].Values[j][1].(string))
					nodea = DeleteContainerEvents(id)
				}
			}
		}
	}
	linka = GenerateLinks(nodea)
	jso := GraphData{Nodes: nodea, Links: linka}

	return jso
}

func GenerateLinks(nodea []Nodes) []Links {
	_, res := GetFlowEvents()
	var linka []Links
	var link Links
	var k int
	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == "FlowEvents" {
			for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
				for i := 0; i < len(nodea); i++ {
					if nodea[i].ID == res.Results[0].Series[0].Values[j][4] {
						link.Source = nodea[i].ID
					} else if nodea[i].ID == res.Results[0].Series[0].Values[j][12] {
						link.Target = nodea[i].ID
					}
					if link.Source != "" && link.Target != "" {
						link.Value = k + 1
						linka = append(linka, link)
						k++
					}
				}
			}
		}
	}
	if len(linka) == 0 {
		link.Source = nodea[0].ID
		link.Target = nodea[0].ID
		link.Value = 2
		linka = append(linka, link)

	}

	return linka
}
