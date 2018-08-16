import {QueryCtrl} from 'app/plugins/sdk';
import './css/query-editor.css!'

const TIME_INDEX = 1;

export class GenericDatasourceQueryCtrl extends QueryCtrl {



    constructor($scope, $injector, $window, $httpParamSerializer) {
        super($scope, $injector);

        this.scope = $scope;


        this.target.filter = this.target.filter || "";
        this.target.secondsInterval = this.target.secondsInterval || 60;
        // this.target.interval = this.target.interval || 60;
        this.graphFunctions = [
            'mean', 'min', 'max', 'sumPerSecond', 'median', 'p10', 'p50', '95', '99', '999', 'p(n)', 'fraction', '', 'rate', 'count'
        ];
        this.target.graphFunction = this.target.graphFunction || this.graphFunctions[0];
        this.queryTypes = [
            'numeric query',
            'facet query',
            'complex numeric query'
        ];
        this.target.type = this.target.type || this.queryTypes[0];
        this.target.percentage = this.target.percentage || 25;
        this.target.placeholder = "target " + this.panel.targets.length;
        this.window = $window;
        this.serializer = $httpParamSerializer
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
             * https://www.scalyr.com/events?barWidth=30%20minutes&severity=0&filter=$serverHost%3D%27app001%27&teamToken=aasdfasdfasdfasdf--&startTime=1519121858224&endTime=1519136228224
             */
            let filter = this.target.filter;
            filter = filter.replace(" = ", "=");
            filter = filter.replace(" == ", "==");

            let timeFrame = this.getTargetTimeframe(this.target.target);


            /**
             *  teamToken is hardcoded as we only have 1 team at the moment and I didn't want to bother with a smart way of retrieving it.
             *  Also couldn't see how to retrieve it
             */
            let queryParams = {
                severity: 0,
                filter: filter,
                teamToken: 'aasdfasdfasdfasdf--'
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
        for (let dataSet of this.panelCtrl.dataList) {
            if (dataSet.target === target) {
                return {
                    startTime: dataSet.datapoints[0][TIME_INDEX],
                    endTime: dataSet.datapoints.slice(-1)[0][TIME_INDEX]
                }
            }
        }
    }
}

GenericDatasourceQueryCtrl.templateUrl = 'partials/query.editor.html';

