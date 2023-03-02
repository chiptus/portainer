import { Meta, Story } from '@storybook/react';
import { PropsWithChildren } from 'react';

import { DownloadButton, Props } from './DownloadButton';

export default {
  component: DownloadButton,
  title: 'Components/Buttons/DownloadButton',
} as Meta;

function Template({
  fileContent,
  fileName,
  children,
  ...rest
}: JSX.IntrinsicAttributes & PropsWithChildren<Props>) {
  return (
    // eslint-disable-next-line react/jsx-props-no-spreading
    <DownloadButton fileContent={fileContent} fileName={fileName} {...rest}>
      {children}
    </DownloadButton>
  );
}

export const Primary: Story<PropsWithChildren<Props>> = Template.bind({});
Primary.args = {
  children: 'Download',
  fileContent: 'Some basic file content',
  fileName: 'test.txt',
};
