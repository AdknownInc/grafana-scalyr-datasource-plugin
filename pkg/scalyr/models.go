// Copyright 2019 Adknown Inc. All rights reserved.
// Created:  26/06/19
// Author:   matt
// Project:  grafana-scalyr-datasource-plugin

package scalyr

//url: https://www.scalyr.com/api/timeseriesQuery
type TimeseriesQueryRequest struct {
	Token   string             `json:"token"`
	Queries []*TimeseriesQuery `json:"queries"`
}

type TimeseriesQuery struct {
	Filter    string `json:"filter"`
	Buckets   int    `json:"buckets"`
	Function  string `json:"function"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Priority  string `json:"priority"`
}

type TimeseriesQueryResponse struct {
	Status        string                  `json:"status"`
	Results       []TimeseriesQueryResult `json:"results"`
	ExecutionTime int                     `json:"executionTime"`
	Message       string                  `json:"message,omitempty"`
}

type TimeseriesQueryResult struct {
	Values              []float64 `json:"values"`
	ExecutionTime       int       `json:"executionTime"`
	FoundExistingSeries bool      `json:"foundExistingSeries"`
}

//url: https://www.scalyr.com/api/numericQuery
type NumericQueryRequest struct {
	Token     string `json:"token"`
	QueryType string `json:"queryType"`
	Filter    string `json:"filter"`
	Function  string `json:"function"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Buckets   int    `json:"buckets"`
	Priority  string `json:"priority"`
}

type NumericQueryResponse struct {
	Status   string    `json:"status"`
	Values   []float64 `json:"values"`
	CpuUsage int       `json:"cpuUsage"`
}
