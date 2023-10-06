import _ from 'lodash';

import { Environment } from '@/react/portainer/environments/types';
import { semverCompare } from '@/react/common/semver-utils';

import { TextTip } from '@@/Tip/TextTip';

import { VersionSelect } from './VersionSelect';
import { ScheduledTimeField } from './ScheduledTimeField';

interface Props {
  environments: Environment[];
  hasTimeZone: boolean;
  hasNoTimeZone: boolean;
  hasGroupSelected: boolean;
  version: string;
}

export function UpdateScheduleDetailsFieldset({
  environments,
  hasTimeZone,
  hasNoTimeZone,
  hasGroupSelected,
  version,
}: Props) {
  const minVersion = _.first(
    _.compact<string>(environments.map((env) => env.Agent.Version)).sort(
      (a, b) => semverCompare(a, b)
    )
  );

  return (
    <>
      <EnvironmentInfo environments={environments} version={version} />

      <VersionSelect minVersion={minVersion} />

      {hasTimeZone && hasGroupSelected && <ScheduledTimeField />}
      {hasNoTimeZone && (
        <TextTip>
          These edge groups have older versions of the edge agent that do not
          support scheduling, these will happen immediately
        </TextTip>
      )}
    </>
  );
}

function EnvironmentInfo({
  environments,
  version,
}: {
  environments: Environment[];
  version: string;
}) {
  if (!environments.length) {
    return (
      <TextTip color="orange">
        No environments options for the selected edge groups
      </TextTip>
    );
  }

  if (!version) {
    return null;
  }

  const environmentsAlreadyOnVersion = environments.filter(
    (env) =>
      !env.Agent.Version || semverCompare(version, env.Agent.Version) <= 0
  ).length;

  const environmentsToUpdate =
    environments.length - environmentsAlreadyOnVersion;

  if (environmentsToUpdate === 0) {
    return (
      <TextTip color="orange">
        All edge agents are already running version {version}
      </TextTip>
    );
  }

  return (
    <TextTip color="blue">
      {environmentsAlreadyOnVersion > 0 && (
        <>
          {environmentsAlreadyOnVersion} edge agent
          {environmentsAlreadyOnVersion > 1 ? 's are' : ' is'} currently running
          version greater than or equal to {version}, and{' '}
        </>
      )}
      {environmentsToUpdate > 0 && (
        <>
          {environments.length - environmentsAlreadyOnVersion} will be updated
          to version {version}
        </>
      )}
    </TextTip>
  );
}
