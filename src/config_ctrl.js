export class ScalyrConfigCtrl {

  constructor(datasourceSrv) {
    this.datasourceSrv = datasourceSrv;
    this.readLogKeyExist = this.current.secureJsonFields.hasOwnProperty("readlogtoken") && this.current.secureJsonFields.readlogtoken;
    this.readConfigKeyExist = this.current.secureJsonFields.hasOwnProperty("readconfigtoken") && this.current.secureJsonFields.readconfigtoken;
  }

  resetReadLogKey() {
    this.readLogKeyExist = false;
  }

  resetReadConfigKey() {
    this.readConfigKeyExist = false;
  }
}

ScalyrConfigCtrl.templateUrl = 'partials/config.html';
