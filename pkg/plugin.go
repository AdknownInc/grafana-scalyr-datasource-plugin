package main

import (
	"github.com/grafana/grafana_plugin_model/go/datasource"
	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
)

var appLogger = hclog.New(&hclog.LoggerOptions{
	Name:            "grafana-scalyr-datasource-plugin",
	Level:           hclog.LevelFromString("DEBUG"),
	Output:          os.Stderr,
	JSONFormat:      false,
	IncludeLocation: false,
})

func main() {
	runtime.SetBlockProfileRate(1)

	go func() {
		http.ListenAndServe(":3010", nil)
	}()

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "grafana_plugin_type",
			MagicCookieValue: "datasource",
		},
		Plugins: map[string]plugin.Plugin{
			"grafana-scalyr-datasource-plugin": &datasource.DatasourcePluginImpl{Plugin: &ScalyrDatasource{
				logger: appLogger,
			}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
