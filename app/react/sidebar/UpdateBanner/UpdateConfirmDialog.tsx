import { Form, Formik } from 'formik';

import { useUpdatePortainerMutation } from '@/react/portainer/system/useUpdatePortainerMutation';
import { notifySuccess } from '@/portainer/services/notifications';

import { Button, LoadingButton } from '@@/buttons';
import { Modal } from '@@/modals/Modal';

const initialValues = {};

export function UpdateConfirmDialog({
  onDismiss,
  goToLoading,
}: {
  onDismiss: () => void;
  goToLoading: () => void;
}) {
  const updateMutation = useUpdatePortainerMutation();

  return (
    <Modal
      onDismiss={onDismiss}
      aria-label="Update Portainer to the latest version"
    >
      <Modal.Header
        title={<h4 className="text-xl font-medium">Update Portainer</h4>}
      />
      <Formik
        initialValues={initialValues}
        onSubmit={handleSubmit}
        validateOnMount
      >
        {() => (
          <Form noValidate>
            <Modal.Body>
              <p className="font-semibold text-gray-7">
                Are you sure you want to update Portainer?
              </p>
            </Modal.Body>
            <Modal.Footer>
              <div className="flex w-full gap-2 [&>*]:w-1/2">
                <Button
                  color="default"
                  size="medium"
                  className="w-full"
                  onClick={onDismiss}
                >
                  Dismiss
                </Button>
                <LoadingButton
                  color="primary"
                  size="medium"
                  loadingText="Starting update"
                  isLoading={updateMutation.isLoading}
                >
                  Start update
                </LoadingButton>
              </div>
            </Modal.Footer>
          </Form>
        )}
      </Formik>
    </Modal>
  );

  function handleSubmit() {
    updateMutation.mutate(undefined, {
      onSuccess() {
        notifySuccess('Success', 'Starting update');
        goToLoading();
      },
    });
  }
}
