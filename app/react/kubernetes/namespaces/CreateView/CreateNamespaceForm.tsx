import { Formik } from 'formik';
import { useRouter } from '@uirouter/react';

import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { notifySuccess } from '@/portainer/services/notifications';
import { useCurrentEnvironment } from '@/react/hooks/useCurrentEnvironment';
import { useEnvironmentRegistries } from '@/react/portainer/environments/queries/useEnvironmentRegistries';

import { Widget, WidgetBody } from '@@/Widget';

import { useIngressControllerClassMapQuery } from '../../cluster/ingressClass/useIngressControllerClassMap';
import { NamespaceInnerForm } from '../components/NamespaceInnerForm';
import { useNamespacesQuery } from '../queries/useNamespacesQuery';
import { StorageQuotaFormValues } from '../components/StorageQuotaFormSection/types';

import {
  CreateNamespaceFormValues,
  CreateNamespacePayload,
  UpdateRegistryPayload,
} from './types';
import { useClusterResourceLimitsQuery } from './queries/useResourceLimitsQuery';
import { getNamespaceValidationSchema } from './CreateNamespaceForm.validation';
import { transformFormValuesToNamespacePayload } from './utils';
import { useCreateNamespaceMutation } from './queries/useCreateNamespaceMutation';

export function CreateNamespaceForm() {
  const router = useRouter();
  const environmentId = useEnvironmentId();
  const { data: environment, ...environmentQuery } = useCurrentEnvironment();
  const resourceLimitsQuery = useClusterResourceLimitsQuery(environmentId);
  const { data: registries } = useEnvironmentRegistries(environmentId, {
    hideDefault: true,
  });
  // for namespace create, show ingress classes that are allowed in the current environment.
  // the ingressClasses show the none option, so we don't need to add it here.
  const { data: ingressClasses } = useIngressControllerClassMapQuery({
    environmentId,
    allowedOnly: true,
  });

  const { data: namespaces } = useNamespacesQuery(environmentId);
  const namespaceNames = Object.keys(namespaces || {});

  const createNamespaceMutation = useCreateNamespaceMutation(environmentId);

  if (resourceLimitsQuery.isLoading || environmentQuery.isLoading) {
    return null;
  }

  const storageClasses =
    environment?.Kubernetes.Configuration.StorageClasses ?? [];

  const memoryLimit = resourceLimitsQuery.data?.Memory ?? 0;

  const storageQuota: StorageQuotaFormValues[] = storageClasses.map((val) => ({
    className: val.Name,
    enabled: false,
    sizeUnit: 'M', // MB
  }));

  const initialValues: CreateNamespaceFormValues = {
    name: '',
    annotations: [],
    loadBalancerQuota: {
      enabled: false,
    },
    ingressClasses: ingressClasses ?? [],
    resourceQuota: {
      enabled: !environment?.Kubernetes.Configuration.EnableResourceOverCommit,
      memory: '0',
      cpu: '0',
    },
    registries: [],
    storageQuota,
  };

  return (
    <Widget>
      <WidgetBody>
        <Formik
          enableReinitialize
          initialValues={initialValues}
          onSubmit={handleSubmit}
          validateOnMount
          validationSchema={getNamespaceValidationSchema(
            memoryLimit,
            namespaceNames
          )}
        >
          {NamespaceInnerForm}
        </Formik>
      </WidgetBody>
    </Widget>
  );

  function handleSubmit(values: CreateNamespaceFormValues) {
    const createNamespacePayload: CreateNamespacePayload =
      transformFormValuesToNamespacePayload(values);
    const updateRegistriesPayload: UpdateRegistryPayload[] =
      values.registries.flatMap((registryFormValues) => {
        // find the matching registry from the cluster registries
        const selectedRegistry = registries?.find(
          (registry) => registryFormValues.Id === registry.Id
        );
        if (!selectedRegistry) {
          return [];
        }
        const envNamespacesWithAccess =
          selectedRegistry.RegistryAccesses[`${environmentId}`]?.Namespaces ||
          [];
        return {
          Id: selectedRegistry.Id,
          Namespaces: [...envNamespacesWithAccess, values.name],
        };
      });

    createNamespaceMutation.mutate(
      {
        createNamespacePayload,
        updateRegistriesPayload,
        namespaceIngressControllerPayload: values.ingressClasses,
      },
      {
        onSuccess: () => {
          notifySuccess(
            'Success',
            `Namespace '${values.name}' created successfully`
          );
          router.stateService.go('kubernetes.resourcePools');
        },
      }
    );
  }
}
