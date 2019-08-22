package scalyr

import (
	"reflect"
	"testing"
)

func TestParseComplexExpression(t *testing.T) {
	type args struct {
		expression             string
		start                  string
		end                    string
		buckets                int
		fullVariableExpression *string
		useNumeric             bool
	}
	aString := ""
	tests := []struct {
		name    string
		args    args
		want    []*ParseVar
		wantErr bool
	}{
		{
			args: args{
				expression:             "count(hello) + mean(world) / 24",
				fullVariableExpression: &aString,
			},
			want: []*ParseVar{
				{
					Id:     "var0",
					Filter: "count(hello)",
					Query: &TimeseriesQuery{
						Filter:   "hello",
						Function: "count",
						Priority: "low",
					},
					Response:   nil,
					ConstValue: 0,
				},
				{
					Id:     "var1",
					Filter: "mean(world)",
					Query: &TimeseriesQuery{
						Filter:   "world",
						Function: "mean",
						Priority: "low",
					},
					Response:   nil,
					ConstValue: 0,
				},
				{
					Id:         "var2",
					Filter:     "24",
					Query:      nil,
					Response:   nil,
					ConstValue: 24,
				},
			},
			wantErr: false,
		},
		{
			args: args{
				expression:             `(count($serverHost='server.dabbler.com' $logfile='/var/log/nginx/access.log' $uriPath matches "\\/r|\\/click\\.php" $status = 302 $redirectLocation matches "adssquared\\.com") + count($serverHost='ttt.server.com' $logfile='/var/log/nginx/access.log' $uriPath = "/fcc/" $status = 302 $redirectLocation matches "adssquared\\.com")) / count($serverHost='feeds.server.com' $logfile='/var/www/cfeeds.adsapphire.com/logs/feed_call_log.txt' $dataPostedDataServeUrl matches "dabbler\\.com|realtor\\.us|recipebook\\.us|castle\\.com|pets\\.us|cars\\.us|healthexchange\\.org|lifestylealive\\.com" $dataStatus == "ok" $dataPostedDataCachedAds == false) * 100`,
				fullVariableExpression: &aString,
			},
			want: []*ParseVar{
				{
					Id:         "0",
					Filter:     "count(hello)",
					Query:      nil,
					Response:   nil,
					ConstValue: 0,
				},
				{
					Id:         "1",
					Filter:     "sum(world)",
					Query:      nil,
					Response:   nil,
					ConstValue: 0,
				},
				{
					Id:         "2",
					Filter:     "",
					Query:      nil,
					Response:   nil,
					ConstValue: 24,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseComplexExpression(tt.args.expression, tt.args.start, tt.args.end, tt.args.buckets, tt.args.fullVariableExpression, tt.args.useNumeric)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseComplexExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseComplexExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEvaluateExpression(t *testing.T) {
	type args struct {
		expression string
		varArray   []*ParseVar
	}
	tests := []struct {
		name    string
		args    args
		want    *TimeseriesQueryResponse
		wantErr bool
	}{
		{
			name: "var0 + var1 / var2",
			args: args{
				expression: "var0 + var1 / var2",
				varArray: []*ParseVar{
					{
						Id:     "0",
						Filter: "",
						Query:  nil,
						Response: &TimeseriesQueryResponse{
							Status: "",
							Results: []TimeseriesQueryResult{
								{
									Values: []float64{
										235, 38456, 34126,
									},
								},
							},
							ExecutionTime: 0,
							Message:       "",
						},
						ConstValue: 0,
					},
					{
						Id:     "1",
						Filter: "",
						Query:  nil,
						Response: &TimeseriesQueryResponse{
							Status: "",
							Results: []TimeseriesQueryResult{
								{
									Values: []float64{
										300, 100, 200,
									},
								},
							},
							ExecutionTime: 0,
							Message:       "",
						},
						ConstValue: 0,
					},
					{
						Id:         "2",
						Filter:     "",
						Query:      nil,
						Response:   nil,
						ConstValue: 5,
					},
				},
			},
			want: &TimeseriesQueryResponse{
				Status: "success",
				Results: []TimeseriesQueryResult{
					{
						Values: []float64{
							295, 38476, 34166,
						},
					},
				},
				ExecutionTime: 0,
				Message:       "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEvaluateExpression(tt.args.expression, tt.args.varArray)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEvaluateExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEvaluateExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}
