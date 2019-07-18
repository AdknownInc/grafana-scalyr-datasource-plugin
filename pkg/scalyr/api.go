// Copyright 2019 Adknown Inc. All rights reserved.
// Created:  03/07/19
// Author:   matt

package scalyr

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
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
)

// NewPtr creates a new Scalyr object and returns a pointer
func NewPtr(readLogToken string, readConfigToken string) *Scalyr {
	return &Scalyr{
		readLogToken:    readLogToken,
		readConfigToken: readConfigToken,
		sessionId:       uuid.New().String(),
		client: http.Client{
			Timeout: 10,
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
	var bodyresp Response
	err = json.Unmarshal(body, &bodyresp)
	if err != nil {
		return nil, errors.Wrap(err, "Error on response body unmarshalling")
	}
	if bodyresp.Status == "success" {
		return "success", nil
	}
	return bodyresp.Status, errors.New(bodyresp.Message)

}

func closer(c io.Closer) {
	err := c.Close()
	if err != nil {
		panic(err)
	}
}
