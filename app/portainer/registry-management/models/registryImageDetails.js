export function RegistryImageDetailsViewModel(data) {
  if (data) {
    this.Id = data.id;
    this.Parent = data.parent;
    this.Created = data.created;
    this.DockerVersion = data.docker_version;
    this.Os = data.os;
    this.Architecture = data.architecture;
    this.Author = data.author;

    // buildx images don't have container_config property
    // besides cotainer_config's cmd value is less intuitive than the same in config's
    const config = data.config ? data.config : data.container_config;
    if (config) {
      this.Command = config.Cmd;
      this.Entrypoint = config.Entrypoint ? config.Entrypoint : '';
      this.ExposedPorts = config.ExposedPorts ? Object.keys(config.ExposedPorts) : [];
      this.Volumes = config.Volumes ? Object.keys(config.Volumes) : [];
      this.Env = config.Env ? config.Env : [];
    }
  }
}
