import { useContext, useEffect, useState } from 'react';

import { Icon } from '@@/Icon';
import { Button } from '@@/buttons';
import {
  LogViewerContext,
  LogViewerContextInterface,
} from '@@/LogViewer/types';

export function FullScreenButton() {
  const { logViewerRef } = useContext(
    LogViewerContext
  ) as LogViewerContextInterface;

  const [isFullScreen, setIsFullScreen] = useState(false);

  async function toggleFullScreen() {
    if (isFullScreen) {
      await document.exitFullscreen();
    } else {
      await logViewerRef.current?.requestFullscreen();
    }
  }

  useEffect(() => {
    document.onfullscreenchange = () => {
      setIsFullScreen(!!document.fullscreenElement);
    };
    return () => {
      document.onfullscreenchange = null;
    };
  }, []);

  return (
    <Button
      onClick={toggleFullScreen}
      disabled={!document.fullscreenEnabled}
      color="none"
      title={isFullScreen ? 'Exit full screen' : 'Full screen'}
    >
      <Icon icon={isFullScreen ? 'minimize-2' : 'maximize-2'} />
    </Button>
  );
}
