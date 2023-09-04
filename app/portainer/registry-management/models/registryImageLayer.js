import _ from 'lodash-es';

export function RegistryImageLayerViewModel(order, history) {
  this.Order = order;
  this.Id = history.id;
  this.Created = history.created;

  if (history.CreatedBy) {
    // image configs blob history comes with created_by field which then is renamed to CreateBy
    this.CreatedBy = history.CreatedBy;
  } else {
    // manifest v1 history comse without CreateBy field
    // assemble CreateBy field from the Cmd field
    // buildx images uses config property instead of container_config
    const config = history.config ? history.config : history.container_config;
    if (config) {
      this.CreatedBy = _.join(config.Cmd, ' ');
    }
  }
}
