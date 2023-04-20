import { Meta, Story } from '@storybook/react';

import { LicenseExpirationPanel } from './LicenseExpirationPanel';

export default {
  title: 'Components/Header/LicenseExpirationPanel',
  args: {
    remainingDays: 15,
  },
} as Meta;

interface StoryProps {
  nodes: number;
  remainingDays: number;
}

function Template({ remainingDays, nodes }: StoryProps) {
  return <LicenseExpirationPanel nodes={nodes} remainingDays={remainingDays} />;
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
