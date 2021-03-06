import {QueryCtrl} from 'app/plugins/sdk';

const TIME_INDEX = 1;
const INTERVAL_TYPE_WINDOW = 'window';
const INTERVAL_TYPE_FIXED = 'fixed';

export class ScalyrDatasourceQueryCtrl extends QueryCtrl {

    constructor($scope, $injector, $window, $httpParamSerializer) {
        super($scope, $injector);
        this.scope = $scope;
        this.target.filter = this.target.filter || "";
        this.target.secondsInterval = this.target.secondsInterval || 60;
        // this.target.interval = this.target.interval || 60;
        this.graphFunctions = [
            'mean', 'min', 'max', 'sumPerSecond', 'median', 'p10', 'p50', 'p95', 'p99', 'p999', 'p(n)', 'fraction', 'rate', 'count'
        ];
        this.intervalTypes = [
            INTERVAL_TYPE_WINDOW, INTERVAL_TYPE_FIXED
        ];
        this.supportedIntervalTypes = [
            'minute', 'hour', 'day', 'week'
        ];
        this.target.graphFunction = this.target.graphFunction || this.graphFunctions[0];
        this.target.expression = this.target.expression || '';
        this.target.n = this.target.n || 66;
        this.target.intervalType = this.target.intervalType || this.intervalTypes[0];
        this.target.chosenType = this.target.chosenType || this.supportedIntervalTypes[0];
        this.queryTypes = [
            'numeric query',
            'facet query',
            'complex numeric query'
        ];
        this.target.type = this.target.type || this.queryTypes[0];
        this.target.percentage = this.target.percentage || 25;
        this.target.placeholder = "target " + this.panel.targets.length;
        this.window = $window;
        this.serializer = $httpParamSerializer;

        this.datasource.queryControls.push(this);
        this.datasource.panelName = this.panel.title;
        this.target.showQueryParts = this.datasource.parseComplex;
    }

    getOptions(query) {
        return this.datasource.metricFindQuery(query || '');
    }

    toggleEditorMode() {
        this.target.rawQuery = !this.target.rawQuery;
    }

    onChangeInternal() {
        this.setTarget();
        switch (this.target.type) {
            case 'numeric query':
                if (this.target.intervalType === INTERVAL_TYPE_FIXED
                    || (this.target.intervalType === INTERVAL_TYPE_WINDOW && this.target.secondsInterval > 0))
                {
                    this.panelCtrl.refresh(); // Asks the panel to refresh data.
                }
                break;
            case 'facet query':
                this.panelCtrl.refresh(); // Asks the panel to refresh data.
                break;
            case 'complex numeric query':
                if (this.target.intervalType === INTERVAL_TYPE_FIXED
                    || (this.target.intervalType === INTERVAL_TYPE_WINDOW && this.target.secondsInterval > 0)) {
                    this.panelCtrl.refresh(); // Asks the panel to refresh data.
                }
                break;
            default:
        }

    }

    getComplexParts() {
        if(this.target.hide === true || this.target.type !== 'complex numeric query' || !this.target.showQueryParts || !this.datasource.parseComplex) {
            return;
        }

        if(this.datasource.response.data) {
            let data = this.datasource.response.data.find((element) => {
                return element.refId === this.target.refId
            });

            this.queryParts = [];
            for (let query of data.queries) {
                if(typeof query === "object") {
                    this.queryParts.push(query.function + "(" + query.filter + ")");
                } else {
                    this.queryParts.push(query);
                }
            }
        }
    }

    setTarget() {
        if (this.target.target === '' && this.target.type && this.target.graphFunction) {
            this.target.placeholder = "target for " + (this.target.type + " " + this.target.graphFunction).replace("/ /g", "_");
        }
    }

    /**
     * Opens the Scalyr logs page in a new window
     */
    openScalyrLogs() {
        if (this.target.filter !== '') {
            //convert filter to the scalyr logs page
            /**
             * Example
             * https://www.scalyr.com/events?barWidth=30%20minutes&severity=0&filter=$serverHost%3D%27app001%27&startTime=1519121858224&endTime=1519136228224
             */
            let filter = this.target.filter;
            filter = filter.replace(" = ", "=");
            filter = filter.replace(" == ", "==");

            let timeFrame = this.getTargetTimeframe(this.target.target);

            let queryParams = {
                severity: 0,
                filter: filter
            };

            queryParams = Object.assign(queryParams, timeFrame);

            let qs = this.serializer(queryParams);
            let url = 'https://www.scalyr.com/events?' + qs;
            this.window.open(url, '_blank');
        }
    }

    /**
     * Gets the first and last time values of the target data points that were returned to grafana.
     *
     * The return values should be unix_timestamp values in milliseconds
     *
     * @param target string The name of the target that is required
     * @returns {{startTime: int, endTime: int}}
     */
    getTargetTimeframe(target) {
        if (this.datasource.response.data.length === 0) {
            //return a default 24 hours if a response doesn't have data
            let now = new Date();
            return {
                startTime: now.getTime() - (1000 * 60 * 60 * 24),
                endTime: now.getTime()
            }
        }
        for (let dataSet of this.datasource.response.data) {
            if (dataSet.target === target) {
                return {
                    startTime: dataSet.datapoints[0][TIME_INDEX],
                    endTime: dataSet.datapoints.slice(-1)[0][TIME_INDEX]
                }
            }
        }
        //default to returning the from and to values of the first panel/dashboard
        let defaultData = this.datasource.response.data[0];
        return {
            startTime: defaultData.datapoints[0][TIME_INDEX],
            endTime: defaultData.datapoints.slice(-1)[0][TIME_INDEX]
        };
    }
}

ScalyrDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html';

