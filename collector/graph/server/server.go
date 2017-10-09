package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

//var js []JSONData

type GraphData struct {
	Nodes []Nodes `json:"nodes"`
	Links []Links `json:"links"`
}

type Nodes struct {
	ID    string `json:"id"`
	Group int    `json:"group"`
	Name  string `json:"name"`
}

type Links struct {
	Source int    `json:"source"`
	Target int    `json:"target"`
	Value  int    `json:"value"`
	Action string `json:"action"`
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
		name := GetName(res.Results[0].Series[0].Values[j][6].(string))
		node.Name = name
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
						name := GetName(res.Results[0].Series[0].Values[j][6].(string))
						node.Name = name
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
						link.Action = CheckIfAccept(res.Results[0].Series[0].Values[j][12].(string))
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

func GetName(tag string) string {
	eachTag := strings.Split(tag, " ")
	name := strings.SplitAfter(eachTag[0], "=")
	return name[1]
}

func CheckIfAccept(id string) string {
	_, res := GetFlowEvents()
	for i := 0; i < len(res.Results[0].Series[0].Values); i++ {
		if id == res.Results[0].Series[0].Values[i][12].(string) {
			if res.Results[0].Series[0].Values[i][12].(string) == "accept" {
				return "nowaccepted"
			}
		}
	}
	return "reject"
}
