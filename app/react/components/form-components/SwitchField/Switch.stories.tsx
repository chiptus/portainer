import { Meta, Story } from '@storybook/react';
import { useState } from 'react';

import { Switch } from './Switch';

export default {
  title: 'Components/Form/SwitchField/Switch',
} as Meta;

export function Example() {
  const [isChecked, setIsChecked] = useState(false);
  function onChange() {
    setIsChecked(!isChecked);
  }

  return <Switch name="name" checked={isChecked} onChange={onChange} id="id" />;
}

interface Args {
  checked: boolean;
  disabled?: boolean;
}

function Template({ checked, disabled }: Args) {
  return (
    <Switch
      name="name"
      checked={checked}
      onChange={() => {}}
      id="id"
      disabled={disabled}
    />
  );
}

export const Checked: Story<Args> = Template.bind({});
Checked.args = {
  checked: true,
};

export const Unchecked: Story<Args> = Template.bind({});
Unchecked.args = {
  checked: false,
};

export const UncheckedDisabled: Story<Args> = Template.bind({});
UncheckedDisabled.args = {
  checked: false,
  disabled: true,
};

export const CheckedDisabled: Story<Args> = Template.bind({});
CheckedDisabled.args = {
  checked: true,
  disabled: true,
};
