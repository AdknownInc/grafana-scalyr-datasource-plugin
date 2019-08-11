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

func (s *Scalyr) TimeSeriesQuery(queries []TimeseriesQuery) (*TimeseriesQueryResponse, error) {
	tsqReq := TimeseriesQueryRequest{
		Token:   s.readLogToken,
		Queries: queries,
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

//GetBuckets gets the number of buckets that would be appropriate for the passed in from and to parameters, giving each bucket
//approximately the number of seconds defined in intervalSeconds
func GetBuckets(from int64, to int64, intervalSeconds int) (int, error) {
	if from >= to {
		return -1, errors.New(fmt.Sprintf("GetBuckets(): param {from} (%v) was greater than param {to} (%v) ", from, to))
	}
	timeframe := to - from
	if intervalSeconds < 1 {
		return -1, errors.New(fmt.Sprintf("GetBuckets(): intervalSeconds must be at least 1, received (%v) ", intervalSeconds))
	}
	if int64(intervalSeconds) >= timeframe {
		return 1, nil
	}

	buckets := timeframe / int64(intervalSeconds)
	if buckets > MaxBuckets {
		return -1, errors.New(fmt.Sprintf("GetBuckets(): calculated buckets too large. Max allowed buckets is %d. Params: from (%v), to (%v), intervalSeconds (%v)", MaxBuckets, from, to, intervalSeconds))
	}
	return int(buckets), nil
}

func closer(c io.Closer) {
	err := c.Close()
	if err != nil {
		panic(err)
	}
}
