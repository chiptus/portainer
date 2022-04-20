import { EnvironmentType } from '@/portainer/environments/types';

import { EditorType } from './types';
import { getValidEditorTypes } from './utils';

interface GetValidEditorTypesTest {
  endpointTypes: EnvironmentType[];
  expected: EditorType[];
  title: string;
}

describe('getValidEditorTypes', () => {
  const tests: GetValidEditorTypesTest[] = [
    {
      endpointTypes: [EnvironmentType.EdgeAgentOnDocker],
      expected: [EditorType.Compose],
      title: 'should return compose for docker envs',
    },
    {
      endpointTypes: [EnvironmentType.EdgeAgentOnKubernetes],
      expected: [EditorType.Compose, EditorType.Kubernetes],
      title: 'should return compose and kubernetes for kubernetes envs',
    },
    {
      endpointTypes: [EnvironmentType.EdgeAgentOnNomad],
      expected: [EditorType.Nomad],
      title: 'should return nomad for nomad envs',
    },
    {
      endpointTypes: [
        EnvironmentType.EdgeAgentOnDocker,
        EnvironmentType.EdgeAgentOnKubernetes,
      ],
      expected: [EditorType.Compose],
      title: 'should return compose for docker and kubernetes envs',
    },
    {
      endpointTypes: [
        EnvironmentType.EdgeAgentOnDocker,
        EnvironmentType.EdgeAgentOnNomad,
      ],
      expected: [],
      title: 'should return empty for docker and nomad envs',
    },
    {
      endpointTypes: [
        EnvironmentType.EdgeAgentOnKubernetes,
        EnvironmentType.EdgeAgentOnNomad,
      ],
      expected: [],
      title: 'should return empty for kubernetes and nomad envs',
    },
    {
      endpointTypes: [
        EnvironmentType.EdgeAgentOnDocker,
        EnvironmentType.EdgeAgentOnKubernetes,
        EnvironmentType.EdgeAgentOnNomad,
      ],
      expected: [],
      title: 'should return empty for all envs',
    },
  ];

  tests.forEach((test) => {
    // eslint-disable-next-line jest/valid-title
    it(test.title, () => {
      expect(getValidEditorTypes(test.endpointTypes)).toEqual(test.expected);
    });
  });
});
