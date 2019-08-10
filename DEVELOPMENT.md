# Adknown Scalyr Api Datasource 
a datasource plugin that utilizes Scalyr's API

# TODO:
- Add screenshots to the README
- Add Authentication support so that the backend can get the Scalyr Read Token / Write Token defined on the Config page of the data source
    - Uncomment the config section that adds read and write tokens
- Add Annotation support for possible values to search scalyr for

# Docs    


More documentation about datasource plugins can be found in the [Docs](https://github.com/grafana/grafana/blob/master/docs/sources/plugins/developing/datasources.md).

The backend needs to implement 4 urls:

 * `/` should return 200 ok. Used for "Test connection" on the datasource config page.
 * `/search` used by the find metric options on the query tab in panels.
 * `/query` should return metrics based on input.
 * `/annotations` should return annotations.

## Installation

```cd``` into your plugin directory (defined in the grafana conf file under [paths])

```
git clone https://github.com/AdknownInc/grafana-scalyr-datasource-plugin.git
```



### Query API

Example `timeserie` request
``` javascript
{
  "panelId": 1,
  "range": {
    "from": "2016-10-31T06:33:44.866Z",
    "to": "2016-10-31T12:33:44.866Z",
    "raw": {
      "from": "now-6h",
      "to": "now"
    }
  },
  "rangeRaw": {
    "from": "now-6h",
    "to": "now"
  },
  "interval": "30s",
  "intervalMs": 30000,
  "targets": [
     { "target": "upper_50", "refId": "A", "type": "timeserie" },
     { "target": "upper_75", "refId": "B", "type": "timeserie" }
  ],
  "format": "json",
  "maxDataPoints": 550
}
```

Example `timeserie` response
``` javascript
[
  {
    "target":"upper_75", // The field being queried for
    "datapoints":[
      [622,1450754160000],  // Metric value as a float , unixtimestamp in milliseconds
      [365,1450754220000]
    ]
  },
  {
    "target":"upper_90",
    "datapoints":[
      [861,1450754160000],
      [767,1450754220000]
    ]
  }
]
```

If the metric selected is `"type": "table"`, an example `table` response:
```json
[
  {
    "columns":[
      {"text":"Time","type":"time"},
      {"text":"Country","type":"string"},
      {"text":"Number","type":"number"}
    ],
    "rows":[
      [1234567,"SE",123],
      [1234567,"DE",231],
      [1234567,"US",321]
    ],
    "type":"table"
  }
]
```

### Annotation API

The annotation request from the Simple JSON Datasource is a POST request to
the /annotations endpoint in your datasource. The JSON request body looks like this:
``` javascript
{
  "range": {
    "from": "2016-04-15T13:44:39.070Z",
    "to": "2016-04-15T14:44:39.070Z"
  },
  "rangeRaw": {
    "from": "now-1h",
    "to": "now"
  },
  "annotation": {
    "name": "deploy",
    "datasource": "Simple JSON Datasource",
    "iconColor": "rgba(255, 96, 96, 1)",
    "enable": true,
    "query": "#deploy"
  }
}
```

Grafana expects a response containing an array of annotation objects in the
following format:

``` javascript
[
  {
    annotation: annotation, // The original annotation sent from Grafana.
    time: time, // Time since UNIX Epoch in milliseconds. (required)
    title: title, // The title for the annotation tooltip. (required)
    tags: tags, // Tags for the annotation. (optional)
    text: text // Text for the annotation. (optional)
  }
]
```

Note: If the datasource is configured to connect directly to the backend, you
also need to implement an OPTIONS endpoint at /annotations that responds
with the correct CORS headers:

```
Access-Control-Allow-Headers:accept, content-type
Access-Control-Allow-Methods:POST
Access-Control-Allow-Origin:*
```

### Search API

Example request
``` javascript
{ target: 'upper_50' }
```

The search api can either return an array or map.

Example array response
``` javascript
["upper_25","upper_50","upper_75","upper_90","upper_95"]
```

Example map response
``` javascript
[ { "text" :"upper_25", "value": 1}, { "text" :"upper_75", "value": 2} ]
```

### Changelog


0.0.2
 - Added a button to go directly to Scalyr's logs from the grafana graph


0.0.1
 - Matt Jourard started this implementation
 - He definitely knows what he's doing
