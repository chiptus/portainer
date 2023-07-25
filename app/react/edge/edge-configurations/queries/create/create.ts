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
  EdgeConfigurationCreatePayload,
  EdgeConfigurationTypeString,
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

function createPayload(values: FormValues): EdgeConfigurationCreatePayload {
  const edgeConfiguration: EdgeConfigurationCreatePayload['edgeConfiguration'] =
    {
      name: values.name,
      baseDir: values.directory,
      edgeGroupIDs: values.groupIds,
      type: translateType(values.type, values.matchingRule),
    };

  if (values.file.method === FormValuesFileMethod.Archive) {
    return {
      edgeConfiguration,
      file: values.file.content,
    };
  }

  if (values.file.method === FormValuesFileMethod.File) {
    return {
      edgeConfiguration,
      files: [],
    };
  }
  throw new PortainerError('Unknown upload type');
}

function toFormData(payload: EdgeConfigurationCreatePayload): FormData {
  const fd = new FormData();
  fd.append('edgeConfiguration', JSON.stringify(payload.edgeConfiguration));
  if (payload.file) {
    fd.append('file', payload.file);
  }
  return fd;
}

async function create(values: FormValues) {
  const payload = createPayload(values);
  const formPayload = toFormData(payload);

  try {
    const { data } = await axios.post<EdgeConfiguration>(
      buildUrl(),
      formPayload
    );

    return data;
  } catch (err) {
    throw parseAxiosError(err as Error, 'Failed to create edge configuration');
  }
}

export function useCreateMutation() {
  const queryClient = useQueryClient();
  return useMutation(create, {
    ...withInvalidate(queryClient, [queryKeys.base()]),
    ...withError(),
  });
}
