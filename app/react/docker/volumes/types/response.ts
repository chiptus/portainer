export type DockerVolumeResponse = {
  CreatedAt?: string;
  Driver: string;
  Labels: { [key: string]: string };
  Mountpoint: string;
  Name: string;
  Options: { [key: string]: string };
  Scope: string;
};
