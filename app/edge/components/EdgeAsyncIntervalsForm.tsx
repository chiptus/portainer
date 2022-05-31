import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Select } from '@/portainer/components/form-components/Input';
import { r2a } from '@/react-tools/react2angular';

import { Options, useIntervalOptions } from './useIntervalOptions';

export const EDGE_ASYNC_INTERVAL_USE_DEFAULT = -1;

interface Values {
  PingInterval: number;
  SnapshotInterval: number;
  CommandInterval: number;
}

interface Props {
  values: Values;
  isDefaultHidden?: boolean;
  onChange(value: Values): void;
}

export const options: Options = [
  { label: 'Use default interval', value: -1, isDefault: true },
  {
    value: 0,
    label: 'disabled',
  },
  {
    value: 60,
    label: '1 minute',
  },
  {
    value: 60 * 60,
    label: '1 hour',
  },
  {
    value: 24 * 60 * 60,
    label: '1 day',
  },
  {
    value: 7 * 24 * 60 * 60,
    label: '1 week',
  },
];

export function EdgeAsyncIntervalsForm({
  onChange,
  values,
  isDefaultHidden = false,
}: Props) {
  const pingIntervalOptions = useIntervalOptions(
    'Edge.PingInterval',
    options,
    isDefaultHidden
  );

  const snapshotIntervalOptions = useIntervalOptions(
    'Edge.SnapshotInterval',
    options,
    isDefaultHidden
  );

  const commandIntervalOptions = useIntervalOptions(
    'Edge.CommandInterval',
    options,
    isDefaultHidden
  );

  return (
    <>
      <FormControl
        inputId="edge_checkin_ping"
        label="Edge agent default ping frequency"
        tooltip="Interval used by default by each Edge agent to ping the Portainer instance. Affects Edge environment management and Edge compute features."
      >
        <Select
          value={values.PingInterval}
          name="PingInterval"
          onChange={handleChange}
          options={pingIntervalOptions}
        />
      </FormControl>

      <FormControl
        inputId="edge_checkin_snapshot"
        label="Edge agent default snapshot frequency"
        tooltip="Interval used by default by each Edge agent to snapshot the agent state."
      >
        <Select
          value={values.SnapshotInterval}
          name="SnapshotInterval"
          onChange={handleChange}
          options={snapshotIntervalOptions}
        />
      </FormControl>

      <FormControl
        inputId="edge_checkin_command"
        label="Edge agent default command frequency"
        tooltip="Interval used by default by each Edge agent to fetch commands from the Portainer instance"
      >
        <Select
          value={values.CommandInterval}
          name="CommandInterval"
          onChange={handleChange}
          options={commandIntervalOptions}
        />
      </FormControl>
    </>
  );

  function handleChange(e: React.ChangeEvent<HTMLSelectElement>) {
    onChange({ ...values, [e.target.name]: parseInt(e.target.value, 10) });
  }
}

export const EdgeAsyncIntervalsFormAngular = r2a(EdgeAsyncIntervalsForm, [
  'values',
  'onChange',
  'isDefaultHidden',
]);
