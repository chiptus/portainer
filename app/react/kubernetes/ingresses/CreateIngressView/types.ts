import { ReactNode } from 'react';

import {
  Annotation,
  AnnotationErrors,
} from '@/react/kubernetes/annotations/types';

import { Option } from '@@/form-components/Input/Select';

export interface Path {
  Key: string;
  Route: string;
  ServiceName: string;
  ServicePort: number;
  PathType?: string;
}

export interface Host {
  Key: string;
  Host: string;
  Secret: string;
  Paths: Path[];
  NoHost?: boolean;
}

export interface Rule {
  Key: string;
  IngressName: string;
  Namespace: string;
  IngressClassName: string;
  Hosts: Host[];
  Annotations?: Annotation[];
  IngressType?: string;
}

export interface ServicePorts {
  [serviceName: string]: Option<string>[];
}

interface ServiceOption extends Option<string> {
  selectedLabel: string;
}

export type GroupedServiceOptions = {
  label: string;
  options: ServiceOption[];
}[];

export type IngressErrors = Record<string, ReactNode> & {
  annotations?: AnnotationErrors;
};
