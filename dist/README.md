# grafana-scalyr-datasource-plugin
A plugin for using scalyr's data in grafana. Requires a proxy server

#Variable Templating
This plugin makes use of Grafana's variables feature. However, due to the fact that Scalyr uses "$" as a special character, this plugin uses "~" as a variable identifier




# TODO:
- Add screenshots to the README
- Add Annotation support for possible values to search scalyr for


# Example Proxy Server
https://github.com/AdknownInc/grafana-scalyr-proxy-server

Note: Updated the required version of Grafana to 5.3.x because JsonData interpoloation of the `routes` section of `plugin.json` was only added for the `url` parameter in 5.3.0 of Grafana 