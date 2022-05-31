export interface FormValues {
  EnableEdgeComputeFeatures: boolean;
  EnforceEdgeID: boolean;
  EdgeAgentCheckinInterval: number;
  Edge: {
    PingInterval: number;
    SnapshotInterval: number;
    CommandInterval: number;
    AsyncMode: boolean;
  };
}
