export type TaskId = string;

export interface DockerTaskResponse {
  ID: string;
  ServiceID: string;
}

export type TaskLogsParams = {
  stdout?: boolean;
  stderr?: boolean;
  timestamps?: boolean;
  since?: number;
  tail?: number;
};
