import _ from 'lodash';
import {
  AlertTriangle,
  CheckCircle,
  type Icon as IconType,
  Loader2,
  XCircle,
  MinusCircle,
} from 'lucide-react';

import { Icon, IconMode } from '@@/Icon';
import { Tooltip } from '@@/Tip/Tooltip';

import { DeploymentStatus, EdgeStack, StatusType } from '../../types';

export function EdgeStackStatus({ edgeStack }: { edgeStack: EdgeStack }) {
  const status = Object.values(edgeStack.Status);
  const lastStatus = _.compact(status.map((s) => _.last(s.Status)));

  const { icon, label, mode, spin, tooltip } = getStatus(
    edgeStack.NumDeployments,
    lastStatus
  );

  return (
    <div className="mx-auto inline-flex items-center gap-2">
      {icon && <Icon icon={icon} spin={spin} mode={mode} />}
      {label}
      {tooltip && <Tooltip message={tooltip} />}
    </div>
  );
}

function getStatus(
  numDeployments: number,
  envStatus: Array<DeploymentStatus>
): {
  label: string;
  icon?: IconType;
  spin?: boolean;
  mode?: IconMode;
  tooltip?: string;
} {
  if (!numDeployments) {
    return {
      label: 'Unavailable',
      icon: MinusCircle,
      spin: false,
      mode: 'secondary',
      tooltip:
        "Your edge stack's status is currently unavailable due to the absence of an available environment in your edge group",
    };
  }

  if (envStatus.length < numDeployments) {
    return {
      label: 'Deploying',
      icon: Loader2,
      spin: true,
      mode: 'primary',
    };
  }

  const allFailed = envStatus.every((s) => s.Type === StatusType.Error);

  if (allFailed) {
    return {
      label: 'Failed',
      icon: XCircle,
      mode: 'danger',
    };
  }

  const allRunning = envStatus.every((s) => s.Type === StatusType.Running);

  if (allRunning) {
    return {
      label: 'Running',
      icon: CheckCircle,
      mode: 'success',
    };
  }

  const hasDeploying = envStatus.some((s) => s.Type === StatusType.Deploying);
  const hasRunning = envStatus.some((s) => s.Type === StatusType.Running);
  const hasFailed = envStatus.some((s) => s.Type === StatusType.Error);

  if (hasRunning && hasFailed && !hasDeploying) {
    return {
      label: 'Partially Running',
      icon: AlertTriangle,
      mode: 'warning',
    };
  }

  return {
    label: 'Deploying',
    icon: Loader2,
    spin: true,
    mode: 'primary',
  };
}
