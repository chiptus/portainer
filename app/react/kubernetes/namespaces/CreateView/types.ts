import { Registry } from '@/react/portainer/registries/types';

import { Annotation, AnnotationsPayload } from '../../annotations/types';
import { IngressControllerClassMap } from '../../cluster/ingressClass/types';
import { ResourceQuotaFormValues } from '../components/ResourceQuotaFormSection/types';
import { LoadBalancerQuotaFormValues } from '../components/LoadBalancerFormSection/types';
import {
  StorageQuotaFormValues,
  StorageQuotaPayload,
} from '../components/StorageQuotaFormSection/types';

export type CreateNamespaceFormValues = {
  name: string;
  annotations: Annotation[];
  resourceQuota: ResourceQuotaFormValues;
  loadBalancerQuota: LoadBalancerQuotaFormValues;
  ingressClasses: IngressControllerClassMap[];
  registries: Registry[];
  storageQuota: StorageQuotaFormValues[];
};

export type CreateNamespacePayload = {
  Name: string;
  Owner: string;
  Annotations: AnnotationsPayload;
  ResourceQuota: ResourceQuotaFormValues;
  LoadBalancerQuota: LoadBalancerQuotaFormValues;
  StorageQuotas: StorageQuotaPayload;
};

export type UpdateRegistryPayload = {
  Id: number;
  Namespaces: string[];
};
