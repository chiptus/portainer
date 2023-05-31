import { Annotation } from '@/react/kubernetes/annotations/types';

export interface ServicePort {
  port?: number;
  targetPort?: number;
  nodePort?: number;
  serviceName?: string;
  name?: string;
  protocol?: string;
  ingress?: object;
}

export type ServiceTypeValue = 1 | 2 | 3;

export type ServiceFormValues = {
  Headless: boolean;
  Ports: ServicePort[];
  Type: ServiceTypeValue;
  Ingress: boolean;
  ClusterIP?: string;
  ApplicationName?: string;
  ApplicationOwner?: string;
  Note?: string;
  Name?: string;
  StackName?: string;
  Annotations: Annotation[];
  Selector?: Record<string, string>;
  Namespace?: string;
};
