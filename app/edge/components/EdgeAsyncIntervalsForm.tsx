import { FormControl } from '@/portainer/components/form-components/FormControl';
import { Select } from '@/portainer/components/form-components/Input';
import { r2a } from '@/react-tools/react2angular';

interface Values {
  EdgeAgentCheckinInterval: number;
  EdgePingInterval: number;
  EdgeSnapshotInterval: number;
  EdgeCommandInterval: number;
}

interface Props {
  values: Values;

  setFieldValue<T>(field: string, value: T): void;
}

export const options = [
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

export function EdgeAsyncIntervalsForm({ setFieldValue, values }: Props) {
  return (
    <>
      <FormControl
        inputId="edge_checkin_ping"
        label="Edge agent default ping frequency"
        tooltip="Interval used by default by each Edge agent to ping the Portainer instance. Affects Edge environment management and Edge compute features."
      >
        <Select
          value={values.EdgePingInterval}
          name="EdgePingInterval"
          onChange={handleChange}
          options={options}
        />
      </FormControl>

      <FormControl
        inputId="edge_checkin_snapshot"
        label="Edge agent default snapshot frequency"
        tooltip="Interval used by default by each Edge agent to snapshot the agent state."
      >
        <Select
          value={values.EdgeSnapshotInterval}
          name="EdgeSnapshotInterval"
          onChange={handleChange}
          options={options}
        />
      </FormControl>

      <FormControl
        inputId="edge_checkin_command"
        label="Edge agent default command frequency"
        tooltip="Interval used by default by each Edge agent to fetch commands from the Portainer instance"
      >
        <Select
          value={values.EdgeCommandInterval}
          name="EdgeCommandInterval"
          onChange={handleChange}
          options={options}
        />
      </FormControl>
    </>
  );

  function handleChange(e: React.ChangeEvent<HTMLSelectElement>) {
    setFieldValue(e.target.name, parseInt(e.target.value, 10));
  }
}

export const EdgeAsyncIntervalsFormAngular = r2a(EdgeAsyncIntervalsForm, [
  'values',
  'setFieldValue',
]);
