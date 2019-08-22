import _ from "lodash";

export class ScalyrDatasource {

    constructor(instanceSettings, $q, backendSrv, templateSrv, timeSrv) {
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
        this.headers = {'Content-Type': 'application/json'};
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

    static modifyTemplateVariableIdentifier(templateSrv, newIdentifier) {
        let regStr = templateSrv.regex.source;

        //There are 2 occurrences of '\$'. Remember to escape!
        regStr = regStr.replace(/\\\$/g, newIdentifier);
        templateSrv.regex = new RegExp(regStr, 'g');
        return templateSrv;
    }

    static addTemplateVariableEscapeChar(templateSrv, escape, identifier) {
        let regStr = templateSrv.regex.source;

        //We have to write our regex backwards because lookbehinds aren't supported everywhere yet.
        regStr = regStr.replace(RegExp(identifier + "\\(\\\\w\\+\\)", 'g'), `(\\w+)${identifier}(?=[^${"\\" + escape}]|$)`);
        templateSrv.regex = new RegExp(regStr, 'g');
        return templateSrv;
    }

    removeEscapeChar(filter) {
        return filter.replace(RegExp("\\" + this.templateVarEscaperChar + this.templateVarIdentifier, 'g'), this.templateVarIdentifier);
    }

    findAndReverse(filter) {
        let newFilter = filter.reverse();
        return newFilter.replace(RegExp(`(\\w+)(?=${this.templateVarIdentifier}(?!\\${this.templateVarEscaperChar}))`, 'g'), function (a, b) {
            return b.reverse();
        });
    }

    reverseAllVariables() {
        for (let variable in this.templateSrv.index) {
            // noinspection JSUnfilteredForInLoop
            if (this.templateSrv.index[variable] instanceof Object
                && this.templateSrv.index[variable].hasOwnProperty("current")) {
                // noinspection JSUnfilteredForInLoop
                this.templateSrv.index[variable].current.value = this.templateSrv.index[variable].current.value.reverse();
            }
        }
    }

    query(options) {
        const parsedOptions = this.buildQueryParameters(options);
        parsedOptions.targets = options.targets.filter(t => !t.hide);
        if (options.targets.length <= 0) {
            return this.q.when({data: []});
        }
        //Deep copy the object. When template variables are swapped out we don't want to modify the original values
        let finalOptions = _.cloneDeep(parsedOptions);

        for (let i = 0; i < finalOptions.targets.length; i++) {
            this.reverseAllVariables();
            let filter = this.findAndReverse(finalOptions.targets[i].filter);
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
        let refIdMap = _.map(options.targets, target => target.refId);

        return this.doTsdbRequest(finalOptions)
            .then(handleTsdbResponse).then((res) => {
                res.data.sort((a, b) => {
                    return refIdMap.indexOf(a.refId) - refIdMap.indexOf(b.refId);
                });
                this.response = res;
                for (let queryControl of this.queryControls) {
                    if (queryControl.target.type === 'complex numeric query') {
                        queryControl.getComplexParts();
                    }
                }
                return res;
            });
    }

    testDatasource() {
        return this.doMetricQueryRequest('test_datasource', {}
        ).then(response => {
            return this.q.when({status: "success", message: "Data source is working", title: "Success"});
        }).catch(err => {
            return {status: "error", message: err.message, title: "Error"};
        });
    }

    metricFindQuery(query) {
        let serverHostsQuery = query.match(/^server_hosts\(\)/);
        if (serverHostsQuery) {
            return this.doMetricQueryRequest('server_hosts', {});
        }

        let logFilesQuery = query.match(/^log_files\((.+)\)/);
        if (logFilesQuery) {
            let serverHost = logFilesQuery[1];
            return this.doMetricQueryRequest('named_query_queries', {
                serverHost: this.templateSrv.replace(serverHost)
            });
        }

        return this.q.when([]);
    }

    /**
     *
     * @param options TSDBRequestOptions
     */
    doTsdbRequest(options) {
        const tsdbRequestData = {
            from: options.range.from.valueOf().toString(),
            to: options.range.to.valueOf().toString(),
            queries: options.targets,
        };

        return this.backendSrv.datasourceRequest({
            url: '/api/tsdb/query',
            method: 'POST',
            data: tsdbRequestData
        });
    }

    buildQueryParameters(options) {
        //remove placeholder targets
        options.targets = _.filter(options.targets, target => {
            return target.target !== 'select metric';
        });

        options.targets = _.map(options.targets, target => {
            return {
                datasourceId: this.datasourceId,
                refId: target.refId,
                hide: target.hide,

                queryType: 'query',
                type: target.type,
                scalyrQueryType: target.type,
                subtype: target.type || 'timeserie',
                chosenType: target.chosenType,
                target: this.templateSrv.replace(target.target, options.scopedVars, 'regex'), //the name of the query
                filter: target.filter, //the filter sent to scalyr
                graphFunction: target.graphFunction, //the type of function that is needed on Scalyr's end
                intervalType: target.intervalType,
                secondsInterval: target.secondsInterval,
                showQueryParts: target.showQueryParts
            };
        });

        return options;
    }

    doMetricQueryRequest(subtype, parameters) {
        let range = this.timeSrv.timeRange();
        return this.backendSrv.datasourceRequest({
            url: '/api/tsdb/query',
            method: 'POST',
            data: {
                from: range.from.valueOf().toString(),
                to: range.to.valueOf().toString(),
                queries: [
                    _.extend(
                        {
                            refId: 'metricFindQuery',
                            datasourceId: this.datasourceId,
                            queryType: 'metricFindQuery',
                            subtype: subtype,
                        },
                        parameters
                    ),
                ],
            }
        }).then(r => {
            return ScalyrDatasource.transformSuggestDataFromTable(r.data);
        });
    }

    static transformSuggestDataFromTable(suggestData) {
        return _.map(suggestData.results['metricFindQuery'].tables[0].rows, v => {
            return {
                text: v[0],
                value: v[1],
            };
        });
    }
}

export function handleTsdbResponse(response) {
    const res = [];
    _.forEach(response.data.results, r => {
        _.forEach(r.series, s => {
            res.push({target: s.name, datapoints: s.points, queries: r.meta, refId: r.refId});
        });
        _.forEach(r.tables, t => {
            t.type = 'table';
            t.refId = r.refId;
            res.push(t);
        });
    });

    response.data = res;
    return response;
}
