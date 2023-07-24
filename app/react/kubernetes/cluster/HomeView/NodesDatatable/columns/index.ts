import { name } from './name';
import { role } from './role';
import { status } from './status';
import { cpu } from './cpu';
import { memory } from './memory';
import { version } from './version';
import { ip } from './ip';
import { getActions } from './actions';

export function getColumns(
  isServerMetricsEnabled: boolean,
  canSSH: boolean,
  canCheckStatus: boolean
) {
  if (!isServerMetricsEnabled && !(canSSH || canCheckStatus)) {
    return [name, role, status, cpu, memory, version, ip];
  }

  return [
    name,
    role,
    status,
    cpu,
    memory,
    version,
    ip,
    getActions(isServerMetricsEnabled, canSSH, canCheckStatus),
  ];
}
