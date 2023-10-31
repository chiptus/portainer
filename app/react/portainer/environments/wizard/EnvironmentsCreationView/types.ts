export interface AnalyticsState {
  dockerAgent: number;
  dockerApi: number;
  dockerEdgeAgentStandard: number;
  dockerEdgeAgentAsync: number;
  kubernetesAgent: number;
  kubernetesEdgeAgentStandard: number;
  kubernetesEdgeAgentAsync: number;
  kaasAgent: number;
  k8sInstallAgent: number;
  aciApi: number;
  localEndpoint: number;
}

export type AnalyticsStateKey = keyof AnalyticsState;
