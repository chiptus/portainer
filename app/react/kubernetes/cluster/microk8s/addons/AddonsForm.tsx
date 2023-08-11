import { useEffect, useMemo, useState } from 'react';
import { Form, FormikProps } from 'formik';
import { Plus } from 'lucide-react';
import clsx from 'clsx';

import { useMicroK8sOptions } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/queries';
import { useAuthorizations } from '@/react/hooks/useUser';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { useEnvironment } from '@/react/portainer/environments/queries';
import { isErrorType } from '@/react/kubernetes/applications/CreateView/application-services/utils';
import { OperationStatus } from '@/react/portainer/environments/types';

import { Button, LoadingButton } from '@@/buttons';
import { TextTip } from '@@/Tip/TextTip';
import { InlineLoader } from '@@/InlineLoader';

import {
  AddOnFormValue,
  AddOnOption,
  GroupedAddonOptions,
  K8sAddOnsForm,
} from './types';
import { AddOnSelector } from './AddonSelector';

type Props = {
  isRefetchingAddons: boolean;
};

export function AddonsForm({
  values,
  errors,
  isValid,
  setFieldValue,
  isSubmitting,
  initialValues,
  isRefetchingAddons,
  resetForm,
}: FormikProps<K8sAddOnsForm> & Props) {
  const [currentOperationStatus, setCurrentOperationStatus] = useState<
    OperationStatus | undefined
  >('');
  const isAllowed = useAuthorizations(['K8sClusterW']);
  const environmentId = useEnvironmentId();
  const { data: environment } = useEnvironment(environmentId);

  const operationStatus = useMemo(
    () => environment?.StatusMessage?.operationStatus,
    [environment?.StatusMessage?.operationStatus]
  );
  const isProcessing = operationStatus === 'processing';

  const { data: microk8sOptions, ...microk8sOptionsQuery } =
    useMicroK8sOptions();

  const [addonOptions, groupedAddonOptions] = useMemo(() => {
    const kubeVersion = parseFloat(values.currentVersion.split('/')[0]);
    const addonOptions: AddOnOption[] =
      microk8sOptions?.availableAddons
        // filter addons that are not available for the current kubernetes version
        .filter((a) => {
          const versionAvailableFrom = parseFloat(a.versionAvailableFrom);
          const versionAvailableTo = parseFloat(a.versionAvailableTo);

          return (
            kubeVersion >= versionAvailableFrom &&
            (Number.isNaN(versionAvailableTo) ||
              kubeVersion <= versionAvailableTo)
          );
        })
        .map((a) => ({
          ...a,
          name: a.label,
        })) ?? [];

    const addonOptionsWithoutExistingValues = addonOptions.filter(
      (addonOption) =>
        !values.addons.some((addon) => addon.name === addonOption.name)
    );

    const groupedAddonOptions =
      addonOptionsWithoutExistingValues
        .reduce<GroupedAddonOptions>(
          (groupedOptions, addon) => {
            // add the current addon option to the groupOption with the same repository
            const newGroupedOptions = groupedOptions.map((group) => {
              if (group.label.toLowerCase() === addon.repository) {
                return {
                  ...group,
                  options: [...group.options, addon],
                };
              }
              return group;
            });
            return newGroupedOptions;
          },
          [
            { label: 'Core', options: [] },
            { label: 'Community', options: [] },
          ]
        )
        .filter((optionsGroup) => optionsGroup.options.length) || [];

    return [addonOptions, groupedAddonOptions];
  }, [microk8sOptions?.availableAddons, values.addons, values.currentVersion]);

  // check if values and initial values are the same (ignore the order)
  const isInitialValues = useMemo(() => {
    if (values.addons.length !== initialValues.addons.length) {
      return false;
    }
    const initialValuesNames = initialValues.addons.map((a) => a.name);
    const initialValuesArgs = initialValues.addons.map((a) => a.arguments);
    return values.addons.every(
      (addon) =>
        initialValuesNames.includes(addon.name) &&
        initialValuesArgs.includes(addon.arguments)
    );
  }, [values.addons, initialValues.addons]);

  useEffect(() => {
    // when the operation finishes, reset the form
    if (
      currentOperationStatus === 'processing' &&
      operationStatus !== 'processing'
    ) {
      resetForm();
    }
    setCurrentOperationStatus(operationStatus);
  }, [currentOperationStatus, environmentId, operationStatus, resetForm]);

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
          {microk8sOptions?.requiredAddons?.join(', ')}
        </span>
      </div>
      <div className="form-group">
        <span className="col-sm-12 control-label !pt-0 text-left">
          Optional addons
        </span>
      </div>

      <div className="mb-2 flex w-full flex-col gap-y-2">
        {values.addons.map((addon, index) => {
          const error = errors.addons?.[index];
          const addonError = isErrorType<AddOnFormValue>(error)
            ? error
            : undefined;
          const initialAddonMatching = initialValues?.addons.find(
            (addonOption) => addonOption.name === addon.name
          );
          const matchingAddonOption = microk8sOptions?.availableAddons.find(
            (addonOption) => addonOption.label === addon.name
          );
          const isRequiredInitialArgumentEmpty =
            initialAddonMatching?.arguments === '' &&
            matchingAddonOption?.argumentsType === 'required';
          return (
            <AddOnSelector
              // include the index in the key too, because the name is empty until an addon is selected
              key={`addon${addon.name}${index}`}
              value={addon}
              options={addonOptions}
              groupedAddonOptions={groupedAddonOptions}
              isRequiredInitialArgumentEmpty={isRequiredInitialArgumentEmpty}
              index={index}
              errors={addonError}
              onChange={(value: AddOnFormValue) => {
                const addons = [...values.addons];
                addons[index] = value;
                setFieldValue('addons', addons);
              }}
              onRemove={() => {
                const addons = [...values.addons];
                addons.splice(index, 1);
                setFieldValue('addons', addons);
              }}
              isProcessing={isProcessing}
              readonly={!isAllowed}
            />
          );
        })}
      </div>
      <div className="mb-5 flex items-center gap-x-4 pt-2">
        {isAllowed && (
          <Button
            className="btn btn-sm btn-light !ml-0"
            type="button"
            onClick={addAddon}
            icon={Plus}
            disabled={isProcessing}
          >
            Add addon
          </Button>
        )}
        <div
          // transition between opacity 0 and 1 for less jarring changes
          className={clsx(
            'flex w-full transition-opacity',
            isRefetchingAddons ? 'opacity-100' : 'opacity-0'
          )}
        >
          <InlineLoader size="xs">Refreshing addons...</InlineLoader>
        </div>
      </div>
      {isAllowed && (
        <>
          <LoadingButton
            isLoading={isSubmitting}
            loadingText="Updating addons"
            type="submit"
            color="secondary"
            className="!ml-0"
            disabled={!isAllowed || isInitialValues || isProcessing || !isValid}
          >
            Apply Changes
          </LoadingButton>
          <Button
            type="reset"
            color="light"
            className="ml-1"
            disabled={isInitialValues || isProcessing}
          >
            Cancel
          </Button>
        </>
      )}
    </Form>
  );

  function addAddon() {
    // Clone the existing addons array to avoid mutating the original
    const addons = structuredClone(values.addons);
    addons.push({
      name: '',
      arguments: '',
    });

    // Update the form values with the new addons array
    setFieldValue('addons', addons);
  }
}
