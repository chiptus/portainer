import _ from 'lodash-es';

export function RegistryImageLayerViewModel(order, data) {
  this.Order = order;
  this.Id = data.id;
  this.Created = data.created;

  // buildx images uses config property instead of container_config
  const config = data.config ? data.config : data.container_config;
  if (config) {
    this.CreatedBy = _.join(config.Cmd, ' ');
  }
}
