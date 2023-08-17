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
  isEdit?: boolean;
}

export function StaggerFieldset({
  values,
  onChange,
  errors,
  isEdit = true,
}: Props) {
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
      {!isEdit && (
        <div className="form-group">
          <div className="col-sm-12">
            <TextTip color="blue">
              Please note that the &apos;Update Configuration&apos; setting
              takes effect exclusively during edge stack updates, whether
              triggered manually, through webhook events, or via GitOps updates
              processes
            </TextTip>
          </div>
        </div>
      )}

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
                For example, if you start with 2 devices and multiply by 5, the
                update will initially cover 2 edge devices, then 10 devices (2 x
                5), followed by 50 devices (10 x 5), and so on.
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
            <div>
              <div style={{ display: 'inline-block', width: '150px' }}>
                <Input
                  name="Timeout"
                  id="stagger-timeout"
                  placeholder="eg. 5 (optional)"
                  value={values.Timeout}
                  onChange={(e) =>
                    handleChange({
                      Timeout: e.currentTarget.value,
                    })
                  }
                />
              </div>
              <span> {' minute(s) '} </span>
            </div>
          </FormControl>

          <FormControl
            label="Update delay"
            inputId="update-delay"
            errors={errors?.UpdateDelay}
          >
            <div>
              <div style={{ display: 'inline-block', width: '150px' }}>
                <Input
                  name="UpdateDelay"
                  id="stagger-update-delay"
                  placeholder="eg. 5 (optional)"
                  value={values.UpdateDelay}
                  onChange={(e) =>
                    handleChange({
                      UpdateDelay: e.currentTarget.value,
                    })
                  }
                />
              </div>
              <span> {' minute(s) '} </span>
            </div>
          </FormControl>

          <FormControl
            label="Update failure action"
            inputId="update-failure-action"
            errors={errors?.UpdateFailureAction}
          >
            <ButtonGroup>
              <Button
                className="btn-box-shadow"
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
                className="btn-box-shadow"
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
                className="btn-box-shadow"
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
      .default(0)
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
        then: (schema) =>
          schema.test(
            'is-number',
            'Timeout must be a number',
            (value) => !Number.isNaN(Number(value))
          ),
      })
      .optional(),
    UpdateDelay: string()
      .default('')
      .when('StaggerOption', {
        is: StaggerOption.Parallel,
        then: (schema) =>
          schema.test(
            'is-number',
            'Timeout must be a number',
            (value) => !Number.isNaN(Number(value))
          ),
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