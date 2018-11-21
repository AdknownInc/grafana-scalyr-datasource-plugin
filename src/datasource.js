import _ from "lodash";

export class GenericDatasource {

    constructor(instanceSettings, $q, backendSrv, templateSrv) {
        this.type = instanceSettings.type;
        this.url = instanceSettings.url;
        this.name = instanceSettings.name;
        this.q = $q;
        this.backendSrv = backendSrv;
        this.templateSrv = GenericDatasource.modifyTemplateVariableIdentifier(templateSrv, '~');
        this.withCredentials = instanceSettings.withCredentials;
        this.headers = {'Content-Type': 'application/json'};
        if (typeof instanceSettings.basicAuth === 'string' && instanceSettings.basicAuth.length > 0) {
            this.headers['Authorization'] = instanceSettings.basicAuth;
        }

        this.parseComplex = instanceSettings.jsonData.parseQueries;

        this.queryControls = [];
    }

    static modifyTemplateVariableIdentifier(templateSrv, newIdentifier) {
        let regStr = templateSrv.regex.source;

        //There are 2 occurrences of '\$'. Remember to escape!
        regStr = regStr.replace(/\\\$/g, newIdentifier);
        templateSrv.regex = new RegExp(regStr, 'g');
        return templateSrv;
    }

    query(options) {
        options.targets = options.targets.filter(t => !t.hide);

        if (options.targets.length <= 0) {
            return this.q.when({data: []});
        }
        //Deep copy the object. When template variables are swapped out we don't want to modify the original values
        let query = JSON.parse(JSON.stringify(options));

        for(let i = 0; i < query.targets.length; i++) {
            let filter = query.targets[i].filter;
            query.targets[i].filter = this.templateSrv.replace(filter, null, 'regex')
        }

        query.parseComplex = this.parseComplex;

        query.user = this.backendSrv.contextSrv.user.name;
        query.userId = this.backendSrv.contextSrv.user.id;
        query.org = this.backendSrv.contextSrv.user.orgName;
        query.orgId = this.backendSrv.contextSrv.user.orgId;
        //Set in query ctrl constructor
        query.panelName = this.panelName;

        return this.doRequest({
            url: this.url + '/query',
            data: query,
            method: 'POST'
        }).then((res) => {
            //Holds on to the response so that it's accessible by the query controls
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
