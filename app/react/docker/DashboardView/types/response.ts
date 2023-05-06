export type DockerInfoSwarmResponse = {
  NodeID: string;
  NodeAddr: string;
  LocalNodeState: string;
  ControlAvailable: boolean;
  Error: string;
};

export type DockerInfoResponse = {
  SystemTime: string;
  Swarm: DockerInfoSwarmResponse;
};
