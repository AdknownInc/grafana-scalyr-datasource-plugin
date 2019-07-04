// Copyright 2019 Adknown Inc. All rights reserved.
// Created:  03/07/19
// Author:   matt

package scalyr

import (
	"github.com/google/uuid"
	"io"
)

type Scalyr struct {
	readLogToken    string
	sessionId       string
	readConfigToken string
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
	}
}

func closer(c io.Closer) {
	err := c.Close()
	if err != nil {
		panic(err)
	}
}
