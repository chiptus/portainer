import { number, string, object, SchemaOf } from 'yup';
import { FormikErrors } from 'formik';

import { FormSection } from '@@/form-components/FormSection';
import { RadioGroup } from '@@/RadioGroup/RadioGroup';
import { Input } from '@@/form-components/Input';
import { TextTip } from '@@/Tip/TextTip';
import { FormControl } from '@@/form-components/FormControl';
import { Button, ButtonGroup } from '@@/buttons';

import {
  StaggerConfig,
  StaggerOption,
  StaggerParallelOption,
  UpdateFailureAction,
} from '../types';

import { StaggerParallelFieldset } from './StaggerParallelFieldset';

interface Props {
  values: StaggerConfig;
  onChange: (value: Partial<StaggerConfig>) => void;
  errors?: FormikErrors<StaggerConfig>;
}

export function StaggerFieldset({ values, onChange, errors }: Props) {
  const staggerOptions = [
    {
      value: StaggerOption.AllAtOnce.toString(),
      label: 'All edge devices at once',
    },
    {
      value: StaggerOption.Parallel.toString(),
      label: 'Parallel edge device(s)',
    },
  ];

  return (
    <FormSection title="Update configurations">
      <div className="form-group">
        <div className="col-sm-12">
          <RadioGroup
            options={staggerOptions}
            selectedOption={values.StaggerOption.toString()}
            onOptionChange={(value) =>
              handleChange({ StaggerOption: parseInt(value, 10) })
            }
            name="StaggerOption"
          />
        </div>
      </div>

      {values.StaggerOption === StaggerOption.Parallel && (
        <div className="mb-2">
          <TextTip color="blue">
            Specify the number of device(s) to be updated concurrently.
            {values.StaggerParallelOption ===
              StaggerParallelOption.Incremental && (
              <div className="mb-2">
                For example, if you set this start from 2 by 5, the update will
                initially cover 2 edge devices, then expand to 10 edge devices
                (2 x 5), followed by 50 edge devices (10 x 5) and so on.
              </div>
            )}
          </TextTip>

          <StaggerParallelFieldset
            values={values}
            onChange={onChange}
            errors={errors}
          />

          <FormControl
            label="Timeout"
            inputId="timeout"
            errors={errors?.Timeout}
          >
            <Input
              name="Timeout"
              id="stagger-timeout"
              className="col-sm-3 col-lg-2"
              placeholder="eg. 5s"
              value={values.Timeout}
              onChange={(e) => handleChange({ Timeout: e.currentTarget.value })}
            />
          </FormControl>

          <FormControl
            label="Update delay"
            inputId="update-delay"
            errors={errors?.UpdateDelay}
          >
            <Input
              name="UpdateDelay"
              id="stagger-update-delay"
              className="col-sm-3 col-lg-2"
              placeholder="eg. 10s"
              value={values.UpdateDelay}
              onChange={(e) =>
                handleChange({ UpdateDelay: e.currentTarget.value })
              }
            />
          </FormControl>

          <FormControl
            label="Update failure action"
            inputId="update-failure-action"
            errors={errors?.UpdateFailureAction}
          >
            <ButtonGroup>
              <Button
                color={
                  values.UpdateFailureAction === UpdateFailureAction.Continue
                    ? 'primary'
                    : 'light'
                }
                onClick={() =>
                  handleChange({
                    UpdateFailureAction: UpdateFailureAction.Continue,
                  })
                }
              >
                Continue
              </Button>
              <Button
                color={
                  values.UpdateFailureAction === UpdateFailureAction.Pause
                    ? 'primary'
                    : 'light'
                }
                onClick={() =>
                  handleChange({
                    UpdateFailureAction: UpdateFailureAction.Pause,
                  })
                }
              >
                Pause
              </Button>
              <Button
                color={
                  values.UpdateFailureAction === UpdateFailureAction.Rollback
                    ? 'primary'
                    : 'light'
                }
                onClick={() =>
                  handleChange({
                    UpdateFailureAction: UpdateFailureAction.Rollback,
                  })
                }
              >
                Rollback
              </Button>
            </ButtonGroup>
          </FormControl>
        </div>
      )}
    </FormSection>
  );

  function handleChange(partialValue: Partial<StaggerConfig>) {
    onChange(partialValue);
  }
}

export function staggerConfigValidation(): SchemaOf<StaggerConfig> {
  return object({
    StaggerOption: number()
      .oneOf([StaggerOption.AllAtOnce, StaggerOption.Parallel])
      .required('Stagger option is required'),
    StaggerParallelOption: number()
      .when('StaggerOption', {
        is: StaggerOption.Parallel,
        then: (schema) =>
          schema.oneOf([
            StaggerParallelOption.Fixed,
            StaggerParallelOption.Incremental,
          ]),
      })
      .optional(),
    DeviceNumber: number()
      .when('StaggerOption', {
        is: StaggerOption.Parallel,
        then: (schema) =>
          schema.when('StaggerParallelOption', {
            is: StaggerParallelOption.Fixed,
            then: (schema) =>
              schema
                .min(1, 'Devices number is at least 1')
                .required('Devices number is required'),
          }),
      })
      .optional(),
    DeviceNumberStartFrom: number()
      .when('StaggerOption', {
        is: StaggerOption.Parallel,
        then: (schema) =>
          schema.when('StaggerParallelOption', {
            is: StaggerParallelOption.Incremental,
            then: (schema) =>
              schema
                .min(1, 'Devices number start from at least 1')
                .required('Devices number is required'),
          }),
      })
      .optional(),
    DeviceNumberIncrementBy: number()
      .default(2)
      .when('StaggerOption', {
        is: StaggerOption.Parallel,
        then: (schema) =>
          schema.when('StaggerParallelOption', {
            is: StaggerParallelOption.Incremental,
            then: (schema) =>
              schema
                .min(2)
                .max(10)
                .required('Devices number increment by is required'),
          }),
      })
      .optional(),
    Timeout: string()
      .default('')
      .when('StaggerOption', {
        is: StaggerOption.Parallel,
        then: (schema) => schema.required('Timeout is required'),
      })
      .optional(),
    UpdateDelay: string()
      .default('')
      .when('StaggerOption', {
        is: StaggerOption.Parallel,
        then: (schema) => schema.required('Update delay is required'),
      })
      .optional(),
    UpdateFailureAction: number()
      .default(UpdateFailureAction.Continue)
      .when('StaggerOption', {
        is: StaggerOption.Parallel,
        then: (schema) =>
          schema.oneOf([
            UpdateFailureAction.Continue,
            UpdateFailureAction.Pause,
            UpdateFailureAction.Rollback,
          ]),
      })
      .optional(),
  });
}
