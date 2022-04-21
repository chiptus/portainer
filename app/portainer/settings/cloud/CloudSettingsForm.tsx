import { Field, Form, Formik } from 'formik';
import { useRouter } from '@uirouter/react';

import { react2angular } from '@/react-tools/react2angular';
import { FormControl } from '@/portainer/components/form-components/FormControl';
import { LoadingButton } from '@/portainer/components/Button/LoadingButton';
import { Input } from '@/portainer/components/form-components/Input';
import { CloudApiKeys } from '@/portainer/environments/components/kaas/kaas.types';

import { useUpdateSettings, useSettings } from '../queries';

import {
  CloudSettingsFormValues,
  CloudSettingsAPIPayload,
} from './cloud.types';
import { validationSchema } from './CloudSettingsForm.validation';

type Props = {
  reroute: boolean;
  showCivo?: boolean;
  showDigitalOcean?: boolean;
  showLinode?: boolean;
};

export function CloudSettingsForm({
  reroute,
  showCivo = false,
  showDigitalOcean = false,
  showLinode = false,
}: Props) {
  const router = useRouter();

  const settingsQuery = useSettings();
  const updateSettings = useUpdateSettings();

  if (settingsQuery.isLoading || updateSettings.isLoading) {
    return null;
  }

  if (settingsQuery.data?.CloudApiKeys) {
    return (
      <Formik
        initialValues={settingsQuery.data?.CloudApiKeys}
        validationSchema={() => validationSchema()}
        onSubmit={onSubmit}
        validateOnMount
        enableReinitialize
      >
        {({ errors, handleSubmit, isSubmitting, isValid }) => (
          <Form className="form-horizontal" onSubmit={handleSubmit} noValidate>
            {showCivo && (
              <FormControl
                inputId="civo_token"
                label="Civo"
                errors={errors?.CivoApiKey}
              >
                <Field
                  as={Input}
                  name="CivoApiKey"
                  autoComplete="off"
                  id="civo_token"
                  placeholder="e.g. DgJ33kwIhnHumQFyc8ihGwWJql9cC8UJDiBhN8YImKqiX"
                  data-cy="cloudSettings-civoTokenInput"
                />
              </FormControl>
            )}
            {showLinode && (
              <FormControl
                inputId="linode_token"
                label="Linode"
                errors={errors?.LinodeToken}
              >
                <Field
                  as={Input}
                  name="LinodeToken"
                  autoComplete="off"
                  id="linode_token"
                  placeholder="e.g. 92gsh9r9u5helgs4eibcuvlo403vm45hrmc6mzbslotnrqmkwc1ovqgmolcyq0wc"
                  data-cy="cloudSettings-linodeTokenInput"
                />
              </FormControl>
            )}
            {showDigitalOcean && (
              <FormControl
                inputId="do_token"
                label="DigitalOcean"
                errors={errors?.DigitalOceanToken}
              >
                <Field
                  as={Input}
                  name="DigitalOceanToken"
                  autoComplete="off"
                  id="do_token"
                  placeholder="e.g. dop_v1_n9rq7dkcbg3zb3bewtk9nnvmfkyfnr94d42n28lym22vhqu23rtkllsldygqm22v"
                  data-cy="cloudSettings-doTokenInput"
                />
              </FormControl>
            )}

            <div className="form-group">
              <div className="col-sm-12">
                <LoadingButton
                  disabled={!isValid}
                  dataCy="cloudSettings-saveSettingsButton"
                  isLoading={isSubmitting}
                  loadingText="Saving credentials..."
                >
                  Save
                </LoadingButton>
              </div>
            </div>
          </Form>
        )}
      </Formik>
    );
  }

  // return nothing for all other states
  return null;

  async function onSubmit(formvalues: CloudSettingsFormValues) {
    if (settingsQuery.data?.CloudApiKeys) {
      const payload = makeApiKeyPayload(
        formvalues,
        settingsQuery.data?.CloudApiKeys
      );
      if (Object.keys(payload.CloudApiKeys).length > 0) {
        updateSettings.mutate(payload, {
          onSuccess: () => {
            if (reroute) {
              router.stateService.reload();
            }
          },
        });
      }
    }
  }
}

// filter the form values to only include the values that have changed
function makeApiKeyPayload(
  formvalues: CloudSettingsFormValues,
  initialValues: Partial<CloudApiKeys>
) {
  const filteredForm: CloudSettingsAPIPayload = {
    CloudApiKeys: {},
  };

  Object.entries(formvalues).forEach(([provider, newApiKey]) => {
    if (
      provider === 'CivoApiKey' ||
      provider === 'DigitalOceanToken' ||
      provider === 'LinodeToken'
    ) {
      const currentProvider = initialValues[provider] || '';
      if (
        apiKeyChanged(newApiKey, currentProvider) &&
        filteredForm.CloudApiKeys
      ) {
        filteredForm.CloudApiKeys[provider] = newApiKey.trim();
      }
    }
  });
  return filteredForm;
}

function apiKeyChanged(newApiKey: string, currentApiKey: string) {
  // if all characters are '*'
  if (/^[*]+$/.test(newApiKey)) {
    return false;
  }
  return newApiKey !== currentApiKey;
}

export const CloudSettingsFormAngular = react2angular(CloudSettingsForm, [
  'reroute',
]);
