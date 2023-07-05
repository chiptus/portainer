import { Formik, Field, Form } from 'formik';
import { Award } from 'lucide-react';
import { useReducer } from 'react';
import { useMutation, useQueryClient } from 'react-query';
import { useRouter } from '@uirouter/react';

import { pluralize } from '@/portainer/helpers/strings';
import { notifySuccess } from '@/portainer/services/notifications';
import { withError } from '@/react-tools/react-query';

import { PageHeader } from '@@/PageHeader';
import { FormControl } from '@@/form-components/FormControl';
import { Widget } from '@@/Widget';
import { Input } from '@@/form-components/Input';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { confirmDestructive } from '@@/modals/confirm';
import { buildConfirmButton } from '@@/modals/utils';

import { attachLicense } from '../license.service';

import { validationSchema } from './CreateView.validation';

interface FormValues {
  key: string;
}

interface Config {
  params: {
    force?: boolean;
  };
}

interface AddFormType {
  values: FormValues;
  config?: Config;
}

const initialValues = {
  key: '',
};

export function CreateView() {
  const addLicenseMutation = useAddLicenseMutation();
  const router = useRouter();

  const [formKey, incFormKey] = useReducer((state: number) => state + 1, 0);

  return (
    <>
      <PageHeader
        title="Add License"
        breadcrumbs={[
          {
            link: 'portainer.licenses',
            label: 'Licenses',
          },
          {
            label: 'Add Licenses',
          },
        ]}
      />
      <div className="row">
        <div className="col-lg-12 col-md-12 col-xs-12">
          <Widget>
            <Widget.Title
              icon={Award}
              title="License registration"
              className="vertical-center"
            />
            <Widget.Body>
              <Formik
                initialValues={initialValues}
                validationSchema={() => validationSchema()}
                onSubmit={handleSubmitClick}
                key={formKey}
                validateOnMount
              >
                {({ errors, handleSubmit, isValid }) => (
                  <Form
                    className="form-horizontal"
                    onSubmit={handleSubmit}
                    noValidate
                  >
                    <FormControl
                      inputId="licenseKey"
                      label="License Key"
                      errors={errors.key}
                      required
                    >
                      <Field
                        as={Input}
                        name="key"
                        id="licenseKey"
                        placeholder="Your license key should start with “2-” or “3-”."
                        required
                        data-cy="licenseKey"
                      />
                    </FormControl>

                    <div className="form-group">
                      <div className="col-sm-12">
                        <LoadingButton
                          disabled={!isValid}
                          data-cy="licenseInputSubmit"
                          isLoading={addLicenseMutation.isLoading}
                          loadingText="Submitting..."
                          className="!ml-0"
                        >
                          Submit
                        </LoadingButton>
                      </div>
                    </div>
                  </Form>
                )}
              </Formik>
            </Widget.Body>
          </Widget>
        </div>
      </div>
    </>
  );

  function onSuccess() {
    incFormKey();
    notifySuccess('License added', 'License key is successfully added');
    router.stateService.go('portainer.licenses');
  }

  async function handleSubmitClick(values: FormValues) {
    addLicenseMutation.mutate(
      { values },
      {
        async onSuccess(data) {
          if (data.conflictingKeys) {
            const confirmed = await confirmDestructive({
              title: 'Are you sure?',
              message: (
                <div>
                  <div className="mb-4">
                    If you have licensing issues, please contact the Portainer
                    Success team at{' '}
                    <a href="mailto:success@portainer.io">
                      success@portainer.io
                    </a>
                  </div>
                  <div className="mb-2">
                    Adding this license will remove your current{' '}
                    {pluralize(data.conflictingKeys.length, 'license')}:
                  </div>
                  <ul className="ml-4 mb-2">
                    {data.conflictingKeys.map((key) => (
                      <li key={key}>{key}</li>
                    ))}
                  </ul>
                  <div>Are you sure you want to proceed?</div>
                </div>
              ),
              cancelButtonLabel: 'Cancel',
              confirmButton: buildConfirmButton('Confirm', 'danger'),
            });
            if (confirmed) {
              addLicenseMutation.mutate(
                { values, config: { params: { force: true } } },
                {
                  onSuccess,
                }
              );
            }
          } else {
            onSuccess();
          }
        },
      }
    );
  }
}

export function useAddLicenseMutation() {
  const queryClient = useQueryClient();

  return useMutation(
    ({ values, config }: AddFormType) => attachLicense(values, config),
    {
      ...withError('Failed to create License'),
      onSuccess() {
        return queryClient.invalidateQueries(['licenses']);
      },
    }
  );
}
