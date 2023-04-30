import { Form, Formik } from 'formik';
import * as yup from 'yup';
import { useCallback } from 'react';
import { FlaskConical } from 'lucide-react';

import { notifySuccess } from '@/portainer/services/notifications';
import { User } from '@/portainer/users/types';

import { LoadingButton } from '@@/buttons/LoadingButton';
import { TextTip } from '@@/Tip/TextTip';

import { OpenAIKeyField } from './OpenAIKeyField';
import { useUpdateUserOpenAIKeyMutation } from './useUpdateOpenAIKeyMutation';

interface FormValues {
  OpenAIApiKey: string;
}
const validation = yup.object({
  OpenAIApiKey: yup.string(),
});

interface Props {
  user: User;
}

export function OpenAIKeyForm({ user }: Props) {
  const initialValues: FormValues = {
    OpenAIApiKey: user.OpenAIApiKey || '',
  };

  const mutation = useUpdateUserOpenAIKeyMutation();

  const { mutate: updateSettings } = mutation;

  const handleSubmit = useCallback(
    (variables: FormValues) => {
      updateSettings(
        {
          ApiKey: variables.OpenAIApiKey,
        },
        {
          onSuccess() {
            notifySuccess(
              'Success',
              'Successfully updated OpenAI configuration'
            );
          },
        }
      );
    },
    [updateSettings]
  );

  return (
    <Formik<FormValues>
      initialValues={initialValues}
      onSubmit={handleSubmit}
      validationSchema={validation}
      validateOnMount
      enableReinitialize
    >
      {({ isValid, dirty }) => (
        <Form className="form-horizontal">
          <TextTip color="blue" icon={FlaskConical}>
            This is an experimental feature.
          </TextTip>

          <br />
          <br />

          <div className="form-group col-sm-12 text-muted small">
            This feature uses{' '}
            <a
              href="https://platform.openai.com/docs/models/gpt-3-5"
              target="_blank"
              rel="noreferrer"
            >
              the GPT-3.5 model
            </a>{' '}
            (<i>gpt-3.5-turbo</i>) from the OpenAI API.
          </div>

          <OpenAIKeyField />

          <div className="form-group">
            <div className="col-sm-12">
              <LoadingButton
                loadingText="Saving..."
                isLoading={mutation.isLoading}
                disabled={!isValid || !dirty}
                className="!ml-0"
                data-cy="account-openAIKeyButton"
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
