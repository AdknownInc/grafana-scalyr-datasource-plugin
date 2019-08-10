package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adknowninc/grafana-scalyr-datasource-plugin/pkg/scalyr"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana_plugin_model/go/datasource"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"sort"
	"strconv"
)

type ScalyrDatasource struct {
	plugin.NetRPCUnsupportedPlugin
}

//TODO: update this to reflect what a target should send
type Target struct {
	RefId           string
	QueryType       string //refers to the grafana query type
	Filter          string
	GraphFunction   string
	SecondsInterval int
	ScalyrQueryType string //one of facet, numerical or complex numerical
	IntervalType    string //fixed or window
	ChosenType      string //chosen interval type, only relevant when the user has selected 'fixed' for IntervalType
	LegendFormat    string
}

const (
	IntervalTypeFixed           = "fixed"
	IntervalTypeWindow          = "window"
	FixedIntervalMinute         = "minute"
	FixedIntervalHour           = "hour"
	FixedIntervalDay            = "day"
	FixedIntervalWeek           = "week"
	FixedIntervalMonth          = "month"
	ScalyrQueryFacet            = "facet"
	ScalyrQueryNumerical        = "numerical"
	ScalyrQueryComplexNumerical = "complex numerical"
)

type suggestData struct {
	Text  string
	Value string
}

func (t *ScalyrDatasource) Query(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	modelJson, err := simplejson.NewJson([]byte(tsdbReq.Queries[0].ModelJson))
	if err != nil {
		return nil, err
	}

	if modelJson.Get("queryType").MustString() == "metricFindQuery" {
		response, err := t.metricFindQuery(ctx, tsdbReq, modelJson, tsdbReq.TimeRange)
		if err != nil {
			return &datasource.DatasourceResponse{
				Results: []*datasource.QueryResult{
					{
						RefId: "metricFindQuery",
						Error: err.Error(),
					},
				},
			}, nil
		}
		return response, nil
	}

	response, err := t.handleQuery(tsdbReq)
	if err != nil {
		return &datasource.DatasourceResponse{
			Results: []*datasource.QueryResult{
				{
					Error: err.Error(),
				},
			},
		}, nil
	}

	return response, nil
}

func (t *ScalyrDatasource) handleQuery(tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	response := &datasource.DatasourceResponse{}

	//create the targets from the initial queries
	targets := make([]Target, 0)
	for _, query := range tsdbReq.Queries {
		target := Target{}
		if err := json.Unmarshal([]byte(query.ModelJson), &target); err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	//setup some parameters and the scalyr service
	//fromRaw, err := strconv.ParseInt(, 10, 64)
	//if err != nil {
	//	return nil, err
	//}
	//from := time.Unix(fromRaw/1000, fromRaw%1000*1000*1000)
	//toRaw, err := strconv.ParseInt(tsdbReq.TimeRange.ToRaw, 10, 64)
	//if err != nil {
	//	return nil, err
	//}
	//to := time.Unix(toRaw/1000, toRaw%1000*1000*1000)
	svc, err := t.getClient(tsdbReq.Datasource)
	if err != nil {
		return nil, err
	}

	//parse the
	for _, target := range targets {
		buckets, err := scalyr.GetBuckets(tsdbReq.TimeRange.FromEpochMs, tsdbReq.TimeRange.ToEpochMs, target.SecondsInterval)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error with target %s", target.RefId))
		}

		switch target.ScalyrQueryType {
		case ScalyrQueryNumerical:
			resp, err := svc.TimeSeriesQuery([]scalyr.TimeseriesQuery{
				{
					Filter:    target.Filter,
					Buckets:   buckets,
					Function:  target.GraphFunction,
					StartTime: strconv.FormatInt(tsdbReq.TimeRange.FromEpochMs, 10),
					EndTime:   strconv.FormatInt(tsdbReq.TimeRange.ToEpochMs, 10),
					Priority:  "low",
				},
			})
			r, err := parseTimeSeriesResponse(resp, target.RefId, tsdbReq.TimeRange.FromEpochMs, tsdbReq.TimeRange.ToEpochMs, target.SecondsInterval)
			if err != nil {
				return nil, err
			}
			response.Results = append(response.Results, r)
		}
	}

	return response, nil
}

func parseTimeSeriesResponse(resp *scalyr.TimeseriesQueryResponse, refId string, from int64, to int64, secondsInterval int) (*datasource.QueryResult, error) {
	series := &datasource.TimeSeries{}

	startTime := from
	endTime := startTime + int64(secondsInterval)
	for _, r := range resp.Results {
		for _, val := range r.Values {
			series.Points = append(series.Points, &datasource.Point{
				Timestamp: endTime,
				Value:     val,
			})
			if endTime > to {
				log.Warn("Set a datapoint to be outside the range of the end value. datapoint ts: %d, end value: %d", endTime, to)
			}
			endTime += int64(secondsInterval)
		}
	}

	s := make([]*datasource.TimeSeries, 0)
	sort.Slice(series.Points, func(i, j int) bool {
		return series.Points[i].Timestamp < series.Points[j].Timestamp
	})
	s = append(s, series)

	return &datasource.QueryResult{
		RefId:  refId,
		Series: s,
	}, nil
}

func (t *ScalyrDatasource) metricFindQuery(ctx context.Context, tsdbReq *datasource.DatasourceRequest, parameters *simplejson.Json, timeRange *datasource.TimeRange) (*datasource.DatasourceResponse, error) {
	//svc, err := t.getClient(tsdbReq.Datasource)
	//if err != nil {
	//	return nil, err
	//}

	subtype := parameters.Get("subtype").MustString()

	data := make([]suggestData, 0)
	switch subtype {
	case "named_query_names":
	default:
		data = append(data, suggestData{Text: "Not yet implemented", Value: "Not yet implemented"})
	}

	table := t.transformToTable(data)

	return &datasource.DatasourceResponse{
		Results: []*datasource.QueryResult{
			{
				RefId:  "metricFindQuery",
				Tables: []*datasource.Table{table},
			},
		},
	}, nil
}

func (t *ScalyrDatasource) transformToTable(data []suggestData) *datasource.Table {
	table := &datasource.Table{
		Columns: make([]*datasource.TableColumn, 0),
		Rows:    make([]*datasource.TableRow, 0),
	}
	table.Columns = append(table.Columns, &datasource.TableColumn{Name: "text"})
	table.Columns = append(table.Columns, &datasource.TableColumn{Name: "value"})

	for _, r := range data {
		row := &datasource.TableRow{}
		row.Values = append(row.Values, &datasource.RowValue{Kind: datasource.RowValue_TYPE_STRING, StringValue: r.Text})
		row.Values = append(row.Values, &datasource.RowValue{Kind: datasource.RowValue_TYPE_STRING, StringValue: r.Value})
		table.Rows = append(table.Rows, row)
	}
	return table
}

//Converts the alert request into the format expected by the proxy server
func handleAlertRequest(tsdbReq *datasource.DatasourceRequest, jsonRequest map[string]interface{}) ([]byte, error) {
	toPass := map[string]interface{}{
		"targets": []interface{}{jsonRequest},
		"range": map[string]interface{}{
			"from": tsdbReq.TimeRange.FromEpochMs,
			"to":   tsdbReq.TimeRange.ToEpochMs,
		},
	}
	jsonToPass, err := json.Marshal(toPass)
	if err != nil {
		return nil, err
	}

	return jsonToPass, nil
}

// Converts the response from the proxy server into the format that grafana expects from the backend plugin
func convertProxyResponse(jsonBytes []byte) ([]*datasource.QueryResult, error) {
	return nil, nil
	//var proxyResponses []ProxyResponse
	//err := json.Unmarshal(jsonBytes, &proxyResponses)
	//if err != nil {
	//	return nil, errors.New("Couldn't unmarshal: " + string(jsonBytes))
	//}
	//
	//queryResults := make([]*datasource.QueryResult, 0)
	//
	//for _, proxyResponse := range proxyResponses {
	//	points := make([]*datasource.Point, 0)
	//
	//	for _, pointArr := range proxyResponse.Datapoints {
	//		point := datasource.Point{
	//			Timestamp: int64(pointArr[1]),
	//			Value:     pointArr[0],
	//		}
	//		points = append(points, &point)
	//	}
	//
	//	timeSeries := datasource.TimeSeries{
	//		Points: points,
	//		Name:   proxyResponse.Target,
	//	}
	//
	//	complexQueryParts := proxyResponse.Queries
	//
	//	queryResult := datasource.QueryResult{
	//		RefId: proxyResponse.RefId,
	//		Series: []*datasource.TimeSeries{
	//			&timeSeries,
	//		},
	//		MetaJson: string(complexQueryParts),
	//	}
	//	queryResults = append(queryResults, &queryResult)
	//}
	//
	//return queryResults, nil
}
