import { Trash2 } from 'lucide-react';
import { FormikErrors } from 'formik';
import { useState } from 'react';
import { SingleValue } from 'react-select';

import { Select } from '@@/form-components/ReactSelect';
import { InputGroup } from '@@/form-components/InputGroup';
import { Tooltip } from '@@/Tip/Tooltip';
import { Button } from '@@/buttons';
import { FormError } from '@@/form-components/FormError';

import { AddOnFormValue, AddOnOption } from './types';

interface Props {
  value: AddOnFormValue;
  onChange(newValue: AddOnFormValue): void;
  filteredOptions: AddOnOption[]; // filtered options that don't include form values
  options: AddOnOption[];
  index: number;
  onRemove(): void;
  errors?: FormikErrors<AddOnFormValue>;
}

export function AddOnSelector({
  value,
  onChange,
  options,
  filteredOptions,
  index,
  onRemove,
  errors,
}: Props) {
  const [selectedOption, setSelectedOption] = useState<
    SingleValue<AddOnOption>
  >(getSelectedOptionFromValue(options, value));
  const addonConfig = getAddonConfig(value.name);
  return (
    <div className="mb-2 flex flex-wrap gap-1">
      <div className="flex min-w-min basis-1/3 flex-col">
        <InputGroup size="small">
          <InputGroup.Addon>Addon</InputGroup.Addon>
          <Select
            name={`addons_${index}`}
            placeholder="Select an addon..."
            className="min-w-[200px]"
            options={filteredOptions}
            onChange={(option) => {
              onChange({
                name: option?.label ?? '',
                arguments: option?.arguments ?? '',
                repository: option?.repository ?? '',
              });
              setSelectedOption(option);
            }}
            value={selectedOption}
          />
        </InputGroup>
        {errors?.name && <FormError>{errors.name}</FormError>}
      </div>
      <div className="flex min-w-min basis-1/2 flex-col">
        <InputGroup size="small">
          <InputGroup.Addon>
            <div className="flex min-w-[90px] items-center">
              <span className={addonConfig?.argumentsType}>Arguments</span>
              {addonConfig?.tooltip && (
                <Tooltip message={addonConfig?.tooltip} setHtmlMessage />
              )}
            </div>
          </InputGroup.Addon>
          <InputGroup.Input
            type="string"
            className="form-control min-w-max"
            name={`arguments_${index}`}
            value={value.arguments ?? ''}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
              onChange({ ...value, arguments: e.target.value ?? '' })
            }
            data-cy={`k8sAppCreate-argument-${index}`}
            disabled={addonConfig?.argumentsType === '' || !value.name}
            placeholder={
              addonConfig?.placeholder && `e.g. ${addonConfig?.placeholder}`
            }
          />
        </InputGroup>
        {errors?.arguments && <FormError>{errors.arguments}</FormError>}
      </div>
      <div className="flex min-w-min flex-col">
        <Button
          className="btn btn-sm btn-only-icon !ml-0"
          color="dangerlight"
          type="button"
          data-cy={`k8sAppCreate-rmAddonButton_${index}`}
          onClick={onRemove}
          icon={Trash2}
        />
      </div>
    </div>
  );

  function getAddonConfig(name: string) {
    return options.find((option) => option.label === name);
  }
}

function getSelectedOptionFromValue(
  addonOptions: AddOnOption[],
  value?: AddOnFormValue
) {
  return addonOptions.find((option) => option.label === value?.name) ?? null;
}
