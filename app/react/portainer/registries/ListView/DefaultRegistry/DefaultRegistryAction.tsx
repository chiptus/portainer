import {
  useSettings,
  useUpdateDefaultRegistrySettingsMutation,
} from 'Portainer/settings/queries';
import { notifySuccess } from 'Portainer/services/notifications';

import { Tooltip } from '@@/Tip/Tooltip';
import { Button } from '@@/buttons';

export function DefaultRegistryAction() {
  const settingsQuery = useSettings(
    (settings) => settings.DefaultRegistry.Hide
  );
  const defaultRegistryMutation = useUpdateDefaultRegistrySettingsMutation();

  if (!settingsQuery.isSuccess) {
    return null;
  }
  const hideDefaultRegistry = settingsQuery.data;

  return (
    <>
      {!hideDefaultRegistry ? (
        <div>
          <Button
            className="btn btn-xs btn-danger"
            onClick={() => handleShowOrHide(true)}
          >
            <i className="fa fa-eye-slash space-right" aria-hidden="true" />{' '}
            Hide for all users
          </Button>

          <Tooltip
            message="This hides the option in any registry dropdown prompts but does not prevent a user
                                  from deploying anonymously from Docker Hub directly via YAML.<br />
                                  Note: Docker Hub (anonymous) will always continue to show if there are NO other registries."
          />
        </div>
      ) : (
        <div>
          <Button
            className="btn btn-xs btn-success"
            onClick={() => handleShowOrHide(false)}
          >
            <i className="fa fa-eye space-right" aria-hidden="true" /> Show for
            all users
          </Button>
          <Tooltip
            message="This reveals the option in any registry dropdown prompts.<br />
                                  Note: Docker Hub (anonymous) will always continue to show if there are NO other registries."
          />
        </div>
      )}
    </>
  );

  function handleShowOrHide(hideDefaultRegistry: boolean) {
    defaultRegistryMutation.mutate(
      {
        Hide: hideDefaultRegistry,
      },
      {
        onSuccess() {
          notifySuccess('Default registry Settings updated successfully');
        },
      }
    );
  }
}
