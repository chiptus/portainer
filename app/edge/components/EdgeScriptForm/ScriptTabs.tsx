import { useEffect } from 'react';

import { Code } from '@/portainer/components/Code';
import { CopyButton } from '@/portainer/components/Button/CopyButton';
import { NavTabs } from '@/portainer/components/NavTabs/NavTabs';

import { EdgeProperties, Platform } from './types';
import {
  buildLinuxKubernetesCommand,
  buildLinuxSwarmCommand,
  buildLinuxNomadCommand,
  buildLinuxStandaloneCommand,
  buildWindowsStandaloneCommand,
  buildWindowsSwarmCommand,
} from './scripts';

const commandsByOs = {
  linux: [
    {
      id: 'k8s',
      label: 'Kubernetes',
      command: buildLinuxKubernetesCommand,
    },
    {
      id: 'swarm',
      label: 'Docker Swarm',
      command: buildLinuxSwarmCommand,
    },
    {
      id: 'standalone',
      label: 'Docker Standalone',
      command: buildLinuxStandaloneCommand,
    },
    {
      id: 'nomad',
      label: 'Nomad',
      command: buildLinuxNomadCommand,
    },
  ],
  win: [
    {
      id: 'swarm',
      label: 'Docker Swarm',
      command: buildWindowsSwarmCommand,
    },
    {
      id: 'standalone',
      label: 'Docker Standalone',
      command: buildWindowsStandaloneCommand,
    },
  ],
};

interface Props {
  values: EdgeProperties;
  edgeKey: string;
  agentVersion: string;
  edgeId?: string;
  agentSecret?: string;
  useAsyncMode: boolean;
  onPlatformChange(platform: Platform): void;
}

export function ScriptTabs({
  agentVersion,
  values,
  edgeKey,
  edgeId,
  agentSecret,
  useAsyncMode,
  onPlatformChange,
}: Props) {
  const { os, platform } = values;

  useEffect(() => {
    if (os && !commandsByOs[os].find((p) => p.id === platform)) {
      onPlatformChange('swarm');
    }
  }, [os, platform, onPlatformChange]);

  if (!os) {
    return null;
  }

  const options = commandsByOs[os].map((c) => {
    const cmd = c.command(
      agentVersion,
      edgeKey,
      values,
      useAsyncMode,
      edgeId,
      agentSecret
    );

    return {
      id: c.id,
      label: c.label,
      children: (
        <>
          <Code>{cmd}</Code>
          <CopyButton copyText={cmd}>Copy</CopyButton>
        </>
      ),
    };
  });

  return (
    <div className="row">
      <div className="col-sm-12">
        <NavTabs
          selectedId={platform}
          options={options}
          onSelect={(id: Platform) => onPlatformChange(id)}
        />
      </div>
    </div>
  );
}
