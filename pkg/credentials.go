package main

import (
	"encoding/json"
	"github.com/adknowninc/grafana-scalyr-datasource-plugin/pkg/scalyr"
	"github.com/grafana/grafana_plugin_model/go/datasource"
)

type DatasourceInfo struct {
	ReadLogKey    string
	ReadConfigKey string
}

func (t *ScalyrDatasource) getDsInfo(datasourceInfo *datasource.DatasourceInfo) (*DatasourceInfo, error) {
	var dsInfo DatasourceInfo
	if err := json.Unmarshal([]byte(datasourceInfo.JsonData), &dsInfo); err != nil {
		return nil, err
	}

	//TODO: detect if the token is not encypted. That's the case for datasources setup before version 6.1 or something
	decryptedData := datasourceInfo.GetDecryptedSecureJsonData()
	if v, ok := decryptedData["readlogtoken"]; ok {
		dsInfo.ReadLogKey = v
	}
	//included for legacy
	if v, ok := decryptedData["readtoken"]; ok {
		dsInfo.ReadLogKey = v
	}
	if v, ok := decryptedData["readconfigtoken"]; ok {
		dsInfo.ReadConfigKey = v
	}
	//included for legacy
	if v, ok := decryptedData["writetoken"]; ok {
		dsInfo.ReadConfigKey = v
	}

	return &dsInfo, nil
}

func (t *ScalyrDatasource) getClient(datasourceInfo *datasource.DatasourceInfo) (*scalyr.Scalyr, error) {
	dsInfo, err := t.getDsInfo(datasourceInfo)
	if err != nil {
		return nil, err
	}
	return scalyr.NewPtr(dsInfo.ReadLogKey, dsInfo.ReadConfigKey), nil
}
