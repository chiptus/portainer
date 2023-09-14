import { useQuery } from 'react-query';
import { FormikHandlers } from 'formik';

import { RegistryId } from '@/react/portainer/registries/types/registry';
import {
  listRegistryCatalogs,
  listRegistryCatalogsRepository,
} from '@/react/portainer/registries/registry.service';
import { useRegistries } from '@/react/portainer/registries/queries/useRegistries';

import { Option } from '@@/form-components/PortainerSelect';
import { FormSection } from '@@/form-components/FormSection';
import { TextTip } from '@@/Tip/TextTip';

import { FormValues } from './types';
import { RegistrySelector } from './RegistrySelector';

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
  const options = registriesQuery.data || [];

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
      <TextTip className="mb-2" color="blue">
        To update using a private registry, you must already be running Edge
        Agent version 2.18.1 or later. Please ensure that both the required
        agent image and the latest portainer-updater image are stored in your
        nominated private registry, as these will be the image files used during
        the upgrade process. This process is applicable in air-gapped systems or
        where Docker Hub is not accessible by internet.
      </TextTip>
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
