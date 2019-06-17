# grafana-scalyr-datasource-plugin
A plugin for using scalyr's data in grafana. Requires a proxy server

#Variable Templating
This plugin makes use of Grafana's variables feature. However, due to the fact that Scalyr uses "$" as a special character, this plugin uses "~" as a variable identifier

You can still use Grafana's [[varname]] syntax. [More info on templating can be found here.](http://docs.grafana.org/reference/templating/)

### Escaping
Say, for example, you want to send a Scalyr query like this:

`$someScalryField contains "my special ~error~"`

If there are no Grafana variables with the name "error" then this will be sent literally

Now, say we have a Grafana variable with the name "error", that has the value "cool" you would escape the first tilde like so:

`$someScalryField contains "my special ~~error~"`

and it would interpolate to 

`$someScalryField contains "my special ~cool~"`

If you have a Grafana variable with the name "error" and you want it to literally say "~error~", simply escape it with a backslash

`$someScalryField contains "my special \~error"`

which will interpolate to

`$someScalryField contains "my special ~error"`

# TODO:
- Add screenshots to the README
- Add Annotation support for possible values to search scalyr for


# Example Proxy Server
https://github.com/AdknownInc/grafana-scalyr-proxy-server

Note: Updated the required version of Grafana to 5.3.x because JsonData interpoloation of the `routes` section of `plugin.json` was only added for the `url` parameter in 5.3.0 of Grafana 

# Development

To build the plugin, run the following in the root directory of the project:
`grunt default`