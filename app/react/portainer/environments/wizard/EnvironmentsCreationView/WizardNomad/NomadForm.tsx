import { Field, Form, useFormikContext } from 'formik';

import { Button } from '@/portainer/components/Button';
import { Code } from '@/portainer/components/Code';
import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Input, Select } from '@/portainer/components/form-components/Input';
import { SwitchField } from '@/portainer/components/form-components/SwitchField';
import { NavTabs } from '@/portainer/components/NavTabs/NavTabs';
import { TextTip } from '@/portainer/components/Tip/TextTip';
import { buildLinuxNomadCommand } from '@/edge/components/EdgeScriptForm/scripts';
import { useSettings } from '@/portainer/settings/queries';

import { MetadataFieldset } from '../shared/MetadataFieldset';

import { EdgeInfo, FormValues } from './types';

interface FormProps {
  isSubmitting: boolean;
  edgeInfo?: EdgeInfo;
}

export function NomadForm({ isSubmitting, edgeInfo }: FormProps) {
  const { errors, isValid, handleSubmit, values, touched, setFieldValue } =
    useFormikContext<FormValues>();

  const settingsQuery = useSettings((settings) => settings.AgentSecret);

  if (!settingsQuery.isSuccess) {
    return null;
  }

  const agentVersion = settingsQuery.data;

  return (
    <Form onSubmit={handleSubmit} noValidate>
      <div className="row">
        <div className="col-sm-12">
          <FormControl
            label="Name"
            inputId="name-input"
            errors={touched.name && errors.name}
            required
          >
            <Field
              name="name"
              as={Input}
              id="name-input"
              required
              placeholder="e.g. nomad-cluster01"
            />
          </FormControl>

          <div className="my-8">
            <SwitchField
              checked={values.authEnabled}
              onChange={(value) => {
                if (!value) {
                  setFieldValue('token', '');
                }
                setFieldValue('authEnabled', value);
              }}
              label="Nomad Authentication Enabled"
              tooltip="Nomad authentication is only required if you have ACL enabled"
            />
          </div>

          {values.authEnabled && (
            <FormControl
              label="Nomad Token"
              inputId="token-input"
              errors={touched.token && errors.token}
            >
              <Field name="token" as={Input} id="token-input" />
            </FormControl>
          )}

          <FormControl
            label="Portainer server URL"
            inputId="portainer-url-input"
            errors={touched.portainerUrl && errors.portainerUrl}
            tooltip="URL of the Portainer instance that the agent will use to initiate the communications"
            required
          >
            <Field
              name="portainerUrl"
              as={Input}
              id="portainer-url-input"
              required
            />
          </FormControl>

          <FormControl
            label="Poll Frequency"
            inputId="poll-frequency-input"
            errors={touched.pollFrequency && errors.pollFrequency}
            tooltip="Interval used by this Edge agent to check in with the Portainer instance. Affects Edge environment management and Edge compute features."
          >
            <Field
              name="pollFrequency"
              as={Select}
              id="poll-frequency-input"
              options={[
                { label: 'Use default interval (5 seconds)', value: 0 },
                {
                  label: '5 seconds',
                  value: 5,
                },
                {
                  label: '10 seconds',
                  value: 10,
                },
                {
                  label: '30 seconds',
                  value: 30,
                },
                { label: '5 minutes', value: 300 },
                { label: '1 hour', value: 3600 },
                { label: '1 day', value: 86400 },
              ]}
            />
          </FormControl>

          <FormControl
            label="Environment variables"
            tooltip="Comma separated list of environment variables that will be sourced from the host where the agent is deployed."
            inputId="env-variables-input"
            errors={touched.envVars && errors.envVars}
          >
            <Field
              name="envVars"
              as={Input}
              placeholder="foo=bar,myvar"
              id="env-variables-input"
            />
          </FormControl>

          <div className="my-8">
            <SwitchField
              checked={values.allowSelfSignedCertificates}
              onChange={(value) =>
                setFieldValue('allowSelfSignedCertificates', value)
              }
              label="Allow self-signed certs"
              tooltip="When allowing self-signed certificates the edge agent will ignore the domain validation when connecting to Portainer via HTTPS"
            />
          </div>

          <MetadataFieldset />
        </div>
      </div>
      <div className="row">
        <div className="col-sm-12">
          <TextTip color="blue">
            As this is our first version of Nomad on Portainer, we do not
            support HTTPS yet
          </TextTip>
        </div>
      </div>

      <div className="row">
        <div className="col-sm-12">
          <NavTabs
            options={[
              {
                id: 'linux',
                label: 'Linux',
                children: (
                  <LinuxTab
                    agentVersion={agentVersion}
                    edgeInfo={edgeInfo}
                    allowSelfSignedCertificates={
                      values.allowSelfSignedCertificates
                    }
                    token={values.token}
                    envVars={values.envVars}
                  />
                ),
              },
            ]}
            selectedId="linux"
          />

          <TextTip color="blue">
            This script, and the join token for pre staging edge environments
            can also be found in the Environment Details page
          </TextTip>
        </div>
      </div>

      <div className="row">
        <div className="col-sm-12">
          {!edgeInfo ? (
            <Button
              color="primary"
              type="submit"
              disabled={!isValid || isSubmitting}
            >
              <i className="fa fa-plug space-right" />
              Create
            </Button>
          ) : (
            <div className="flex justify-end">
              <Button color="primary" type="reset">
                Add another environment
              </Button>
            </div>
          )}
        </div>
      </div>
    </Form>
  );
}

interface LinuxTabProps {
  token: string;
  allowSelfSignedCertificates: boolean;
  edgeInfo?: EdgeInfo;
  agentVersion: string;
  envVars: string;
}

function LinuxTab({
  token,
  allowSelfSignedCertificates,
  edgeInfo,
  agentVersion,
  envVars,
}: LinuxTabProps) {
  if (!edgeInfo || !edgeInfo.key) {
    return (
      <Code>
        {`
      
      `}
      </Code>
    );
  }

  const code = buildLinuxNomadCommand(
    agentVersion,
    edgeInfo.key,
    {
      nomadToken: token,
      allowSelfSignedCertificates,
      envVars,
    },
    edgeInfo.id,
    ''
  );

  return (
    <>
      <b>
        <small>
          Copy and run this script in CLI on edge host to connect this Edge
          Agent to your environment:
        </small>
      </b>

      <Code showCopyButton>{code}</Code>
    </>
  );
}
