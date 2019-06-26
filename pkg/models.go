// Copyright 2019 Adknown Inc. All rights reserved.
// Created:  25/06/19
// Author:   matt
// Project:  grafana-scalyr-datasource-plugin

package main

import (
	"encoding/json"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"log"
)

type QueryRequest struct {
	From    string             `json:"from"`
	To      string             `json:"to"`
	Queries []*simplejson.Json `json:"queries"`
}

type ProxyResponse struct {
	Target     string        `json:"target"`
	Datapoints [][]float64   `json:"datapoints"`
	Queries    []interface{} `json:"queries"`
	RefId      string        `json:"refId"`
}

type NumericQuery struct {
	Filter    string `json:"filter"`
	Function  string `json:"function"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Buckets   int    `json:"buckets"`
	Priority  string `json:"priority"`
}

//UnmarshalJSON is one way to deal with the Query types that this is one way to deal with it, but its not great having to double unmarshal each query
func (p *ProxyResponse) UnmarshalJSON(data []byte) error {
	var err error
	var tp struct {
		Target     string        `json:"target"`
		Datapoints [][]float64   `json:"datapoints"`
		Queries    []interface{} `json:"queries"`
		RefId      string        `json:"refId"`
	}
	if err := json.Unmarshal(data, &tp); err != nil {
		log.Fatal(err)
	}
	p.Target = tp.Target
	p.Datapoints = tp.Datapoints
	p.RefId = tp.RefId
	p.Queries = tp.Queries
	for idx, query := range tp.Queries {
		switch query.(type) {
		case int:
			p.Queries[idx] = query.(int)
		case string:
			p.Queries[idx] = query.(string)
		default:
			d, err := json.Marshal(query)
			if err != nil {
				log.Fatal(err)
			}
			p.Queries[idx] = &NumericQuery{}
			if err = json.Unmarshal(d, p.Queries[idx]); err != nil {
				log.Fatal(err)
			}
		}
	}
	return err
}
