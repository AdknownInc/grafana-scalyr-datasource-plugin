"use strict";

System.register([], function (_export, _context) {
  "use strict";

  var _createClass, ScalyrConfigCtrl;

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  return {
    setters: [],
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

      _export("ScalyrConfigCtrl", ScalyrConfigCtrl = function () {
        function ScalyrConfigCtrl(datasourceSrv) {
          _classCallCheck(this, ScalyrConfigCtrl);

          this.datasourceSrv = datasourceSrv;
          this.readKeyExist = this.current.secureJsonFields.hasOwnProperty("readtoken") && this.current.secureJsonFields.readtoken;
          this.writeKeyExist = this.current.secureJsonFields.hasOwnProperty("writetoken") && this.current.secureJsonFields.writetoken;
        }

        _createClass(ScalyrConfigCtrl, [{
          key: "resetReadKey",
          value: function resetReadKey() {
            this.readKeyExist = false;
          }
        }, {
          key: "resetWriteKey",
          value: function resetWriteKey() {
            this.writeKeyExist = false;
          }
        }]);

        return ScalyrConfigCtrl;
      }());

      _export("ScalyrConfigCtrl", ScalyrConfigCtrl);

      ScalyrConfigCtrl.templateUrl = 'partials/config.html';
    }
  };
});
//# sourceMappingURL=config_ctrl.js.map
