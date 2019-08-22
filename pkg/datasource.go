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
	"time"
)

type ScalyrDatasource struct {
	plugin.NetRPCUnsupportedPlugin
}

//From time.go
const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
	secondsPerDay    = 24 * secondsPerHour
	secondsPerWeek   = 7 * secondsPerDay
)

//TODO: update this to reflect what a target should send
type Target struct {
	RefId           string
	QueryType       string //refers to the grafana query type
	Name            string `json:"target"`
	Filter          string
	GraphFunction   string
	SecondsInterval int
	ScalyrQueryType string //one of facet, numerical or complex numerical
	IntervalType    string //fixed or window
	ChosenType      string //chosen interval type, only relevant when the user has selected 'fixed' for IntervalType
}

const (
	IntervalTypeWindow          = "window" //the datapoint timestamps are evenly spaced in relation to the current time
	IntervalTypeFixed           = "fixed"  //the end datapoint timestamp is offset to make the rest of of the datapoints exactly align with some time value (the start of the minute, hour, day, etc.)
	FixedIntervalMinute         = "minute"
	FixedIntervalHour           = "hour"
	FixedIntervalDay            = "day"
	FixedIntervalWeek           = "week"
	ScalyrQueryFacet            = "facet query"
	ScalyrQueryNumerical        = "numeric query"
	ScalyrQueryComplexNumerical = "complex numeric query"
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

	svc, err := t.getClient(tsdbReq.Datasource)
	if err != nil {
		return nil, err
	}

	//TODO: have the target parsing be done in channels
	//parse the target requests
	for _, target := range targets {
		bucketRequest, err := t.getSelectedInterval(tsdbReq.TimeRange, target)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error with target %s", target.RefId))
		}
		buckets, err := scalyr.GetBuckets(bucketRequest)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error with target %s", target.RefId))
		}
		switch target.ScalyrQueryType {
		case ScalyrQueryNumerical:
			//TODO: gonna need to call the remainder between bucketRequest.To and tsdbReq.TimeRange.EpochToMS on IntervalTypeFixed and then combine results if we want Fixed to work properly like with the proxy server
			resp, err := svc.TimeSeriesQuery(&scalyr.TimeseriesQuery{
				Filter:    target.Filter,
				Buckets:   buckets,
				Function:  target.GraphFunction,
				StartTime: strconv.FormatInt(bucketRequest.From, 10),
				EndTime:   strconv.FormatInt(bucketRequest.To, 10),
				Priority:  "low",
			})
			if err != nil {
				return nil, errors.Wrap(err, "Error returned on a numeric query")
			}
			r, err := parseTimeSeriesResponse(resp, target, bucketRequest)
			if err != nil {
				return nil, err
			}
			response.Results = append(response.Results, r)
		case ScalyrQueryComplexNumerical:
			resp, parts, err := svc.ComplexTimeSeriesQuery(&scalyr.TimeseriesQuery{
				Filter:    target.Filter,
				Buckets:   buckets,
				Function:  target.GraphFunction,
				StartTime: strconv.FormatInt(bucketRequest.From, 10),
				EndTime:   strconv.FormatInt(bucketRequest.To, 10),
				Priority:  "low",
			})
			if err != nil {
				return nil, errors.Wrap(err, "Error returned on a complex numeric query")
			}
			r, err := parseTimeSeriesResponse(resp, target, bucketRequest)
			if err != nil {
				return nil, err
			}
			rawBytes, err := json.Marshal(parts)
			if err != nil {
				return nil, err
			}
			r.MetaJson = string(rawBytes)
			response.Results = append(response.Results, r)
		}
	}

	return response, nil
}

func parseTimeSeriesResponse(resp *scalyr.TimeseriesQueryResponse, target Target, bucketRequest *scalyr.BucketRequest) (*datasource.QueryResult, error) {
	series := &datasource.TimeSeries{}
	series.Name = target.Name
	startTime := bucketRequest.From * 1000
	interval := int64(bucketRequest.IntervalSeconds) * 1000 //convert seconds to milliseconds as that's what grafana is expecting
	endTime := startTime + interval
	//endTime := startTime
	for _, r := range resp.Results {
		for _, val := range r.Values {
			series.Points = append(series.Points, &datasource.Point{
				Timestamp: endTime,
				Value:     val,
			})
			if endTime > bucketRequest.To {
				log.Warn("Set a datapoint to be outside the range of the end value. datapoint ts: %d, end value: %d", endTime, bucketRequest.To)
			}
			endTime += interval
		}
	}

	s := make([]*datasource.TimeSeries, 0)
	sort.Slice(series.Points, func(i, j int) bool {
		return series.Points[i].Timestamp < series.Points[j].Timestamp
	})
	s = append(s, series)

	return &datasource.QueryResult{
		RefId:  target.RefId,
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

//getSelectedInterval gets the time range of the request based on the interval type selected.
//note that if it is IntervalTypeFixed that a remainder bucket value must be computed, otherwise there aren't enough
//datapoints to reach the right-side of the graph. Extending the endtime passed the graph throws off the shape
func (t *ScalyrDatasource) getSelectedInterval(trange *datasource.TimeRange, target Target) (*scalyr.BucketRequest, error) {
	if target.IntervalType == IntervalTypeWindow {
		return &scalyr.BucketRequest{
			From:            trange.FromEpochMs/1000 - int64(target.SecondsInterval),
			To:              trange.ToEpochMs / 1000,
			IntervalSeconds: target.SecondsInterval,
		}, nil
	}
	start := time.Unix(trange.FromEpochMs/1000, (trange.FromEpochMs%1000)*int64(time.Millisecond))
	end := time.Unix(trange.ToEpochMs/1000, (trange.ToEpochMs%1000)*int64(time.Millisecond))
	seconds := secondsPerMinute //default option for seconds in minute
	switch target.ChosenType {
	case FixedIntervalMinute:
		endMinute := end.UTC().Minute() + ((int(time.Minute) % secondsPerHour) / secondsPerMinute)
		start = time.Date(start.UTC().Year(), start.UTC().Month(), start.UTC().Day(), start.UTC().Hour(), start.UTC().Minute(), 0, 0, start.UTC().Location())
		end = time.Date(end.UTC().Year(), end.UTC().Month(), end.UTC().Day(), end.UTC().Hour(), endMinute, 0, 0, end.UTC().Location())
	case FixedIntervalHour:
		endHour := end.UTC().Hour() + ((int(time.Hour) % secondsPerDay) / secondsPerHour)
		start = time.Date(start.UTC().Year(), start.UTC().Month(), start.UTC().Day(), start.UTC().Hour(), 0, 0, 0, start.UTC().Location())
		end = time.Date(end.UTC().Year(), end.UTC().Month(), end.UTC().Day(), endHour, 0, 0, 0, end.UTC().Location())
		seconds = secondsPerHour
	case FixedIntervalDay:
		endDay := end.UTC().Day() + 1
		start = time.Date(start.UTC().Year(), start.UTC().Month(), start.UTC().Day(), 0, 0, 0, 0, start.UTC().Location())
		end = time.Date(end.UTC().Year(), end.UTC().Month(), endDay, 0, 0, 0, 0, end.UTC().Location())
		seconds = secondsPerDay
	case FixedIntervalWeek:
		endDay := end.UTC().Day() + 7
		start = time.Date(start.UTC().Year(), start.UTC().Month(), start.UTC().Day(), 0, 0, 0, 0, start.UTC().Location())
		end = time.Date(end.UTC().Year(), end.UTC().Month(), endDay, 0, 0, 0, 0, end.UTC().Location())
		seconds = secondsPerWeek
	}

	return &scalyr.BucketRequest{
		From:            start.Unix() - int64(seconds),
		To:              end.Unix(),
		IntervalSeconds: seconds,
	}, nil
}
