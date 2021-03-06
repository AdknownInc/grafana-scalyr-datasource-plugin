'use strict';

System.register(['app/plugins/sdk'], function (_export, _context) {
    "use strict";

    var QueryCtrl, _typeof, _createClass, TIME_INDEX, INTERVAL_TYPE_WINDOW, INTERVAL_TYPE_FIXED, ScalyrDatasourceQueryCtrl;

    function _classCallCheck(instance, Constructor) {
        if (!(instance instanceof Constructor)) {
            throw new TypeError("Cannot call a class as a function");
        }
    }

    function _possibleConstructorReturn(self, call) {
        if (!self) {
            throw new ReferenceError("this hasn't been initialised - super() hasn't been called");
        }

        return call && (typeof call === "object" || typeof call === "function") ? call : self;
    }

    function _inherits(subClass, superClass) {
        if (typeof superClass !== "function" && superClass !== null) {
            throw new TypeError("Super expression must either be null or a function, not " + typeof superClass);
        }

        subClass.prototype = Object.create(superClass && superClass.prototype, {
            constructor: {
                value: subClass,
                enumerable: false,
                writable: true,
                configurable: true
            }
        });
        if (superClass) Object.setPrototypeOf ? Object.setPrototypeOf(subClass, superClass) : subClass.__proto__ = superClass;
    }

    return {
        setters: [function (_appPluginsSdk) {
            QueryCtrl = _appPluginsSdk.QueryCtrl;
        }],
        execute: function () {
            _typeof = typeof Symbol === "function" && typeof Symbol.iterator === "symbol" ? function (obj) {
                return typeof obj;
            } : function (obj) {
                return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj;
            };

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

            TIME_INDEX = 1;
            INTERVAL_TYPE_WINDOW = 'window';
            INTERVAL_TYPE_FIXED = 'fixed';

            _export('ScalyrDatasourceQueryCtrl', ScalyrDatasourceQueryCtrl = function (_QueryCtrl) {
                _inherits(ScalyrDatasourceQueryCtrl, _QueryCtrl);

                function ScalyrDatasourceQueryCtrl($scope, $injector, $window, $httpParamSerializer) {
                    _classCallCheck(this, ScalyrDatasourceQueryCtrl);

                    var _this = _possibleConstructorReturn(this, (ScalyrDatasourceQueryCtrl.__proto__ || Object.getPrototypeOf(ScalyrDatasourceQueryCtrl)).call(this, $scope, $injector));

                    _this.scope = $scope;
                    _this.target.filter = _this.target.filter || "";
                    _this.target.secondsInterval = _this.target.secondsInterval || 60;
                    // this.target.interval = this.target.interval || 60;
                    _this.graphFunctions = ['mean', 'min', 'max', 'sumPerSecond', 'median', 'p10', 'p50', 'p95', 'p99', 'p999', 'p(n)', 'fraction', 'rate', 'count'];
                    _this.intervalTypes = [INTERVAL_TYPE_WINDOW, INTERVAL_TYPE_FIXED];
                    _this.supportedIntervalTypes = ['minute', 'hour', 'day', 'week'];
                    _this.target.graphFunction = _this.target.graphFunction || _this.graphFunctions[0];
                    _this.target.expression = _this.target.expression || '';
                    _this.target.n = _this.target.n || 66;
                    _this.target.intervalType = _this.target.intervalType || _this.intervalTypes[0];
                    _this.target.chosenType = _this.target.chosenType || _this.supportedIntervalTypes[0];
                    _this.queryTypes = ['numeric query', 'facet query', 'complex numeric query'];
                    _this.target.type = _this.target.type || _this.queryTypes[0];
                    _this.target.percentage = _this.target.percentage || 25;
                    _this.target.placeholder = "target " + _this.panel.targets.length;
                    _this.window = $window;
                    _this.serializer = $httpParamSerializer;

                    _this.datasource.queryControls.push(_this);
                    _this.datasource.panelName = _this.panel.title;
                    _this.target.showQueryParts = _this.datasource.parseComplex;
                    return _this;
                }

                _createClass(ScalyrDatasourceQueryCtrl, [{
                    key: 'getOptions',
                    value: function getOptions(query) {
                        return this.datasource.metricFindQuery(query || '');
                    }
                }, {
                    key: 'toggleEditorMode',
                    value: function toggleEditorMode() {
                        this.target.rawQuery = !this.target.rawQuery;
                    }
                }, {
                    key: 'onChangeInternal',
                    value: function onChangeInternal() {
                        this.setTarget();
                        switch (this.target.type) {
                            case 'numeric query':
                                if (this.target.intervalType === INTERVAL_TYPE_FIXED || this.target.intervalType === INTERVAL_TYPE_WINDOW && this.target.secondsInterval > 0) {
                                    this.panelCtrl.refresh(); // Asks the panel to refresh data.
                                }
                                break;
                            case 'facet query':
                                this.panelCtrl.refresh(); // Asks the panel to refresh data.
                                break;
                            case 'complex numeric query':
                                if (this.target.intervalType === INTERVAL_TYPE_FIXED || this.target.intervalType === INTERVAL_TYPE_WINDOW && this.target.secondsInterval > 0) {
                                    this.panelCtrl.refresh(); // Asks the panel to refresh data.
                                }
                                break;
                            default:
                        }
                    }
                }, {
                    key: 'getComplexParts',
                    value: function getComplexParts() {
                        var _this2 = this;

                        if (this.target.hide === true || this.target.type !== 'complex numeric query' || !this.target.showQueryParts || !this.datasource.parseComplex) {
                            return;
                        }

                        if (this.datasource.response.data) {
                            var data = this.datasource.response.data.find(function (element) {
                                return element.refId === _this2.target.refId;
                            });

                            this.queryParts = [];
                            var _iteratorNormalCompletion = true;
                            var _didIteratorError = false;
                            var _iteratorError = undefined;

                            try {
                                for (var _iterator = data.queries[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                                    var query = _step.value;

                                    if ((typeof query === 'undefined' ? 'undefined' : _typeof(query)) === "object") {
                                        this.queryParts.push(query.function + "(" + query.filter + ")");
                                    } else {
                                        this.queryParts.push(query);
                                    }
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
                        }
                    }
                }, {
                    key: 'setTarget',
                    value: function setTarget() {
                        if (this.target.target === '' && this.target.type && this.target.graphFunction) {
                            this.target.placeholder = "target for " + (this.target.type + " " + this.target.graphFunction).replace("/ /g", "_");
                        }
                    }
                }, {
                    key: 'openScalyrLogs',
                    value: function openScalyrLogs() {
                        if (this.target.filter !== '') {
                            //convert filter to the scalyr logs page
                            /**
                             * Example
                             * https://www.scalyr.com/events?barWidth=30%20minutes&severity=0&filter=$serverHost%3D%27app001%27&startTime=1519121858224&endTime=1519136228224
                             */
                            var filter = this.target.filter;
                            filter = filter.replace(" = ", "=");
                            filter = filter.replace(" == ", "==");

                            var timeFrame = this.getTargetTimeframe(this.target.target);

                            var queryParams = {
                                severity: 0,
                                filter: filter
                            };

                            queryParams = Object.assign(queryParams, timeFrame);

                            var qs = this.serializer(queryParams);
                            var url = 'https://www.scalyr.com/events?' + qs;
                            this.window.open(url, '_blank');
                        }
                    }
                }, {
                    key: 'getTargetTimeframe',
                    value: function getTargetTimeframe(target) {
                        if (this.datasource.response.data.length === 0) {
                            //return a default 24 hours if a response doesn't have data
                            var now = new Date();
                            return {
                                startTime: now.getTime() - 1000 * 60 * 60 * 24,
                                endTime: now.getTime()
                            };
                        }
                        var _iteratorNormalCompletion2 = true;
                        var _didIteratorError2 = false;
                        var _iteratorError2 = undefined;

                        try {
                            for (var _iterator2 = this.datasource.response.data[Symbol.iterator](), _step2; !(_iteratorNormalCompletion2 = (_step2 = _iterator2.next()).done); _iteratorNormalCompletion2 = true) {
                                var dataSet = _step2.value;

                                if (dataSet.target === target) {
                                    return {
                                        startTime: dataSet.datapoints[0][TIME_INDEX],
                                        endTime: dataSet.datapoints.slice(-1)[0][TIME_INDEX]
                                    };
                                }
                            }
                            //default to returning the from and to values of the first panel/dashboard
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

                        var defaultData = this.datasource.response.data[0];
                        return {
                            startTime: defaultData.datapoints[0][TIME_INDEX],
                            endTime: defaultData.datapoints.slice(-1)[0][TIME_INDEX]
                        };
                    }
                }]);

                return ScalyrDatasourceQueryCtrl;
            }(QueryCtrl));

            _export('ScalyrDatasourceQueryCtrl', ScalyrDatasourceQueryCtrl);

            ScalyrDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html';
        }
    };
});
//# sourceMappingURL=query_ctrl.js.map
