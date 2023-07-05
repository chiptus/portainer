import { useMemo } from 'react';
import { Form, FormikProps } from 'formik';

import {
  AddOnOption,
  Microk8sAddOnSelector,
} from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/Microk8sCreateClusterForm/AddonSelector';
import { useMicroK8sOptions } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/queries';
import { useAuthorizations } from '@/react/hooks/useUser';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';

import { Button, LoadingButton } from '@@/buttons';
import { FormControl } from '@@/form-components/FormControl';
import { TextTip } from '@@/Tip/TextTip';

import { K8sAddOnsForm } from './types';

export function AddonsForm({
  values,
  setFieldValue,
  isSubmitting,
  initialValues,
}: FormikProps<K8sAddOnsForm>) {
  const isAllowed = useAuthorizations(['K8sClusterW']);

  const environmentId = useEnvironmentId();
  const { data: isProcessing } = useEnvironment(
    environmentId,
    (env) => env?.StatusMessage?.operationStatus === 'processing'
  );

  const { data: microk8sOptions, ...microk8sOptionsQuery } =
    useMicroK8sOptions();

  const addonOptions: AddOnOption[] = useMemo(() => {
    const addonOptions: AddOnOption[] = [];
    microk8sOptions?.availableAddons.forEach((a) => {
      const kubeVersion = parseFloat(values.currentVersion.split('/')[0]);
      const versionAvailableFrom = parseFloat(a.versionAvailableFrom);
      if (kubeVersion >= versionAvailableFrom) {
        addonOptions.push({ name: a.label, type: a.type });
      }
    });

    addonOptions.sort(
      (a, b) => b.type.localeCompare(a.type) || a.name.localeCompare(b.name)
    );
    return addonOptions;
  }, [microk8sOptions?.availableAddons, values.currentVersion]);

  const requiredAddons: string[] = useMemo(
    () => microk8sOptions?.requiredAddons || [],
    [microk8sOptions?.requiredAddons]
  );

  // check if values and initial values are the same (ignore the order)
  const isInitialValues = useMemo(() => {
    if (values.addons.length !== initialValues.addons.length) {
      return false;
    }
    return values.addons.every((addon) => initialValues.addons.includes(addon));
  }, [values.addons, initialValues.addons]);

  if (microk8sOptionsQuery.isError) {
    return <TextTip color="orange">Unable to get microk8s options.</TextTip>;
  }

  return (
    <Form className="form-horizontal">
      <div className="form-group">
        <span className="col-sm-3 col-lg-2 control-label !pt-0 text-left">
          Required addons (already installed)
        </span>
        <span className="col-sm-9 col-lg-10 text-muted">
          {requiredAddons?.join(', ')}
        </span>
      </div>
      <FormControl
        label="Addons"
        tooltip={
          <>
            You may add or remove{' '}
            <a
              href="https://microk8s.io/docs/addons"
              target="_blank"
              rel="noreferrer noopener"
            >
              addons
            </a>{' '}
            here and they will be installed or uninstalled from your cluster.
          </>
        }
        inputId="microk8s-addons"
      >
        <Microk8sAddOnSelector
          value={values.addons}
          options={addonOptions}
          onChange={(value: AddOnOption[]) => {
            setFieldValue('addons', value);
          }}
          disabled={!isAllowed}
        />
      </FormControl>
      <LoadingButton
        isLoading={isSubmitting}
        loadingText="Updating addons"
        type="submit"
        color="secondary"
        className="!ml-0"
        disabled={!isAllowed || isInitialValues || isProcessing}
      >
        Apply Changes
      </LoadingButton>
      <Button
        type="reset"
        color="light"
        className="ml-1"
        disabled={isInitialValues}
      >
        Cancel
      </Button>
    </Form>
  );
}
