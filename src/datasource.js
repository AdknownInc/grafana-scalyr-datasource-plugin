import _ from "lodash";

export class GenericDatasource {

    constructor(instanceSettings, $q, backendSrv, templateSrv) {
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
        return filter.replace(RegExp("\\" + this.templateVarEscaperChar + this.templateVarIdentifier,'g'), this.templateVarIdentifier);
    }

    findAndReverse(filter) {
        let newFilter = filter.reverse();
        return newFilter.replace(RegExp(`(\\w+)(?=${this.templateVarIdentifier}(?!\\${this.templateVarEscaperChar}))`,'g'), function (a,b){
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
        options.targets = options.targets.filter(t => !t.hide);
        if (options.targets.length <= 0) {
            return this.q.when({data: []});
        }
        //Deep copy the object. When template variables are swapped out we don't want to modify the original values
        let query = JSON.parse(JSON.stringify(options));

        for(let i = 0; i < query.targets.length; i++) {
            this.reverseAllVariables();
            let filter = this.findAndReverse(query.targets[i].filter);
            query.targets[i].filter = this.findAndReverse(this.templateSrv.replace(filter, null, 'regex'));
            query.targets[i].filter = this.removeEscapeChar(query.targets[i].filter);
            this.reverseAllVariables();
        }

        query.parseComplex = this.parseComplex;

        query.user = this.backendSrv.contextSrv.user.name;
        query.userId = this.backendSrv.contextSrv.user.id;
        query.org = this.backendSrv.contextSrv.user.orgName;
        query.orgId = this.backendSrv.contextSrv.user.orgId;
        //Set in query ctrl constructor
        query.panelName = this.panelName;

        const tsdbRequest = {
            from: options.range.from.valueOf().toString(),
            to: options.range.to.valueOf().toString(),
            queries: [{
                datasourceId: this.datasourceId,
                backendUse: query,
            }]
        };

        return this.backendSrv.datasourceRequest({
            url: '/api/tsdb/query',
            method: 'POST',
            data: tsdbRequest
        }).then(handleTsdbResponse).then((res) => {
            this.response = res;
            for(let queryControl of this.queryControls) {
                queryControl.getComplexParts();
            }
            return res;
        } );
    }

    testDatasource() {
        return this.doRequest({
            url: this.url + '/',
            method: 'GET',
        }).then(response => {
            if (response.status === 200) {
                return {status: "success", message: "Data source is working", title: "Success"};
            }
        });
    }

    annotationQuery(options) {
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
        }).then(result => {
            return result.data;
        });
    }

    metricFindQuery(query) {
        var interpolated = {
            target: this.templateSrv.replace(query, null, 'regex')
        };

        return this.doRequest({
            url: this.url + '/search',
            data: interpolated,
            method: 'POST',
        }).then(this.mapToTextValue);
    }

    mapToTextValue(result) {
        return _.map(result.data, (d, i) => {
            if (d && d.text && d.value) {
                return {text: d.text, value: d.value};
            } else if (_.isObject(d)) {
                return {text: d, value: i};
            }
            return {text: d, value: d};
        });
    }

    doRequest(options) {
        options.withCredentials = this.withCredentials;
        options.headers = this.headers;

        this.options = options;

        return this.backendSrv.datasourceRequest(options);
    }

    buildQueryParameters(options) {
        //remove placeholder targets
        options.targets = _.filter(options.targets, target => {
            return target.target !== 'select metric';
        });

        var targets = _.map(options.targets, target => {
            return {
                target: this.templateSrv.replace(target.target, options.scopedVars, 'regex'),
                refId: target.refId,
                hide: target.hide,
                type: target.type || 'timeserie'
            };
        });

        options.targets = targets;

        return options;
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
