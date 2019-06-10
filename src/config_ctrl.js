export class ScalyrConfigCtrl {

  constructor(datasourceSrv) {
    console.log(this.current);
    this.datasourceSrv = datasourceSrv;
    this.readKeyExist = this.current.secureJsonFields.hasOwnProperty("readtoken") && this.current.secureJsonFields.readtoken;
    this.writeKeyExist = this.current.secureJsonFields.hasOwnProperty("writetoken") && this.current.secureJsonFields.writetoken;
    console.log(this);
  }

  resetReadKey() {
    this.readKeyExist = false;
  }

  resetWriteKey() {
    this.writeKeyExist = false;
  }
}

ScalyrConfigCtrl.templateUrl = 'partials/config.html';