package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana_plugin_model/go/datasource"
	"github.com/hashicorp/go-plugin"
	"math"
	"regexp"
	"sort"
	"strconv"
	"time"
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

	targets := make([]Target, 0)
	for _, query := range tsdbReq.Queries {
		target := Target{}
		if err := json.Unmarshal([]byte(query.ModelJson), &target); err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	fromRaw, err := strconv.ParseInt(tsdbReq.TimeRange.FromRaw, 10, 64)
	if err != nil {
		return nil, err
	}
	from := time.Unix(fromRaw/1000, fromRaw%1000*1000*1000)
	toRaw, err := strconv.ParseInt(tsdbReq.TimeRange.ToRaw, 10, 64)
	if err != nil {
		return nil, err
	}
	to := time.Unix(toRaw/1000, toRaw%1000*1000*1000)
	svc, err := t.getClient(tsdbReq.Datasource)
	if err != nil {
		return nil, err
	}
	for _, target := range targets {
		switch target.ScalyrQueryType {
		case ScalyrQueryFacet:
			r, err := parseTimeSeriesResponse(resp, target.RefId, from, to, target.TimestampColumn, target.ValueColumn, target.LegendFormat)
			if err != nil {
				return nil, err
			}
			response.Results = append(response.Results, r)
		case ScalyrQueryNumerical:
			r, err := parseTableResponse(resp, target.RefId, from, to, target.TimestampColumn)
			if err != nil {
				return nil, err
			}
			response.Results = append(response.Results, r)
		case ScalyrQueryComplexNumerical:

		}
	}

	return response, nil
}

func (t *ScalyrDatasource) metricFindQuery(ctx context.Context, tsdbReq *datasource.DatasourceRequest, parameters *simplejson.Json, timeRange *datasource.TimeRange) (*datasource.DatasourceResponse, error) {
	region := parameters.Get("region").MustString()
	svc, err := t.getClient(tsdbReq.Datasource, region)
	if err != nil {
		return nil, err
	}

	subtype := parameters.Get("subtype").MustString()

	data := make([]suggestData, 0)
	switch subtype {
	case "named_query_names":
		li := &athena.ListNamedQueriesInput{}
		lo := &athena.ListNamedQueriesOutput{}
		err = svc.ListNamedQueriesPages(li,
			func(page *athena.ListNamedQueriesOutput, lastPage bool) bool {
				lo.NamedQueryIds = append(lo.NamedQueryIds, page.NamedQueryIds...)
				return !lastPage
			})
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(lo.NamedQueryIds); i += 50 {
			e := int64(math.Min(float64(i+50), float64(len(lo.NamedQueryIds))))
			bi := &athena.BatchGetNamedQueryInput{NamedQueryIds: lo.NamedQueryIds[i:e]}
			bo, err := svc.BatchGetNamedQuery(bi)
			if err != nil {
				return nil, err
			}
			for _, q := range bo.NamedQueries {
				data = append(data, suggestData{Text: *q.Name, Value: *q.Name})
			}
		}
	case "named_query_queries":
		pattern := parameters.Get("pattern").MustString()
		r := regexp.MustCompile(pattern)
		li := &athena.ListNamedQueriesInput{}
		lo := &athena.ListNamedQueriesOutput{}
		err = svc.ListNamedQueriesPages(li,
			func(page *athena.ListNamedQueriesOutput, lastPage bool) bool {
				lo.NamedQueryIds = append(lo.NamedQueryIds, page.NamedQueryIds...)
				return !lastPage
			})
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(lo.NamedQueryIds); i += 50 {
			e := int64(math.Min(float64(i+50), float64(len(lo.NamedQueryIds))))
			bi := &athena.BatchGetNamedQueryInput{NamedQueryIds: lo.NamedQueryIds[i:e]}
			bo, err := svc.BatchGetNamedQuery(bi)
			if err != nil {
				return nil, err
			}
			for _, q := range bo.NamedQueries {
				if r.MatchString(*q.Name) {
					data = append(data, suggestData{Text: *q.QueryString, Value: *q.QueryString})
				}
			}
		}
	case "query_execution_ids":
		toRaw, err := strconv.ParseInt(timeRange.ToRaw, 10, 64)
		if err != nil {
			return nil, err
		}
		to := time.Unix(toRaw/1000, toRaw%1000*1000*1000)

		pattern := parameters.Get("pattern").MustString()
		var workGroupParam *string
		workGroupParam = nil
		if workGroup, ok := parameters.CheckGet("work_group"); ok {
			temp := workGroup.MustString()
			workGroupParam = &temp
		}
		r := regexp.MustCompile(pattern)
		limit := parameters.Get("limit").MustInt()
		li := &athena.ListQueryExecutionsInput{
			WorkGroup: workGroupParam,
		}
		lo := &athena.ListQueryExecutionsOutput{}
		err = svc.ListQueryExecutionsPagesWithContext(ctx, li,
			func(page *athena.ListQueryExecutionsOutput, lastPage bool) bool {
				lo.QueryExecutionIds = append(lo.QueryExecutionIds, page.QueryExecutionIds...)
				return !lastPage
			})
		fbo := make([]*athena.QueryExecution, 0)
		for i := 0; i < len(lo.QueryExecutionIds); i += 50 {
			e := int64(math.Min(float64(i+50), float64(len(lo.QueryExecutionIds))))
			bi := &athena.BatchGetQueryExecutionInput{QueryExecutionIds: lo.QueryExecutionIds[i:e]}
			bo, err := svc.BatchGetQueryExecution(bi)
			if err != nil {
				return nil, err
			}
			for _, q := range bo.QueryExecutions {
				if *q.Status.State != "SUCCEEDED" {
					continue
				}
				if (*q.Status.CompletionDateTime).After(to) {
					continue
				}
				if r.MatchString(*q.Query) {
					fbo = append(fbo, q)
				}
			}
		}
		sort.Slice(fbo, func(i, j int) bool {
			return fbo[i].Status.CompletionDateTime.After(*fbo[j].Status.CompletionDateTime)
		})
		limit = int(math.Min(float64(limit), float64(len(fbo))))
		for _, q := range fbo[0:limit] {
			data = append(data, suggestData{Text: *q.QueryExecutionId, Value: *q.QueryExecutionId})
		}
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
			Name:   proxyResponse.Target,
		}

		complexQueryParts := proxyResponse.Queries

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
