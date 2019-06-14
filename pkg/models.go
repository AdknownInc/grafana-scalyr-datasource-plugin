package main

type ProxyRequest struct {
	Targets map[string]interface{} `json:"targets"`
}
