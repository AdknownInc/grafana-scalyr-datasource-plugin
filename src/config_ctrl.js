export class ScalyrConfigCtrl {

  constructor(datasourceSrv) {
    this.datasourceSrv = datasourceSrv;
    this.readKeyExist = this.current.secureJsonFields.hasOwnProperty("readtoken") && this.current.secureJsonFields.readtoken;
    this.writeKeyExist = this.current.secureJsonFields.hasOwnProperty("writetoken") && this.current.secureJsonFields.writetoken;
  }

  resetReadKey() {
    this.readKeyExist = false;
  }

  resetWriteKey() {
    this.writeKeyExist = false;
  }
}

ScalyrConfigCtrl.templateUrl = 'partials/config.html';