import { Field, Form, Formik } from 'formik';

import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Input } from '@/portainer/components/form-components/Input';

import { validationSchema } from './validation';

interface Props {
  environmentName: string;
  setEnvironmentName: (name: string) => void;
}

export function EnvironmentNameForm({
  environmentName,
  setEnvironmentName,
}: Props) {
  return (
    <Formik
      initialValues={{ name: environmentName }}
      onSubmit={() => {}}
      validateOnMount
      validateOnChange
      enableReinitialize
      validationSchema={() => validationSchema()}
    >
      {({ errors, setFieldValue }) => (
        <Form className="form-horizontal" noValidate>
          <FormControl
            label="Name"
            tooltip="Name of the cluster and environment"
            inputId="kaas-name"
            errors={errors.name}
          >
            <Field
              name="name"
              as={Input}
              id="kaas-name"
              data-cy="kaasCreateForm-nameInput"
              placeholder="e.g. my-cluster-name"
              onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                setEnvironmentName(e.target.value);
                setFieldValue('name', e.target.value);
              }}
              value={environmentName}
            />
          </FormControl>
        </Form>
      )}
    </Formik>
  );
}
