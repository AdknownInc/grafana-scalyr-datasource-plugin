"use strict";

System.register(["lodash"], function (_export, _context) {
    "use strict";

    var _, _createClass, ScalyrDatasource;

    function _classCallCheck(instance, Constructor) {
        if (!(instance instanceof Constructor)) {
            throw new TypeError("Cannot call a class as a function");
        }
    }

    function handleTsdbResponse(response) {
        var res = [];
        _.forEach(response.data.results, function (r) {
            _.forEach(r.series, function (s) {
                res.push({ target: s.name, datapoints: s.points, queries: r.meta, refId: r.refId });
            });
            _.forEach(r.tables, function (t) {
                t.type = 'table';
                t.refId = r.refId;
                res.push(t);
            });
        });

        response.data = res;
        return response;
    }

    _export("handleTsdbResponse", handleTsdbResponse);

    return {
        setters: [function (_lodash) {
            _ = _lodash.default;
        }],
        execute: function () {
            _createClass = function () {
                function defineProperties(target, props) {
                    for (var i = 0; i < props.length; i++) {
                        var descriptor = props[i];
                        descriptor.enumerable = descriptor.enumerable || false;
                        descriptor.configurable = true;
                        if ("value" in descriptor) descriptor.writable = true;
                        Object.defineProperty(target, descriptor.key, descriptor);
                    }
                }

                return function (Constructor, protoProps, staticProps) {
                    if (protoProps) defineProperties(Constructor.prototype, protoProps);
                    if (staticProps) defineProperties(Constructor, staticProps);
                    return Constructor;
                };
            }();

            _export("ScalyrDatasource", ScalyrDatasource = function () {
                function ScalyrDatasource(instanceSettings, $q, backendSrv, templateSrv, timeSrv) {
                    _classCallCheck(this, ScalyrDatasource);

                    this.datasourceId = instanceSettings.id;
                    this.type = instanceSettings.type;
                    this.url = instanceSettings.url;
                    this.name = instanceSettings.name;
                    this.q = $q;
                    this.backendSrv = backendSrv;
                    this.templateVarIdentifier = '~';
                    this.templateVarEscaperChar = "\\";
                    this.templateSrv = ScalyrDatasource.modifyTemplateVariableIdentifier(templateSrv, this.templateVarIdentifier);
                    this.templateSrv = ScalyrDatasource.addTemplateVariableEscapeChar(this.templateSrv, this.templateVarEscaperChar, this.templateVarIdentifier);
                    this.timeSrv = timeSrv;
                    this.withCredentials = instanceSettings.withCredentials;
                    this.headers = { 'Content-Type': 'application/json' };
                    if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
                        this.headers['Authorization'] = instanceSettings.basicAuth;
                    }

                    this.parseComplex = instanceSettings.jsonData.parseQueries || false;

                    this.queryControls = [];

                    //Used for escaping template variables, needed to write regex backwards to take advantage of lookbehinds
                    String.prototype.reverse = function () {
                        return this.split('').reverse().join('');
                    };
                }

                _createClass(ScalyrDatasource, [{
                    key: "removeEscapeChar",
                    value: function removeEscapeChar(filter) {
                        return filter.replace(RegExp("\\" + this.templateVarEscaperChar + this.templateVarIdentifier, 'g'), this.templateVarIdentifier);
                    }
                }, {
                    key: "findAndReverse",
                    value: function findAndReverse(filter) {
                        var newFilter = filter.reverse();
                        return newFilter.replace(RegExp("(\\w+)(?=" + this.templateVarIdentifier + "(?!\\" + this.templateVarEscaperChar + "))", 'g'), function (a, b) {
                            return b.reverse();
                        });
                    }
                }, {
                    key: "reverseAllVariables",
                    value: function reverseAllVariables() {
                        for (var variable in this.templateSrv.index) {
                            // noinspection JSUnfilteredForInLoop
                            if (this.templateSrv.index[variable] instanceof Object && this.templateSrv.index[variable].hasOwnProperty("current")) {
                                // noinspection JSUnfilteredForInLoop
                                this.templateSrv.index[variable].current.value = this.templateSrv.index[variable].current.value.reverse();
                            }
                        }
                    }
                }, {
                    key: "query",
                    value: function query(options) {
                        var _this = this;

                        var parsedOptions = this.buildQueryParameters(options);
                        parsedOptions.targets = options.targets.filter(function (t) {
                            return !t.hide;
                        });
                        if (options.targets.length <= 0) {
                            return this.q.when({ data: [] });
                        }
                        //Deep copy the object. When template variables are swapped out we don't want to modify the original values
                        var finalOptions = _.cloneDeep(parsedOptions);

                        for (var i = 0; i < finalOptions.targets.length; i++) {
                            this.reverseAllVariables();
                            var filter = this.findAndReverse(finalOptions.targets[i].filter);
                            finalOptions.targets[i].filter = this.findAndReverse(this.templateSrv.replace(filter, null, 'regex'));
                            finalOptions.targets[i].filter = this.removeEscapeChar(finalOptions.targets[i].filter);
                            this.reverseAllVariables();

                            //Run through forwards for square bracket variable syntax
                            finalOptions.targets[i].filter = this.templateSrv.replace(finalOptions.targets[i].filter, null, 'regex');

                            //Grafana adds regex escapes to the variables for some reason
                            finalOptions.targets[i].filter = finalOptions.targets[i].filter.replace(/\\(.)/g, "$1");
                        }

                        finalOptions.parseComplex = this.parseComplex;

                        finalOptions.user = this.backendSrv.contextSrv.user.name;
                        finalOptions.userId = this.backendSrv.contextSrv.user.id;
                        finalOptions.org = this.backendSrv.contextSrv.user.orgName;
                        finalOptions.orgId = this.backendSrv.contextSrv.user.orgId;

                        //This is needed because Grafana messes up the ordering when moving the response from backend to frontend
                        var refIdMap = _.map(options.targets, function (target) {
                            return target.refId;
                        });

                        return this.doTsdbRequest(finalOptions).then(handleTsdbResponse).then(function (res) {
                            res.data.sort(function (a, b) {
                                return refIdMap.indexOf(a.refId) - refIdMap.indexOf(b.refId);
                            });
                            _this.response = res;
                            var _iteratorNormalCompletion = true;
                            var _didIteratorError = false;
                            var _iteratorError = undefined;

                            try {
                                for (var _iterator = _this.queryControls[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                                    var queryControl = _step.value;

                                    queryControl.getComplexParts();
                                }
                            } catch (err) {
                                _didIteratorError = true;
                                _iteratorError = err;
                            } finally {
                                try {
                                    if (!_iteratorNormalCompletion && _iterator.return) {
                                        _iterator.return();
                                    }
                                } finally {
                                    if (_didIteratorError) {
                                        throw _iteratorError;
                                    }
                                }
                            }

                            return res;
                        });
                    }
                }, {
                    key: "testDatasource",
                    value: function testDatasource() {
                        var _this2 = this;

                        return this.doMetricQueryRequest('test_datasource', {}).then(function (response) {
                            return _this2.q.when({ status: "success", message: "Data source is working", title: "Success" });
                        }).catch(function (err) {
                            return { status: "error", message: err.message, title: "Error" };
                        });
                    }
                }, {
                    key: "metricFindQuery",
                    value: function metricFindQuery(query) {
                        var serverHostsQuery = query.match(/^server_hosts\(\)/);
                        if (serverHostsQuery) {
                            return this.doMetricQueryRequest('server_hosts', {});
                        }

                        var logFilesQuery = query.match(/^log_files\((.+)\)/);
                        if (logFilesQuery) {
                            var serverHost = logFilesQuery[1];
                            return this.doMetricQueryRequest('named_query_queries', {
                                serverHost: this.templateSrv.replace(serverHost)
                            });
                        }

                        return this.q.when([]);
                    }
                }, {
                    key: "doTsdbRequest",
                    value: function doTsdbRequest(options) {
                        var tsdbRequestData = {
                            from: options.range.from.valueOf().toString(),
                            to: options.range.to.valueOf().toString(),
                            queries: options.targets
                        };

                        return this.backendSrv.datasourceRequest({
                            url: '/api/tsdb/query',
                            method: 'POST',
                            data: tsdbRequestData
                        });
                    }
                }, {
                    key: "buildQueryParameters",
                    value: function buildQueryParameters(options) {
                        var _this3 = this;

                        //remove placeholder targets
                        options.targets = _.filter(options.targets, function (target) {
                            return target.target !== 'select metric';
                        });

                        options.targets = _.map(options.targets, function (target) {
                            return {
                                datasourceId: _this3.datasourceId,
                                refId: target.refId,
                                hide: target.hide,

                                queryType: 'query',
                                type: target.type,
                                scalyrQueryType: target.type,
                                subtype: target.type || 'timeserie',
                                chosenType: target.chosenType,
                                target: _this3.templateSrv.replace(target.target, options.scopedVars, 'regex'), //the name of the query
                                filter: target.filter, //the filter sent to scalyr
                                graphFunction: target.graphFunction, //the type of function that is needed on Scalyr's end
                                intervalType: target.intervalType,
                                secondsInterval: target.secondsInterval,
                                showQueryParts: target.showQueryParts
                            };
                        });

                        return options;
                    }
                }, {
                    key: "doMetricQueryRequest",
                    value: function doMetricQueryRequest(subtype, parameters) {
                        var range = this.timeSrv.timeRange();
                        return this.backendSrv.datasourceRequest({
                            url: '/api/tsdb/query',
                            method: 'POST',
                            data: {
                                from: range.from.valueOf().toString(),
                                to: range.to.valueOf().toString(),
                                queries: [_.extend({
                                    refId: 'metricFindQuery',
                                    datasourceId: this.datasourceId,
                                    queryType: 'metricFindQuery',
                                    subtype: subtype
                                }, parameters)]
                            }
                        }).then(function (r) {
                            return ScalyrDatasource.transformSuggestDataFromTable(r.data);
                        });
                    }
                }], [{
                    key: "modifyTemplateVariableIdentifier",
                    value: function modifyTemplateVariableIdentifier(templateSrv, newIdentifier) {
                        var regStr = templateSrv.regex.source;

                        //There are 2 occurrences of '\$'. Remember to escape!
                        regStr = regStr.replace(/\\\$/g, newIdentifier);
                        templateSrv.regex = new RegExp(regStr, 'g');
                        return templateSrv;
                    }
                }, {
                    key: "addTemplateVariableEscapeChar",
                    value: function addTemplateVariableEscapeChar(templateSrv, escape, identifier) {
                        var regStr = templateSrv.regex.source;

                        //We have to write our regex backwards because lookbehinds aren't supported everywhere yet.
                        regStr = regStr.replace(RegExp(identifier + "\\(\\\\w\\+\\)", 'g'), "(\\w+)" + identifier + "(?=[^" + ("\\" + escape) + "]|$)");
                        templateSrv.regex = new RegExp(regStr, 'g');
                        return templateSrv;
                    }
                }, {
                    key: "transformSuggestDataFromTable",
                    value: function transformSuggestDataFromTable(suggestData) {
                        return _.map(suggestData.results['metricFindQuery'].tables[0].rows, function (v) {
                            return {
                                text: v[0],
                                value: v[1]
                            };
                        });
                    }
                }]);

                return ScalyrDatasource;
            }());

            _export("ScalyrDatasource", ScalyrDatasource);
        }
    };
});
//# sourceMappingURL=datasource.js.map
