import { PropsWithChildren } from 'react';
import saveAs from 'file-saver';
import { Download } from 'lucide-react';

import { notifyError } from '@/portainer/services/notifications';

import { Button, type Props as ButtonProps } from '../Button';

export interface Props extends ButtonProps {
  fileName: string;
  fileContent: string;
  errorMessage?: string;
  options?: BlobPropertyBag;
  saveFunction?: typeof saveAs; // optional prop to allow mocking saveAs in tests
}

export function DownloadButton({
  fileName,
  fileContent,
  children,
  errorMessage = 'Unable to download file',
  size = 'small',
  icon = Download,
  options,
  saveFunction = saveAs,
  ...rest
}: PropsWithChildren<Props>) {
  return (
    <Button
      // eslint-disable-next-line react/jsx-props-no-spreading
      {...rest}
      size={size}
      icon={icon}
      disabled={!fileContent || !fileName}
      onClick={() => {
        try {
          saveFunction(new Blob([fileContent], options), fileName);
        } catch (error) {
          notifyError(errorMessage, error as Error);
        }
      }}
    >
      {children}
    </Button>
  );
}
