import { useField } from 'formik';
import { ChangeEvent, useMemo, useEffect } from 'react';

import { CustomTemplate } from '@/react/portainer/templates/custom-templates/types';
import {
  CustomTemplatesVariablesField,
  VariablesFieldValue,
  getVariablesFieldDefaultValues,
} from '@/react/portainer/custom-templates/components/CustomTemplatesVariablesField';
import { renderTemplate } from '@/react/portainer/custom-templates/components/utils';
import { getCustomTemplateFile } from '@/react/portainer/templates/custom-templates/queries/useCustomTemplateFile';

import { FormControl } from '@@/form-components/FormControl';
import { Select, Option } from '@@/form-components/Input/Select';

import { useSetAvailableOption } from '../WizardKaaS/useSetAvailableOption';
import { BetaAlert } from '../../../update-schedules/common/BetaAlert';

interface Props {
  customTemplates: CustomTemplate[];
}

export function CustomTemplateSelector({ customTemplates }: Props) {
  const customTemplateOptions: Option<number>[] = useMemo(() => {
    if (customTemplates.length === 0) {
      return [{ value: 0, label: 'No Custom Templates available' }];
    }
    const options = customTemplates.map((c) => ({
      value: c.Id,
      label: c.Title,
    }));
    return [{ value: 0, label: 'Select a Custom Template' }, ...options];
  }, [customTemplates]);

  const [fieldProps, meta, helpers] = useField<number>('meta.customTemplateId');
  useSetAvailableOption(
    customTemplateOptions,
    fieldProps.value,
    'customTemplateId'
  );

  const [templateVariables, , templateVariableHelpers] =
    useField<VariablesFieldValue>('meta.variables');
  const [, , customTemplateContentHelpers] = useField<string>(
    'meta.customTemplateContent'
  );

  const selectedTemplate = customTemplates.find(
    (c) => c.Id === fieldProps.value
  );

  useEffect(() => {
    const value = getVariablesFieldDefaultValues(
      selectedTemplate?.Variables || []
    );
    templateVariableHelpers.setValue(value);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedTemplate]);

  useEffect(() => {
    async function fetchTemplate() {
      if (selectedTemplate) {
        const fileContent = await getCustomTemplateFile({
          id: selectedTemplate.Id,
          git: !!selectedTemplate.GitConfig,
        });
        const template = renderTemplate(
          fileContent,
          templateVariables.value,
          selectedTemplate?.Variables || []
        );
        customTemplateContentHelpers.setValue(template);
      }
    }

    fetchTemplate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [templateVariables.value, selectedTemplate]);

  return (
    <FormControl
      label="Custom Template"
      tooltip={
        <>
          <div>
            Select a custom template to be installed when the environment first
            connects. A template manifest can add a namespace and applications
            to be deployed there, or just deploy to the default namespace.
          </div>
          {customTemplateOptions.length <= 1 && (
            <div className="mt-2">
              You don&apos;t currently have any, but you can create them via the
              Custom Templates menu option, once you have at least one
              Kubernetes environment set up.
            </div>
          )}
        </>
      }
      inputId="kaas-customtemplate"
      errors={meta.error}
    >
      <Select
        name={fieldProps.name}
        id="kaas-customtemplate"
        data-cy="kaasCreateForm-customtemplateSelect"
        disabled={customTemplateOptions.length <= 1}
        options={customTemplateOptions}
        value={fieldProps.value}
        onChange={handleChange}
      />
      <BetaAlert
        className="mt-2"
        message="Beta feature - so far, this has only been tested on a limited set of Kubernetes environments"
      />

      {selectedTemplate && (
        <CustomTemplatesVariablesField
          value={templateVariables.value}
          definitions={selectedTemplate?.Variables}
          onChange={handleTemplateVariablesChange}
        />
      )}
    </FormControl>
  );

  function handleTemplateVariablesChange(v: VariablesFieldValue) {
    templateVariableHelpers.setValue(v);
  }

  function handleChange(e: ChangeEvent<HTMLSelectElement>) {
    const value = parseInt(e.target.value, 10);
    if (!Number.isNaN(value)) {
      helpers.setValue(value);
    }
  }
}
