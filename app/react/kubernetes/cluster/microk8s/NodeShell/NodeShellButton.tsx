import { useState } from 'react';
import { createPortal } from 'react-dom';
import { Terminal } from 'lucide-react';

import { EnvironmentId } from '@/react/portainer/environments/types';
import { useAnalytics } from '@/react/hooks/useAnalytics';

import { Button } from '@@/buttons';

import { NodeShell } from './NodeShell';

interface Props {
  windowTitle: string;
  btnTitle?: string;
  environmentId: EnvironmentId;
  nodeIp: string;
}
export function NodeShellButton({
  windowTitle,
  btnTitle,
  environmentId,
  nodeIp,
}: Props) {
  const [open, setOpen] = useState(false);
  const { trackEvent } = useAnalytics();
  return (
    <>
      <Button
        title={btnTitle}
        color="none"
        size="small"
        disabled={open}
        data-cy="nodeShellButton"
        onClick={() => handleOpen()}
        className="!text-blue-8"
        icon={Terminal}
      />

      {open &&
        createPortal(
          <NodeShell
            title={windowTitle}
            environmentId={environmentId}
            nodeIp={nodeIp}
            onClose={() => setOpen(false)}
          />,
          document.body
        )}
    </>
  );

  function handleOpen() {
    setOpen(true);

    trackEvent('microk8s-shell', { category: 'kubernetes' });
  }
}
