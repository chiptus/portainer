import { Trash2 } from 'lucide-react';
import { FormikErrors } from 'formik';
import { useState } from 'react';
import { SingleValue } from 'react-select';

import { Select } from '@@/form-components/ReactSelect';
import { InputGroup } from '@@/form-components/InputGroup';
import { Tooltip } from '@@/Tip/Tooltip';
import { Button } from '@@/buttons';
import { FormError } from '@@/form-components/FormError';
import { TextTip } from '@@/Tip/TextTip';

import { AddOnFormValue, AddOnOption } from './types';

interface Props {
  value: AddOnFormValue;
  onChange(newValue: AddOnFormValue): void;
  filteredOptions: AddOnOption[]; // filtered options that don't include form values
  options: AddOnOption[];
  index: number;
  onRemove(): void;
  isRequiredInitialArgumentEmpty?: boolean;
  errors?: FormikErrors<AddOnFormValue>;
}

export function AddOnSelector({
  value,
  onChange,
  options,
  filteredOptions,
  index,
  onRemove,
  isRequiredInitialArgumentEmpty,
  errors,
}: Props) {
  const [selectedOption, setSelectedOption] = useState<
    SingleValue<AddOnOption>
  >(getSelectedOptionFromValue(options, value));
  const addonConfig = getAddonConfig(value.name);

  return (
    <div className="flex flex-wrap gap-y-1 gap-x-2">
      <div className="inline-flex min-w-min grow basis-12 flex-col">
        <InputGroup size="small">
          <InputGroup.Addon>Addon</InputGroup.Addon>
          <Select
            name={`addons_${index}`}
            placeholder="Select an addon..."
            className="min-w-[200px] [&>div]:!rounded-r-[5px] [&>div]:!rounded-l-none"
            options={filteredOptions}
            onChange={(option) => {
              onChange({
                name: option?.label ?? '',
                arguments: option?.arguments ?? '',
                repository: option?.repository ?? '',
              });
              setSelectedOption(option);
            }}
            size="sm"
            value={selectedOption}
          />
        </InputGroup>
        {errors?.name && <FormError>{errors.name}</FormError>}
      </div>
      <div className="inline-flex min-w-min grow basis-12 gap-x-2">
        <div className="flex grow flex-col">
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
          {isRequiredInitialArgumentEmpty && (
            <TextTip color="blue">
              Arguments from addons enabled outside portainer won&apos;t show
              here.
            </TextTip>
          )}
        </div>
        <div className="flex flex-none flex-col">
          <Button
            className="!ml-0"
            size="medium"
            color="dangerlight"
            type="button"
            data-cy={`k8sAppCreate-rmAddonButton_${index}`}
            onClick={onRemove}
            icon={Trash2}
          />
        </div>
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
