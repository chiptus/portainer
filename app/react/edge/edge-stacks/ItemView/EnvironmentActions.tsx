import { Environment } from '@/portainer/environments/types';

import { Button } from '@@/buttons';
import { Link } from '@@/Link';

import { EdgeStack } from '../types';

import { LogsActions } from './LogsActions';

interface Props {
  environment: Environment;
  edgeStackId: EdgeStack['Id'];
}

export function EnvironmentActions({ environment, edgeStackId }: Props) {
  return (
    <div className="space-x-2">
      {environment.Snapshots.length > 0 && (
        <Link
          to="edge.devices.containers"
          params={{ environmentId: environment.Id, edgeStackId }}
          className="!text-inherit hover:!no-underline"
        >
          <Button color="none" title="Browse Snapshot">
            <i className="fa fa-search" aria-hidden="true" />
          </Button>
        </Link>
      )}
      {environment.Edge.AsyncMode && (
        <LogsActions environmentId={environment.Id} edgeStackId={edgeStackId} />
      )}
    </div>
  );
}
