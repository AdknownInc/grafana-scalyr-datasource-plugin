"use strict";

System.register(["lodash"], function (_export, _context) {
    "use strict";

    var _, _createClass, GenericDatasource;

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

            _export("GenericDatasource", GenericDatasource = function () {
                function GenericDatasource(instanceSettings, $q, backendSrv, templateSrv) {
                    _classCallCheck(this, GenericDatasource);

                    this.datasourceId = instanceSettings.id;
                    this.type = instanceSettings.type;
                    this.url = instanceSettings.url;
                    this.name = instanceSettings.name;
                    this.q = $q;
                    this.backendSrv = backendSrv;
                    this.templateVarIdentifier = '~';
                    this.templateVarEscaperChar = "\\";
                    this.templateSrv = GenericDatasource.modifyTemplateVariableIdentifier(templateSrv, this.templateVarIdentifier);
                    this.templateSrv = GenericDatasource.addTemplateVariableEscapeChar(this.templateSrv, this.templateVarEscaperChar, this.templateVarIdentifier);
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

                _createClass(GenericDatasource, [{
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

                        options.targets = options.targets.filter(function (t) {
                            return !t.hide;
                        });
                        if (options.targets.length <= 0) {
                            return this.q.when({ data: [] });
                        }
                        //Deep copy the object. When template variables are swapped out we don't want to modify the original values
                        var query = JSON.parse(JSON.stringify(options));

                        for (var i = 0; i < query.targets.length; i++) {
                            this.reverseAllVariables();
                            var filter = this.findAndReverse(query.targets[i].filter);
                            query.targets[i].filter = this.findAndReverse(this.templateSrv.replace(filter, null, 'regex'));
                            query.targets[i].filter = this.removeEscapeChar(query.targets[i].filter);
                            this.reverseAllVariables();

                            //Run through forwards for square bracket variable syntax
                            query.targets[i].filter = this.templateSrv.replace(query.targets[i].filter, null, 'regex');

                            //Grafana adds regex escapes to the variables for some reason
                            query.targets[i].filter = query.targets[i].filter.replace(/\\(.)/g, "$1");
                        }

                        query.parseComplex = this.parseComplex;

                        query.user = this.backendSrv.contextSrv.user.name;
                        query.userId = this.backendSrv.contextSrv.user.id;
                        query.org = this.backendSrv.contextSrv.user.orgName;
                        query.orgId = this.backendSrv.contextSrv.user.orgId;
                        //Set in query ctrl constructor
                        query.panelName = this.panelName;

                        var tsdbRequest = {
                            from: options.range.from.valueOf().toString(),
                            to: options.range.to.valueOf().toString(),
                            queries: [{
                                datasourceId: this.datasourceId,
                                backendUse: query
                            }]
                        };

                        //This is needed because Grafana messes up the ordering when moving the response from backend to frontend
                        var refIdMap = [];

                        var _iteratorNormalCompletion = true;
                        var _didIteratorError = false;
                        var _iteratorError = undefined;

                        try {
                            for (var _iterator = query.targets[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                                var target = _step.value;

                                refIdMap.push(target.refId);
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

                        return this.backendSrv.datasourceRequest({
                            url: '/api/tsdb/query',
                            method: 'POST',
                            data: tsdbRequest
                        }).then(handleTsdbResponse).then(function (res) {
                            res.data.sort(function (a, b) {
                                return refIdMap.indexOf(a.refId) - refIdMap.indexOf(b.refId);
                            });
                            _this.response = res;
                            var _iteratorNormalCompletion2 = true;
                            var _didIteratorError2 = false;
                            var _iteratorError2 = undefined;

                            try {
                                for (var _iterator2 = _this.queryControls[Symbol.iterator](), _step2; !(_iteratorNormalCompletion2 = (_step2 = _iterator2.next()).done); _iteratorNormalCompletion2 = true) {
                                    var queryControl = _step2.value;

                                    queryControl.getComplexParts();
                                }
                            } catch (err) {
                                _didIteratorError2 = true;
                                _iteratorError2 = err;
                            } finally {
                                try {
                                    if (!_iteratorNormalCompletion2 && _iterator2.return) {
                                        _iterator2.return();
                                    }
                                } finally {
                                    if (_didIteratorError2) {
                                        throw _iteratorError2;
                                    }
                                }
                            }

                            return res;
                        });

                        //#region old way
                        //kept in for faster reverts if need be
                        // return this.doRequest({
                        //     url: this.url + '/query',
                        //     data: query,
                        //     method: 'POST'
                        // }).then((res) => {
                        //     //Holds on to the response so that it's accessible by the query controls
                        //     this.response = res;
                        //     for(let queryControl of this.queryControls) {
                        //         queryControl.getComplexParts();
                        //     }
                        //     return res;
                        // } );
                        //#endregion
                    }
                }, {
                    key: "testDatasource",
                    value: function testDatasource() {
                        return this.doRequest({
                            url: this.url + '/',
                            method: 'GET'
                        }).then(function (response) {
                            if (response.status === 200) {
                                return { status: "success", message: "Data source is working", title: "Success" };
                            }
                        });
                    }
                }, {
                    key: "annotationQuery",
                    value: function annotationQuery(options) {
                        var query = this.templateSrv.replace(options.annotation.query, {}, 'glob');
                        var annotationQuery = {
                            range: options.range,
                            annotation: {
                                name: options.annotation.name,
                                datasource: options.annotation.datasource,
                                enable: options.annotation.enable,
                                iconColor: options.annotation.iconColor,
                                query: query
                            },
                            rangeRaw: options.rangeRaw
                        };

                        return this.doRequest({
                            url: this.url + '/annotations',
                            method: 'POST',
                            data: annotationQuery
                        }).then(function (result) {
                            return result.data;
                        });
                    }
                }, {
                    key: "metricFindQuery",
                    value: function metricFindQuery(query) {
                        var interpolated = {
                            target: this.templateSrv.replace(query, null, 'regex')
                        };

                        return this.doRequest({
                            url: this.url + '/search',
                            data: interpolated,
                            method: 'POST'
                        }).then(this.mapToTextValue);
                    }
                }, {
                    key: "mapToTextValue",
                    value: function mapToTextValue(result) {
                        return _.map(result.data, function (d, i) {
                            if (d && d.text && d.value) {
                                return { text: d.text, value: d.value };
                            } else if (_.isObject(d)) {
                                return { text: d, value: i };
                            }
                            return { text: d, value: d };
                        });
                    }
                }, {
                    key: "doRequest",
                    value: function doRequest(options) {
                        options.withCredentials = this.withCredentials;
                        options.headers = this.headers;

                        this.options = options;

                        return this.backendSrv.datasourceRequest(options);
                    }
                }, {
                    key: "buildQueryParameters",
                    value: function buildQueryParameters(options) {
                        var _this2 = this;

                        //remove placeholder targets
                        options.targets = _.filter(options.targets, function (target) {
                            return target.target !== 'select metric';
                        });

                        var targets = _.map(options.targets, function (target) {
                            return {
                                target: _this2.templateSrv.replace(target.target, options.scopedVars, 'regex'),
                                refId: target.refId,
                                hide: target.hide,
                                type: target.type || 'timeserie'
                            };
                        });

                        options.targets = targets;

                        return options;
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
                }]);

                return GenericDatasource;
            }());

            _export("GenericDatasource", GenericDatasource);
        }
    };
});
//# sourceMappingURL=datasource.js.map
