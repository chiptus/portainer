import { RelativePathModel } from '@/react/portainer/gitops/types';

import { SwitchField } from '@@/form-components/SwitchField';
import { TextTip } from '@@/Tip/TextTip';
import { FormControl } from '@@/form-components/FormControl';
import { Input, Select } from '@@/form-components/Input';

interface Props {
  value: RelativePathModel;
  readonly?: boolean;
}

export function RelativePathFieldset({ value, readonly }: Props) {
  return (
    <>
      <div className="form-group">
        <div className="col-sm-12">
          <SwitchField
            name="EnableRelativePaths"
            label="Enable relative path volumes"
            labelClass="col-sm-3 col-lg-2"
            tooltip="Enabling this means you can specify relative path volumes in your Compose files, with Portainer pulling the content from your git repository to the environment the stack is deployed to."
            disabled={readonly}
            checked={value.SupportRelativePath}
            onChange={(value) => value}
          />
        </div>
      </div>

      {value.SupportRelativePath && (
        <>
          <div className="form-group">
            <div className="col-sm-12">
              <TextTip color="blue">
                For relative path volumes use with Docker Swarm, you must have a
                network filesystem which all of your nodes can access.
              </TextTip>
            </div>
          </div>

          <div className="form-group">
            <div className="col-sm-12">
              <FormControl label="Local filesystem path">
                <Input
                  name="FilesystemPath"
                  placeholder="/mnt"
                  disabled={readonly}
                  value={value.FilesystemPath}
                  onChange={(value) => value}
                />
              </FormControl>
            </div>
          </div>
        </>
      )}

      <div className="form-group">
        <div className="col-sm-12">
          <TextTip color="blue">
            When enabled, corresponding Edge ID will be passed through as an
            environment variable: PORTAINER_EDGE_ID.
          </TextTip>
        </div>
      </div>

      <div className="form-group">
        <div className="col-sm-12">
          <SwitchField
            name="EnablePerDeviceConfigs"
            label="GitOps Edge configurations"
            labelClass="col-sm-3 col-lg-2"
            tooltip="By enabling the GitOps Edge Configurations feature, you gain the ability to define relative path volumes in your configuration files. Portainer will then automatically fetch the content from your git repository by matching the folder name or file name with the Portainer Edge ID, and apply it to the environment where the stack is deployed"
            disabled={readonly}
            checked={!!value.SupportPerDeviceConfigs}
            onChange={(value) => value}
          />
        </div>
      </div>

      {value.SupportPerDeviceConfigs && (
        <>
          <div className="form-group">
            <div className="col-sm-12">
              <TextTip color="blue">
                Specify the directory name where your configuration will be
                located. This will allow you to manage device configuration
                settings with a Git repo as your template.
              </TextTip>
            </div>
          </div>

          <div className="form-group">
            <div className="col-sm-12">
              <FormControl label="Directory">
                <Input
                  placeholder="config"
                  disabled={readonly}
                  value={value.PerDeviceConfigsPath}
                />
              </FormControl>
            </div>
          </div>

          <div className="form-group">
            <div className="col-sm-12">
              <TextTip color="blue">
                Select which rule to use when matching configuration with
                Portainer Edge ID either on a per-device basis or group-wide
                with an Edge Group. Only configurations that match the selected
                rule will be accessible through their corresponding paths.
                Deployments that rely on accessing the configuration may
                experience errors.
              </TextTip>
            </div>
          </div>

          <div className="form-group">
            <div className="col-sm-12">
              <FormControl label="Device matching rule">
                <Select
                  value={value.PerDeviceConfigsMatchType}
                  options={[
                    {
                      label: '',
                      value: '',
                    },
                    {
                      label: 'Match file name with Portainer Edge ID',
                      value: 'file',
                    },
                    {
                      label: 'Match folder name with Portainer Edge ID',
                      value: 'dir',
                    },
                  ]}
                  disabled={readonly}
                />
              </FormControl>
            </div>
          </div>

          <div className="form-group">
            <div className="col-sm-12">
              <FormControl label="Group matching rule">
                <Select
                  value={value.PerDeviceConfigsGroupMatchType}
                  options={[
                    {
                      label: '',
                      value: '',
                    },
                    {
                      label: 'Match file name with Edge Group',
                      value: 'file',
                    },
                    {
                      label: 'Match folder name with Edge Group',
                      value: 'dir',
                    },
                  ]}
                  disabled={readonly}
                />
              </FormControl>
            </div>
          </div>

          <div className="form-group">
            <div className="col-sm-12">
              <TextTip color="blue">
                You can use it as an environment variable with an image:
                myapp:$&#123;PORTAINER_EDGE_ID&#125; or
                myapp:$&#123;PORTAINER_EDGE_GROUP&#125;. You can also use it
                with the relative path for volumes -
                ./config/$&#123;PORTAINER_EDGE_ID&#125;:/myapp/config or
                ./config/$&#123;EDGE_GROUP&#125;:/myapp/groupconfig. More
                documentation can be found{' '}
                <a href="https://docs.portainer.io/user/edge/stacks/add#gitops-edge-configurations">
                  here
                </a>
                .
              </TextTip>
            </div>
          </div>
        </>
      )}
    </>
  );
}
