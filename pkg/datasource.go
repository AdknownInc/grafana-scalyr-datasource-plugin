package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/grafana/grafana_plugin_model/go/datasource"
	"github.com/hashicorp/go-plugin"
	"io/ioutil"
	"net/http"
)

type ScalyrDatasource struct {
	plugin.NetRPCUnsupportedPlugin
}

type ProxyResponse struct {
	Target     string
	Datapoints [][]float64
	Queries    []map[string]interface{}
	RefId      string
}

func (t *ScalyrDatasource) Query(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	url := tsdbReq.Datasource.Url + "/query"

	var bodyToSend map[string]interface{}

	err := json.Unmarshal([]byte(tsdbReq.Queries[0].ModelJson), &bodyToSend)
	if err != nil {
		return nil, err
	}

	var jsonToPass []byte

	if _, ok := bodyToSend["backendUse"]; !ok {
		toPass := map[string]interface{}{
			"targets": []interface{}{bodyToSend},
			"range": map[string]interface{}{
				"from": tsdbReq.TimeRange.FromEpochMs,
				"to":   tsdbReq.TimeRange.ToEpochMs,
			},
		}
		jsonToPass, err = json.Marshal(toPass)
		if err != nil {
			return nil, err
		}
	} else {
		toPass := bodyToSend["backendUse"]
		jsonToPass, err = json.Marshal(toPass)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonToPass))
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	queryResults, err := convertProxyResponse(body)
	if err != nil {
		return nil, err
	}

	grafanaResponse := datasource.DatasourceResponse{
		Results: queryResults,
	}

	return &grafanaResponse, nil
}

func convertProxyResponse(jsonBytes []byte) ([]*datasource.QueryResult, error) {
	var proxyResponses []ProxyResponse
	err := json.Unmarshal(jsonBytes, &proxyResponses)
	if err != nil {
		return nil, errors.New("Couldn't unmarshal: " + string(jsonBytes))
	}

	queryResults := make([]*datasource.QueryResult, 0)

	for _, proxyResponse := range proxyResponses {
		points := make([]*datasource.Point, 0)

		for _, pointArr := range proxyResponse.Datapoints {
			point := datasource.Point{
				Timestamp: int64(pointArr[1]),
				Value:     pointArr[0],
			}
			points = append(points, &point)
		}

		timeSeries := datasource.TimeSeries{
			Points: points,
		}

		complexQueryParts, err := json.Marshal(proxyResponse.Queries)
		if err != nil {
			return nil, err
		}

		queryResult := datasource.QueryResult{
			RefId: proxyResponse.RefId,
			Series: []*datasource.TimeSeries{
				&timeSeries,
			},
			MetaJson: string(complexQueryParts),
		}
		queryResults = append(queryResults, &queryResult)
	}

	return queryResults, nil
}
