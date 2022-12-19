import { Field, useFormikContext } from 'formik';
import { useMemo, ChangeEvent } from 'react';
import { FlaskConical } from 'lucide-react';

import {
  Credential,
  CustomTemplate,
} from '@/react/portainer/settings/cloud/types';

import { FormControl } from '@@/form-components/FormControl';
import { Input, Select } from '@@/form-components/Input';
import { TextTip } from '@@/Tip/TextTip';

import { CredentialsField } from '../shared/CredentialsField';
import { FormValues, Microk8sAddOn } from '../types';
import { useSetAvailableOption } from '../useSetAvailableOption';
import { ActionsSection } from '../shared/ActionsSection';
import { MoreSettingsSection } from '../../shared/MoreSettingsSection';
import { KaaSInfoText } from '../shared/KaaSInfoText';
import { NameField } from '../../shared/NameField';

import { Microk8sAddOnSelector } from './AddonSelector';
import { CustomTemplateSelector } from './CustomTemplateSelector';

type Props = {
  credentials: Credential[];
  customTemplates: CustomTemplate[];
  isSubmitting: boolean;
};

// ApiCreateClusterForm handles form changes, conditionally renders inputs, and manually set values
export function Microk8sCreateClusterForm({
  credentials,
  customTemplates,
  isSubmitting,
}: Props) {
  const { values, setFieldValue, errors } = useFormikContext<FormValues>();

  const { credentialId } = values;

  const credentialOptions = useMemo(
    () =>
      credentials.map((c) => ({
        value: c.id,
        label: c.name,
      })),
    [credentials]
  );

  const nodeCountOptions = [
    { value: 1, label: '1' },
    { value: 3, label: '3' },
  ];

  // ensure the form values are valid when the options change
  useSetAvailableOption(credentialOptions, credentialId, 'credentialId');

  const addonOptions = [
    { Name: 'metrics-server' },
    { Name: 'ingress' },
    { Name: 'cert-manager' },
    { Name: 'host-access' },
    { Name: 'gpu' },
    { Name: 'observability' },
    { Name: 'registry' },
  ];

  return (
    <>
      <KaaSInfoText />
      <TextTip icon={FlaskConical}>
        MicroK8s provisioning is an experimental feature. Some features may not
        work as expected.
      </TextTip>
      <NameField
        tooltip="Name of the cluster and environment."
        placeholder="e.g. my-cluster-name"
      />

      <CredentialsField credentials={credentials} />

      <FormControl
        label="Node count"
        tooltip="Number of nodes to provision in the cluster."
        inputId="kaas-nodeCount"
        errors={errors.nodeCount}
      >
        <Field
          name="nodeCount"
          id="kaas-nodeCount"
          as={Select}
          data-cy="kaasCreateForm-nodeCountInput"
          options={nodeCountOptions}
          onChange={handleChange}
        />
      </FormControl>

      <FormControl
        label="Node IP"
        tooltip="IP address of node 1."
        inputId="kaas-nodeIp-1"
      >
        <Field
          name="microk8s.nodeIP1"
          as={Input}
          type="text"
          data-cy="kaasCreateForm-nodeIp1"
          id="kaas-nodeIp-1"
        />
      </FormControl>

      {values.nodeCount === 3 && (
        <>
          <FormControl
            label="Node 2 IP"
            tooltip="IP address of node 2."
            inputId="kaas-nodeIp-2"
          >
            <Field
              name="microk8s.nodeIP2"
              as={Input}
              type="text"
              data-cy="kaasCreateForm-nodeIp2"
              id="kaas-nodeIp-2"
            />
          </FormControl>

          <FormControl
            label="Node 3 IP"
            tooltip="IP address of node 3."
            inputId="kaas-nodeIp-3"
          >
            <Field
              name="microk8s.nodeIP3"
              as={Input}
              type="text"
              data-cy="kaasCreateForm-nodeIp3"
              id="kaas-nodeIp-3"
            />
          </FormControl>
        </>
      )}

      <FormControl
        label="Add-ons"
        tooltip="You may specify add-ons to be auto installed in your cluster. The following add-ons will also be installed by default: ha-cluster, helm, helm3, hostpath-storage, dns and rbac."
        inputId="kaas-addons"
      >
        <Microk8sAddOnSelector
          value={values.microk8s.addons}
          options={addonOptions}
          onChange={(value: Microk8sAddOn[]) =>
            setFieldValue('microk8s.addons', value)
          }
        />
      </FormControl>

      <CustomTemplateSelector customTemplates={customTemplates} />

      <MoreSettingsSection>
        <TextTip color="blue">
          Metadata is only assigned to the environment in Portainer, i.e. the
          group and tags are not assigned to the cluster at the cloud provider
          end.
        </TextTip>
      </MoreSettingsSection>

      <ActionsSection
        onReloadClick={() => {}}
        isReloading={false}
        isSubmitting={isSubmitting}
      />
    </>
  );

  function handleChange(e: ChangeEvent<HTMLInputElement>) {
    const value = parseInt(e.target.value, 10);
    setFieldValue('nodeCount', Number.isNaN(value) ? '' : value);
  }
}
