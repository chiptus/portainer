import { FormikErrors } from 'formik';

import { Select, Input } from '@@/form-components/Input';
import { FormError } from '@@/form-components/FormError';

import { StaggerConfig, StaggerParallelOption } from '../types';

interface Props {
  values: StaggerConfig;
  onChange: (value: Partial<StaggerConfig>) => void;
  errors?: FormikErrors<StaggerConfig>;
}

export function StaggerParallelFieldset({ values, onChange, errors }: Props) {
  const staggerParallelOptions = [
    {
      value: StaggerParallelOption.Fixed.toString(),
      label: 'Number of device(s)',
    },
    {
      value: StaggerParallelOption.Incremental.toString(),
      label: 'Number of increment',
    },
  ];

  const deviceNumberIncrementBy = [
    {
      value: '2',
      label: '2',
    },
    {
      value: '5',
      label: '5',
    },
    {
      value: '10',
      label: '10',
    },
  ];

  return (
    <div
      className='form-group mb-5 mt-2 after:clear-both after:table after:content-[""]' // to fix issues with float"
    >
      <div className="col-sm-3 col-lg-2">
        <Select
          id="stagger-parallel-option"
          value={values.StaggerParallelOption?.toString()}
          onChange={(e) =>
            handleChange({
              StaggerParallelOption: parseInt(e.currentTarget.value, 10),
            })
          }
          options={staggerParallelOptions}
        />
      </div>

      {values.StaggerParallelOption === StaggerParallelOption.Fixed && (
        <div className="col-sm-9 col-lg-10">
          <Input
            name="DeviceNumber"
            type="number"
            id="device-number"
            min={1}
            placeholder="eg. 1 or 10"
            value={values.DeviceNumber}
            onChange={(e) =>
              handleChange({
                DeviceNumber: parseInt(e.currentTarget.value, 10),
              })
            }
          />
          {errors?.DeviceNumber && (
            <FormError>{errors?.DeviceNumber}</FormError>
          )}
        </div>
      )}

      {values.StaggerParallelOption === StaggerParallelOption.Incremental && (
        <div className="col-sm-9 col-lg-10">
          <div>
            <span> {' start from '} </span>
            <div style={{ display: 'inline-block', width: '150px' }}>
              <Input
                name="DeviceNumberStartFrom"
                type="number"
                id="device-number-start-from"
                min={1}
                placeholder="eg. 1"
                value={values.DeviceNumberStartFrom}
                onChange={(e) =>
                  handleChange({
                    DeviceNumberStartFrom: parseInt(e.currentTarget.value, 10),
                  })
                }
              />
            </div>
            <span> {' device(s) by '} </span>
            <Select
              id="device-number-incremental"
              value={values.DeviceNumberIncrementBy}
              style={{ display: 'inline-block', width: '150px' }}
              onChange={(e) =>
                handleChange({
                  DeviceNumberIncrementBy: parseInt(e.currentTarget.value, 10),
                })
              }
              options={deviceNumberIncrementBy}
            />
            <span>{' times '} </span>
          </div>
          {errors?.DeviceNumberStartFrom && (
            <FormError>{errors?.DeviceNumberStartFrom}</FormError>
          )}
        </div>
      )}
    </div>
  );

  function handleChange(partialValue: Partial<StaggerConfig>) {
    onChange(partialValue);
  }
}
