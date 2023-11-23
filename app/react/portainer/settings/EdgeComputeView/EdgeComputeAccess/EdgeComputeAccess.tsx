import { Formik, Form } from 'formik';
import { useMemo } from 'react';
import { orderBy } from 'lodash';
import { Plus, Users2 } from 'lucide-react';
import { useRouter } from '@uirouter/react';

import { notifySuccess } from '@/portainer/services/notifications';
import {
  Option,
  PorAccessManagementUsersSelector,
} from '@/react/portainer/access-control/AccessManagement/PorAccessManagementUsersSelector';
import { Role, User } from '@/portainer/users/types';
import { useUsers } from '@/portainer/users/queries';
import { useUpdateUserMutation } from '@/portainer/users/queries/useUpdateUserMutation';

import { Widget, WidgetBody, WidgetTitle } from '@@/Widget';
import { LoadingButton } from '@@/buttons/LoadingButton';
import { TextTip } from '@@/Tip/TextTip';

import { validationSchema } from './EdgeComputeAccess.validation';
import { FormValues } from './types';
import { Datatable } from './Datatable';

// EE-6176: the specs wanted to have both teams and users be assignable as Edge Admins
// However because of the technical path we've chosen to implement Edge Admins
// It is currently not suitable for teams.
// This comment is a reminder for RBAC refactor that this feature should also support teams

const initialValues: FormValues = {
  selectedUsersAndTeams: [],
};

export function EdgeComputeAccess() {
  const router = useRouter();
  const usersQuery = useUsers();
  const updateUserMutation = useUpdateUserMutation();

  const options = useMemo(() => {
    if (!usersQuery.data) {
      return [];
    }
    return buildOptions(usersQuery.data);
  }, [usersQuery.data]);

  if (usersQuery.isLoading) {
    return null;
  }

  return (
    <div className="row">
      <Widget>
        <WidgetTitle icon={Users2} title="Edge Compute access" />

        <WidgetBody>
          <Formik
            initialValues={initialValues}
            enableReinitialize
            validationSchema={() => validationSchema()}
            onSubmit={onSubmit}
            validateOnMount
          >
            {({
              values,
              handleSubmit,
              setFieldValue,
              isSubmitting,
              isValid,
              dirty,
            }) => (
              <Form
                className="form-horizontal"
                onSubmit={handleSubmit}
                noValidate
              >
                <TextTip
                  className="mb-2"
                  childrenWrapperClassName="text-warning"
                >
                  Adding user access will require the affected user(s) to logout
                  and login for the changes to be taken into account.
                </TextTip>

                <PorAccessManagementUsersSelector
                  options={options}
                  onChange={(opts) =>
                    setFieldValue('selectedUsersAndTeams', opts)
                  }
                  value={values.selectedUsersAndTeams}
                />

                <TextTip color="blue" className="mb-2">
                  When you designate certain users as &apos;Edge
                  administrators&apos;, you grant them comprehensive control
                  over all resources within every environment, including edge
                  resources, by providing access to the edge compute feature.
                </TextTip>

                <div className="form-group mt-5">
                  <div className="col-sm-12">
                    <LoadingButton
                      disabled={!isValid || !dirty}
                      data-cy="settings-edgeComputAccessSubmit"
                      isLoading={isSubmitting}
                      loadingText="Saving access..."
                      icon={Plus}
                    >
                      Create access
                    </LoadingButton>
                  </div>
                </div>
              </Form>
            )}
          </Formik>
          <Datatable />
        </WidgetBody>
      </Widget>
    </div>
  );

  function onSubmit(values: FormValues) {
    values.selectedUsersAndTeams.forEach((ut) => {
      if (ut.Type === 'user') {
        updateUserMutation.mutate(
          {
            userId: ut.Id,
            payload: { role: Role.EdgeAdmin },
          },
          {
            onSuccess: () => {
              notifySuccess('Success', 'User successfully updated');
            },
          }
        );
      }
    });
    router.stateService.reload();
  }
}

// EE-6176: see top-file comment on why there are 2 option types (team | user)
function buildOptions(users: User[]): Option[] {
  const options: Option[] = [];
  options.push(
    ...users
      .filter((u) => u.Role === Role.Standard)
      .map(
        ({ Id, Username: Name }): Option => ({
          Type: 'user',
          Id,
          Name,
        })
      )
  );
  return orderBy(options, 'Name', 'asc');
}
