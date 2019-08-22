// Copyright 2018 Adknown Inc. All rights reserved.
// Created:  06/06/18
// Author:   matt
// Project:  gen

package scalyr

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

const (
	testHost       = "goScalyrApi"
	testTier       = "test"
	testSystem     = "goAPITesting"
	testThreadName = "Test Thread"
	testLogfile    = "/arbitary/file/path.txt"
	testParser     = "json"
)

var (
	testReadKey        = os.Getenv("SCALYR_READ_KEY")
	existingServerHost = os.Getenv("SCALYR_EXISTING_SERVERHOST")
)

func TestScalyr_TimeSeriesQuery(t *testing.T) {

	type fields struct {
		readLogToken    string
		sessionId       string
		readConfigToken string
		client          http.Client
	}
	type args struct {
		query TimeseriesQuery
	}

	//set the defaults
	sessionId := uuid.New().String()
	defaultFields := fields{
		readLogToken:    testReadKey,
		sessionId:       sessionId,
		readConfigToken: testReadKey,
		client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
	endTime := time.Now().Unix()
	startTime := endTime - (10 * 60)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *TimeseriesQueryResponse
		wantErr bool
	}{
		{
			name:   "Basic Test",
			fields: defaultFields,
			args: args{
				query: TimeseriesQuery{
					Filter:    fmt.Sprintf("$serverHost = '%s'", existingServerHost),
					Buckets:   5,
					Function:  "count",
					StartTime: strconv.FormatInt(startTime, 10),
					EndTime:   strconv.FormatInt(endTime, 10),
					Priority:  "low",
				},
			},
			want: &TimeseriesQueryResponse{
				Status: "success",
				Results: []TimeseriesQueryResult{
					{
						Values:              []float64{0, 0, 0, 0, 0},
						ExecutionTime:       0,
						FoundExistingSeries: false,
					},
				},
				ExecutionTime: 100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scalyr{
				readLogToken:    tt.fields.readLogToken,
				sessionId:       tt.fields.sessionId,
				readConfigToken: tt.fields.readConfigToken,
				client:          tt.fields.client,
			}
			got, err := s.TimeSeriesQuery(tt.args.queries)
			if (err != nil) != tt.wantErr {
				t.Errorf("TimeSeriesQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("TimeSeriesQuery() error: returned a nil response object")
				return
			}

			if got.Status != tt.want.Status {
				t.Errorf("TimeSeriesQuery() status, got = %v, want %v", got.Status, tt.want.Status)
				return
			}
			if len(got.Results[0].Values) != len(tt.want.Results[0].Values) {
				t.Errorf("TimeSeriesQuery() results did not have the expected number of values. got %v, want %v", len(got.Results[0].Values), len(tt.want.Results[0].Values))
				return
			}
		})
	}
}
