import { useQuery } from 'react-query';
import { FormikHandlers } from 'formik';

import { RegistryId } from '@/react/portainer/registries/types/registry';

import { Option } from '@@/form-components/PortainerSelect';
import { FormSection } from '@@/form-components/FormSection';

import {
  listRegistryCatalogs,
  listRegistryCatalogsRepository,
} from '../../../registries/registry.service';
import { useRegistries } from '../../../registries/queries/queries';

import { FormValues } from './types';
import { RegistrySelector } from './RegistrySelector';

const defaultRegistry: Option<RegistryId> = {
  value: 0,
  label: '[Default]',
};

interface Props {
  onBlur: FormikHandlers['handleBlur'];
  value: RegistryId;
  onChange(value: RegistryId): void;
  imageVersion: FormValues['version'];
  disabled?: boolean;
}

export function AdvancedSettings({
  onBlur,
  value,
  onChange,
  imageVersion,
  disabled,
}: Props) {
  const errMessage = useValidateRegistry(value, imageVersion);
  const errorMessage = errMessage.data || '';

  const registriesQuery = useRegistries<Array<Option<RegistryId>>>({
    select: (registries) => [
      ...registries.map((registry) => ({
        value: registry.Id,
        label: registry.Name,
      })),
    ],
  });
  const options = [defaultRegistry, ...(registriesQuery.data || [])];

  return (
    <FormSection title="Advanced Settings">
      <RegistrySelector
        errorMessage={errorMessage}
        value={value}
        onBlur={onBlur}
        onChange={onChange}
        disabled={disabled}
        registries={options}
      />
    </FormSection>
  );
}

export default AdvancedSettings;

function useValidateRegistry(registryId: RegistryId, imageVersion: string) {
  return useQuery(
    [registryId, imageVersion],
    async () => {
      const catalogs = await listRegistryCatalogs(registryId);

      if (
        catalogs &&
        catalogs.repositories &&
        catalogs.repositories.includes('agent') &&
        catalogs.repositories.includes('portainer-updater')
      ) {
        // Check agent image tag
        const agentRepository = await listRegistryCatalogsRepository(
          registryId,
          'agent'
        );
        if (
          !(
            agentRepository &&
            agentRepository.tags &&
            agentRepository.tags.includes(imageVersion)
          )
        ) {
          return `The image agent:${imageVersion} cannot be found. Ensure that you upload it into the selected registry`;
        }

        // Check portainer-updater image tag
        const updaterRepository = await listRegistryCatalogsRepository(
          registryId,
          'portainer-updater'
        );
        if (
          !(
            updaterRepository &&
            updaterRepository.tags &&
            updaterRepository.tags.includes('latest')
          )
        ) {
          return 'The image portainer-updater:latest cannot be found. Ensure that you upload it into the selected registry';
        }

        return '';
      }
      return 'Either portainer-updater or agent image is missing. Ensure that you upload both images into the selected registry';
    },
    { enabled: registryId !== 0 }
  );
}
