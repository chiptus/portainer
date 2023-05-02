import { Meta, Story } from '@storybook/react';

import { LicenseExpirationPanel } from './LicenseExpirationPanel';

export default {
  title: 'Components/Header/LicenseExpirationPanel',
  args: {
    remainingDays: 15,
  },
} as Meta;

interface StoryProps {
  remainingDays: number;
  noValidLicense: boolean;
}

function Template({ remainingDays, noValidLicense }: StoryProps) {
  return (
    <LicenseExpirationPanel
      noValidLicense={noValidLicense}
      remainingDays={remainingDays}
    />
  );
}

export const Example: Story<StoryProps> = Template.bind({});
export const ExpiredLicense: Story<StoryProps> = Template.bind({});
ExpiredLicense.args = {
  remainingDays: -1,
};

export const ExpiringToday: Story<StoryProps> = Template.bind({});
ExpiringToday.args = {
  remainingDays: 0,
};
