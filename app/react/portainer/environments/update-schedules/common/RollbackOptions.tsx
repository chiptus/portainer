import { useFormikContext } from 'formik';
import _ from 'lodash';
import { useMemo, useEffect } from 'react';
import { useCurrentStateAndParams } from '@uirouter/react';

import { useEdgeGroups } from '@/react/edge/edge-groups/queries/useEdgeGroups';

import { TextTip } from '@@/Tip/TextTip';

import { usePreviousVersions } from '../queries/usePreviousVersions';

import { FormValues } from './types';
import { useEdgeGroupsEnvironmentIds } from './useEdgeGroupsEnvironmentIds';

export function RollbackOptions() {
  const { isLoading, count, version, versionError } = useSelectVersionOnMount();

  const groupNames = useGroupNames();

  if (versionError) {
    return <TextTip>{versionError}</TextTip>;
  }

  if (!count) {
    return (
      <TextTip>
        There are no rollback options available for your selected groups(s)
      </TextTip>
    );
  }

  if (isLoading || !groupNames) {
    return null;
  }

  return (
    <div className="form-group">
      <div className="col-sm-12">
        {count} edge device(s) from {groupNames} will rollback to version{' '}
        {version}
      </div>
    </div>
  );
}

function useSelectVersionOnMount() {
  const {
    values: { groupIds, version },
    setFieldValue,
    setFieldError,
    errors: { version: versionError },
  } = useFormikContext<FormValues>();

  const environmentIdsQuery = useEdgeGroupsEnvironmentIds(groupIds);

  const {
    params: { id: idParam },
  } = useCurrentStateAndParams();

  const id = parseInt(idParam, 10);

  const previousVersionsQuery = usePreviousVersions({
    enabled: !!environmentIdsQuery.data,
    skipScheduleID: id,
  });

  const { previousVersions, count } = useMemo(() => {
    const previousVersions = previousVersionsQuery.data;
    const environmentIds = environmentIdsQuery.data;

    if (!previousVersions || !environmentIds) {
      return { previousVersions: [], count: 0 };
    }

    const filteredVersions = Object.fromEntries(
      environmentIds
        .map((envId) => [envId, previousVersions[envId]])
        .filter(([, version]) => !!version)
    );

    return {
      previousVersions: _.uniq(
        _.compact(Object.values(filteredVersions).flat())
      ),
      count: Object.keys(filteredVersions).length,
    };
  }, [environmentIdsQuery.data, previousVersionsQuery.data]);

  useEffect(() => {
    switch (previousVersions.length) {
      case 0:
        setFieldValue('version', '');
        setFieldError('version', 'No rollback options available');
        break;
      case 1:
        setFieldValue('version', previousVersions[0]);
        break;
      default:
        setFieldError(
          'version',
          'Rollback is not available for these edge group as there are multiple version types to rollback to'
        );
    }
  }, [previousVersions, setFieldError, setFieldValue]);

  return {
    isLoading: previousVersionsQuery.isLoading,
    versionError,
    version,
    count,
  };
}

function useGroupNames() {
  const {
    values: { groupIds },
  } = useFormikContext<FormValues>();

  const groupsQuery = useEdgeGroups({
    select: (groups) => Object.fromEntries(groups.map((g) => [g.Id, g.Name])),
  });

  if (!groupsQuery.data) {
    return null;
  }

  return groupIds.map((id) => groupsQuery.data[id]).join(', ');
}
