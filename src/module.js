import {GenericDatasource} from './datasource';
import {ScalyrDatasourceQueryCtrl} from './query_ctrl';
import {ScalyrConfigCtrl} from './config_ctrl';

class GenericQueryOptionsCtrl {}
GenericQueryOptionsCtrl.templateUrl = 'partials/query.options.html';

class GenericAnnotationsQueryCtrl {}
GenericAnnotationsQueryCtrl.templateUrl = 'partials/annotations.editor.html';

export {
  GenericDatasource as Datasource,
  ScalyrDatasourceQueryCtrl as QueryCtrl,
  ScalyrConfigCtrl as ConfigCtrl,
  GenericQueryOptionsCtrl as QueryOptionsCtrl,
  GenericAnnotationsQueryCtrl as AnnotationsQueryCtrl
};
