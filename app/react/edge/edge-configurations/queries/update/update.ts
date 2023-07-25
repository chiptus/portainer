import { useMutation, useQueryClient } from 'react-query';

import PortainerError from '@/portainer/error';
import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withError, withInvalidate } from '@/react-tools/react-query';

import {
  FormValues,
  FormValuesEdgeConfigurationMatchingRule,
  FormValuesEdgeConfigurationType,
  FormValuesFileMethod,
} from '../../common/types';
import { EdgeConfiguration } from '../../types';
import { queryKeys } from '../query-keys';
import { buildUrl } from '../urls';

import {
  EdgeConfigurationTypeString,
  EdgeConfigurationUpdatePayload,
} from './types';

export function translateType(
  type: FormValues['type'],
  rule: FormValues['matchingRule']
) {
  if (type === FormValuesEdgeConfigurationType.General) {
    return EdgeConfigurationTypeString.General;
  }

  if (
    type === FormValuesEdgeConfigurationType.DeviceSpecific &&
    rule === FormValuesEdgeConfigurationMatchingRule.MatchFile
  ) {
    return EdgeConfigurationTypeString.DeviceSpecificFile;
  }

  if (
    type === FormValuesEdgeConfigurationType.DeviceSpecific &&
    rule === FormValuesEdgeConfigurationMatchingRule.MatchFolder
  ) {
    return EdgeConfigurationTypeString.DeviceSpecificFolder;
  }

  throw new PortainerError('Unknown configuration type');
}

function createPayload(
  values: Partial<FormValues>
): EdgeConfigurationUpdatePayload {
  const edgeConfiguration: EdgeConfigurationUpdatePayload['edgeConfiguration'] =
    {};
  if (values.groupIds) {
    edgeConfiguration.edgeGroupIDs = values.groupIds;
  }

  if (values.type) {
    edgeConfiguration.type = translateType(values.type, values.matchingRule);
  }

  if (values.file && values.file.method === FormValuesFileMethod.Archive) {
    return {
      edgeConfiguration,
      file: values.file.content,
    };
  }

  if (values.file && values.file.method === FormValuesFileMethod.File) {
    return {
      edgeConfiguration,
      files: [],
    };
  }
  return { edgeConfiguration };
}

function toFormData(payload: EdgeConfigurationUpdatePayload): FormData {
  const fd = new FormData();
  fd.append('edgeConfiguration', JSON.stringify(payload.edgeConfiguration));
  if (payload.file) {
    fd.append('file', payload.file);
  }
  return fd;
}

interface Update {
  id: EdgeConfiguration['id'];
  values: Partial<FormValues>;
}

async function update({ id, values }: Update) {
  const payload = createPayload(values);
  const formPayload = toFormData(payload);

  try {
    const { data } = await axios.put<EdgeConfiguration>(
      buildUrl({ id }),
      formPayload
    );

    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Failed to update edge configuration');
  }
}

export function useUpdateMutation() {
  const queryClient = useQueryClient();
  return useMutation(update, {
    ...withInvalidate(queryClient, [queryKeys.base()]),
    ...withError(),
  });
}
