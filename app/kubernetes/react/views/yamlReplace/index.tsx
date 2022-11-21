import YAML from 'yaml';
import { useCurrentStateAndParams, useRouter } from '@uirouter/react';
import * as JsonPatch from 'fast-json-patch';

import { useAuthorizations } from '@/react/hooks/useUser';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';
import { notifySuccess, notifyError } from '@/portainer/services/notifications';

import { TooltipWithChildren } from '@@/Tip/TooltipWithChildren';
import { Button } from '@@/buttons';

import { ApplyPatch } from '../../proxy/service';

interface Props {
  yml: string;
  originalYml: string;
}

export function YAMLReplace({ yml, originalYml }: Props) {
  const environmentId = useEnvironmentId();
  const { params } = useCurrentStateAndParams();
  const router = useRouter();
  const canWrite = useAuthorizations(['K8sYAMLW']);

  function getOriginalResource(k: string, n: string, ns: string) {
    try {
      const docs = originalYml.split('---');
      const res = docs.find((yml) => {
        const y = YAML.parse(yml);
        const { name, namespace } = y.metadata;
        return k === y.kind && name === n && namespace === ns;
      });
      return res;
    } catch (e) {
      return null;
    }
  }

  async function replaceYAML() {
    const docs = yml.split('---');

    const res = await Promise.all(
      docs.map(async (yml) => {
        const newJson = YAML.parse(yml);
        const { name, namespace } = newJson.metadata;

        const oldResource = getOriginalResource(newJson.kind, name, namespace);
        if (!oldResource) {
          // skipping resource which were not in the original resource
          return true;
        }
        const originalJson = YAML.parse(oldResource);
        if (originalJson.metadata) {
          delete originalJson.metadata.managedFields;
          delete originalJson.metadata.resourceVersion;
        }

        const patch = JsonPatch.compare(originalJson, newJson);
        if (!patch.length) {
          // skipping resource if there are no changes
          return true;
        }

        const successFlag = await ApplyPatch(
          newJson.kind,
          newJson.apiVersion,
          environmentId,
          namespace || params.namespace,
          name,
          patch
        )
          .then(() => {
            notifySuccess(
              'Success',
              `${newJson.kind} ${name} updated successfully`
            );
            return true;
          })
          .catch(() => {
            notifyError(
              'Error',
              new Error(`Unable to update ${newJson.kind} ${name}`)
            );
            return false;
          });
        return successFlag;
      })
    );

    if (res.indexOf(false) === -1) {
      router.stateService.reload();
    }
  }

  if (!canWrite) {
    return null;
  }

  return (
    <TooltipWithChildren
      wrapperClassName="float-right"
      message="Applies any changes that you make in the YAML editor by calling the Kubernetes API to patch the relevant resources. Any resource removals or unexpected resource additions that you make in the YAML will be ignored. Note that editing is disabled for resources in namespaces marked as system."
    >
      <Button
        type="button"
        color="primary"
        size="small"
        onClick={replaceYAML}
        disabled={originalYml === yml}
      >
        Apply changes
      </Button>
    </TooltipWithChildren>
  );
}
