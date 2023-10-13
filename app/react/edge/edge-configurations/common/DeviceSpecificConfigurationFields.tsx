import { useField } from 'formik';
import { Info } from 'lucide-react';

import { useCategory } from '@/react/edge/edge-configurations/CreateView/useCategory';

import { FormSection } from '@@/form-components/FormSection';
import { TextTip } from '@@/Tip/TextTip';
import { FormControl } from '@@/form-components/FormControl';
import { Option, PortainerSelect } from '@@/form-components/PortainerSelect';

import { FormValues, FormValuesEdgeConfigurationMatchingRule } from './types';

const options: Option<FormValuesEdgeConfigurationMatchingRule>[] = [
  {
    label: 'Match file name with Portainer Edge ID',
    value: FormValuesEdgeConfigurationMatchingRule.MatchFile,
  },
  {
    label: 'Match folder name with Portainer Edge ID',
    value: FormValuesEdgeConfigurationMatchingRule.MatchFolder,
  },
];

export function DeviceSpecificConfigurationFields() {
  const [category] = useCategory();

  const [{ value }, { error }, { setValue }] =
    useField<FormValues['matchingRule']>('matchingRule');

  return (
    <FormSection title="Target devices">
      <TextTip color="blue" icon={Info}>
        Select the rule that you want to use for matching {category} with
        Portainer Edge ID
      </TextTip>
      <FormControl label="Matching rule" errors={error}>
        <PortainerSelect onChange={setValue} value={value} options={options} />
      </FormControl>
    </FormSection>
  );
}
