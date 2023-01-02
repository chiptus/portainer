import { useEffect, useState } from 'react';

import { Icon } from '@@/Icon';
import { Button } from '@@/buttons';
import { useLogViewerContext } from '@@/LogViewer/context';

export function FullScreenButton() {
  const { logViewerRef } = useLogViewerContext();

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
