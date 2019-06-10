'use strict';

System.register(['./datasource', './query_ctrl', './config_ctrl'], function (_export, _context) {
  "use strict";

  var GenericDatasource, ScalyrDatasourceQueryCtrl, ScalyrConfigCtrl, GenericQueryOptionsCtrl, GenericAnnotationsQueryCtrl;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  return {
    setters: [function (_datasource) {
      GenericDatasource = _datasource.GenericDatasource;
    }, function (_query_ctrl) {
      ScalyrDatasourceQueryCtrl = _query_ctrl.ScalyrDatasourceQueryCtrl;
    }, function (_config_ctrl) {
      ScalyrConfigCtrl = _config_ctrl.ScalyrConfigCtrl;
    }],
    execute: function () {
      _export('QueryOptionsCtrl', GenericQueryOptionsCtrl = function GenericQueryOptionsCtrl() {
        _classCallCheck(this, GenericQueryOptionsCtrl);
      });

      GenericQueryOptionsCtrl.templateUrl = 'partials/query.options.html';

      _export('AnnotationsQueryCtrl', GenericAnnotationsQueryCtrl = function GenericAnnotationsQueryCtrl() {
        _classCallCheck(this, GenericAnnotationsQueryCtrl);
      });

      GenericAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html';

      _export('Datasource', GenericDatasource);

      _export('QueryCtrl', ScalyrDatasourceQueryCtrl);

      _export('ConfigCtrl', ScalyrConfigCtrl);

      _export('QueryOptionsCtrl', GenericQueryOptionsCtrl);

      _export('AnnotationsQueryCtrl', GenericAnnotationsQueryCtrl);
    }
  };
});
//# sourceMappingURL=module.js.map
