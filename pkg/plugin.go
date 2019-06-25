package main

import (
	"github.com/grafana/grafana_plugin_model/go/datasource"
	"os"

	plugin "github.com/hashicorp/go-plugin"
	"log"
)

func main() {
	log.SetOutput(os.Stderr) // the plugin sends logs to the host process on strErr

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "grafana_plugin_type",
			MagicCookieValue: "datasource",
		},
		Plugins: map[string]plugin.Plugin{
			"backend-datasource": &datasource.DatasourcePluginImpl{Plugin: &ScalyrDatasource{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
