import { FormikErrors, FormikHandlers } from 'formik';

import { EdgeGroup } from '@/react/edge/edge-groups/types';
import { useEdgeGroups } from '@/react/edge/edge-groups/queries/useEdgeGroups';

import { FormControl } from '@@/form-components/FormControl';
import { Select } from '@@/form-components/ReactSelect';
import { TextTip } from '@@/Tip/TextTip';

import { EnvironmentType } from '../../types';

import { FormValues } from './types';

interface Props {
  disabled?: boolean;
  onBlur: FormikHandlers['handleBlur'];
  value: FormValues['groupIds'];
  error?: FormikErrors<FormValues>['groupIds'];
  onChange(value: FormValues['groupIds']): void;
}

export function EdgeGroupsField({
  disabled,
  onBlur,
  value,
  error,
  onChange,
}: Props) {
  const groupsQuery = useEdgeGroups();

  const selectedGroups = groupsQuery.data?.filter((group) =>
    value.includes(group.Id)
  );

  let errorMessage = error;
  if (selectedGroups && selectedGroups.length >= 1) {
    errorMessage = validateMultipleGroups(selectedGroups);
  }

  const upgradeVersionInstruction =
    getUpgradeVersionInstruction(selectedGroups);

  return (
    <>
      <div>
        <FormControl
          label="Edge Groups"
          required
          inputId="groups-select"
          errors={errorMessage}
          tooltip="Updates are done based on groups, allowing you to choose multiple devices at the same time and the ability to roll out progressively across all environments by scheduling them for different days."
        >
          <Select
            name="groupIds"
            onBlur={onBlur}
            value={selectedGroups}
            inputId="groups-select"
            placeholder="Select one or multiple group(s)"
            onChange={(selectedGroups) =>
              onChange(selectedGroups.map((g) => g.Id))
            }
            isMulti
            options={groupsQuery.data || []}
            getOptionLabel={(group) => group.Name}
            getOptionValue={(group) => group.Id.toString()}
            closeMenuOnSelect={false}
            isDisabled={disabled}
          />
        </FormControl>
        <TextTip color="blue">
          Select edge groups of edge environments to update
        </TextTip>
      </div>
      {upgradeVersionInstruction}
    </>
  );
}

function validateMultipleGroups(groups: EdgeGroup[]) {
  let uniqueEndpointType = EnvironmentType.Docker;

  for (let i = 0; i < groups.length; i++) {
    for (let j = 0; j < groups[i].EndpointTypes.length; j++) {
      if (i === 0 && j === 0) {
        uniqueEndpointType = groups[i].EndpointTypes[j];
      }

      if (uniqueEndpointType !== groups[i].EndpointTypes[j]) {
        return 'Please select edge group(s) that have edge environments of the same type.';
      }
    }
  }

  return '';
}

function getUpgradeVersionInstruction(groups: EdgeGroup[] | undefined) {
  let upgradeVersionInstruction = (
    <TextTip color="blue" className="mt-1">
      You can upgrade from any agent version to 2.17 or later only. You can not
      upgrade to an agent version prior to 2.17 . The ability to rollback to
      originating version is for 2.15.0+ only.
    </TextTip>
  );

  if (groups && groups.length >= 1) {
    let uniqueEndpointType = EnvironmentType.Docker;
    for (let i = 0; i < groups.length; i++) {
      for (let j = 0; j < groups[i].EndpointTypes.length; j++) {
        if (i === 0 && j === 0) {
          uniqueEndpointType = groups[i].EndpointTypes[j];
        }

        if (uniqueEndpointType !== groups[i].EndpointTypes[j]) {
          return upgradeVersionInstruction;
        }
      }
    }

    switch (uniqueEndpointType) {
      case EnvironmentType.EdgeAgentOnNomad:
        upgradeVersionInstruction = (
          <TextTip color="blue" className="mt-1">
            You can upgrade nomad agent from 2.18 to later only. You can not
            upgrade to nomad agent version prior to 2.18.
          </TextTip>
        );
        break;
      default:
    }
  }

  return upgradeVersionInstruction;
}
