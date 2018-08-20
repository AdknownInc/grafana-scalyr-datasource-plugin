'use strict';

System.register(['app/plugins/sdk', './css/query-editor.css!'], function (_export, _context) {
    "use strict";

    var QueryCtrl, _createClass, TIME_INDEX, GenericDatasourceQueryCtrl;

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
        }, function (_cssQueryEditorCss) {}],
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

            TIME_INDEX = 1;

            _export('GenericDatasourceQueryCtrl', GenericDatasourceQueryCtrl = function (_QueryCtrl) {
                _inherits(GenericDatasourceQueryCtrl, _QueryCtrl);

                function GenericDatasourceQueryCtrl($scope, $injector, $window, $httpParamSerializer) {
                    _classCallCheck(this, GenericDatasourceQueryCtrl);

                    var _this = _possibleConstructorReturn(this, (GenericDatasourceQueryCtrl.__proto__ || Object.getPrototypeOf(GenericDatasourceQueryCtrl)).call(this, $scope, $injector));

                    _this.scope = $scope;

                    _this.target.filter = _this.target.filter || "";
                    _this.target.secondsInterval = _this.target.secondsInterval || 60;
                    // this.target.interval = this.target.interval || 60;
                    _this.graphFunctions = ['mean', 'min', 'max', 'sumPerSecond', 'median', 'p10', 'p50', '95', '99', '999', 'p(n)', 'fraction', '', 'rate', 'count'];
                    _this.target.graphFunction = _this.target.graphFunction || _this.graphFunctions[0];
                    _this.queryTypes = ['numeric query', 'facet query', 'complex numeric query'];
                    _this.target.type = _this.target.type || _this.queryTypes[0];
                    _this.target.percentage = _this.target.percentage || 25;
                    _this.target.placeholder = "target " + _this.panel.targets.length;
                    _this.window = $window;
                    _this.serializer = $httpParamSerializer;
                    return _this;
                }

                _createClass(GenericDatasourceQueryCtrl, [{
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
                                if (this.target.secondsInterval > 0) {
                                    this.panelCtrl.refresh(); // Asks the panel to refresh data.
                                }
                                break;
                            case 'facet query':
                                this.panelCtrl.refresh(); // Asks the panel to refresh data.
                                break;
                            case 'complex numeric query':
                                if (this.target.secondsInterval > 0) {
                                    this.panelCtrl.refresh(); // Asks the panel to refresh data.
                                }
                                break;
                            default:
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
                        var _iteratorNormalCompletion = true;
                        var _didIteratorError = false;
                        var _iteratorError = undefined;

                        try {
                            for (var _iterator = this.panelCtrl.dataList[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                                var dataSet = _step.value;

                                if (dataSet.target === target) {
                                    return {
                                        startTime: dataSet.datapoints[0][TIME_INDEX],
                                        endTime: dataSet.datapoints.slice(-1)[0][TIME_INDEX]
                                    };
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
                }]);

                return GenericDatasourceQueryCtrl;
            }(QueryCtrl));

            _export('GenericDatasourceQueryCtrl', GenericDatasourceQueryCtrl);

            GenericDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html';
        }
    };
});
//# sourceMappingURL=query_ctrl.js.map
