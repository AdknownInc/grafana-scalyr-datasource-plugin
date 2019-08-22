// Copyright 2019 Adknown Inc. All rights reserved.
// Created:  03/07/19
// Author:   matt

package scalyr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type Scalyr struct {
	readLogToken    string
	sessionId       string
	readConfigToken string
	client          http.Client
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type BucketRequest struct {
	From            int64
	To              int64
	IntervalSeconds int
}

const (
	QueryURL           = "https://www.scalyr.com/api/query"
	NumericQueryURL    = "https://www.scalyr.com/api/numericQuery"
	TimeseriesQueryURL = "https://www.scalyr.com/api/timeseriesQuery"
	FacetQueryURL      = "https://www.scalyr.com/api/facetQuery"
	MaxBuckets         = 5000
)

// NewPtr creates a new Scalyr object and returns a pointer
func NewPtr(readLogToken string, readConfigToken string) *Scalyr {
	return &Scalyr{
		readLogToken:    readLogToken,
		readConfigToken: readConfigToken,
		sessionId:       uuid.New().String(),
		client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

//SetHttpClient overrides the existing http client
func (s *Scalyr) SetHttpClient(client http.Client) {
	s.client = client
}

func (s *Scalyr) TimeSeriesQuery(query *TimeseriesQuery) (*TimeseriesQueryResponse, error) {
	tsqReq := TimeseriesQueryRequest{
		Token: s.readLogToken,
		Queries: []*TimeseriesQuery{
			query,
		},
	}
	payload, err := json.Marshal(tsqReq)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to json.Marshal the time series request being sent")
	}
	fmt.Println(string(payload))
	req, err := http.NewRequest("POST", TimeseriesQueryURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, errors.Wrap(err, "Received an error after creating the post request object")
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Received an error in response to the request")
	}
	defer closer(resp.Body)

	body, _ := ioutil.ReadAll(resp.Body)
	var bodyresp TimeseriesQueryResponse
	err = json.Unmarshal(body, &bodyresp)
	if err != nil {
		return nil, errors.Wrap(err, "Error on response body unmarshalling")
	}
	if bodyresp.Status != "success" {
		return nil, errors.New(fmt.Sprintf("Scalyr api call failed. Status (%s), message: '%s'", bodyresp.Status, bodyresp.Message))
	}
	return &bodyresp, nil
}

func (s *Scalyr) ComplexTimeSeriesQuery(query *TimeseriesQuery) (*TimeseriesQueryResponse, []string, error) {
	//parse out the different filters
	totalResponse := &TimeseriesQueryResponse{
		Status:        "success",
		Results:       []TimeseriesQueryResult{},
		ExecutionTime: 0,
		Message:       "",
	}

	fullVarExpr := ""
	simpleExpressions, err := ParseComplexExpression(query.Filter, query.StartTime, query.EndTime, query.Buckets, &fullVarExpr, false)
	if err != nil {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("Error parsing query with filter %s", query.Filter[0:8]))
	}
	//individualExpressions := simpleExpressions
	for k, params := range simpleExpressions {
		if params.Query != nil {
			resp, err := s.TimeSeriesQuery(params.Query)
			if err != nil {
				return nil, nil, errors.Wrap(err, fmt.Sprintf("Error on segment %s", params.Filter))
			}
			simpleExpressions[k].Response = resp
		}
	}
	fullResponse, err := NewEvaluateExpression(fullVarExpr, simpleExpressions)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to evaluate complex query after making the simple api calls")
	}
	//add fullresponse to the total response
	totalResponse.Results = append(totalResponse.Results, fullResponse.Results...)
	totalResponse.ExecutionTime += fullResponse.ExecutionTime

	parts := make([]string, len(simpleExpressions))
	for i, part := range simpleExpressions {
		parts[i] = part.Filter
	}

	return totalResponse, parts, nil
}

//GetBuckets gets the number of buckets that would be appropriate for the passed in from and to parameters, giving each bucket
//approximately the number of seconds defined in intervalSeconds
func GetBuckets(data *BucketRequest) (int, error) {
	if data.From >= data.To {
		return -1, errors.New(fmt.Sprintf("GetBuckets(): param {from} (%v) was greater than param {to} (%v) ", data.From, data.To))
	}
	timeframe := data.To - data.From
	if data.IntervalSeconds < 1 {
		return -1, errors.New(fmt.Sprintf("GetBuckets(): intervalSeconds must be at least 1, received (%v) ", data.IntervalSeconds))
	}
	if int64(data.IntervalSeconds) >= timeframe {
		return 1, nil
	}

	buckets := timeframe / int64(data.IntervalSeconds)
	if buckets > MaxBuckets {
		return -1, errors.New(fmt.Sprintf("GetBuckets(): calculated buckets too large. Max allowed buckets is %d. Params: from (%v), to (%v), intervalSeconds (%v)", MaxBuckets, data.From, data.To, data.IntervalSeconds))
	}
	return int(buckets), nil
}

func closer(c io.Closer) {
	err := c.Close()
	if err != nil {
		panic(err)
	}
}
