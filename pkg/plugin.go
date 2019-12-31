package main

import (
	"github.com/grafana/grafana_plugin_model/go/datasource"
	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	_ "net/http/pprof"
	"os"
	"strings"
)

var appLogger = hclog.New(&hclog.LoggerOptions{
	Name:            "grafana-scalyr-datasource-plugin",
	Level:           hclog.LevelFromString("DEBUG"),
	Output:          os.Stderr,
	JSONFormat:      false,
	IncludeLocation: false,
})

func main() {
	//log.SetOutput(os.Stderr) // the plugin sends logs to the host process on strErr
	logLevel := os.Getenv("GF_SCALYR_LOGGING_LEVEL")
	if logLevel != "" {
		logLevel = strings.ToUpper(logLevel)
		//appLogger.SetLevel(hclog.LevelFromString(logLevel))
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "grafana_plugin_type",
			MagicCookieValue: "datasource",
		},
		Plugins: map[string]plugin.Plugin{
			"backend-datasource": &datasource.DatasourcePluginImpl{Plugin: &ScalyrDatasource{
				logger: appLogger,
			}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
