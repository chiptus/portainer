import { FileCode2, Lock } from 'lucide-react';
import { useField } from 'formik';

import { BoxSelector } from '@@/BoxSelector';
import { BoxSelectorOption } from '@@/BoxSelector/types';

import { FormValues, FormValuesEdgeConfigurationType } from './types';
import { DeviceSpecificConfigurationFields } from './DeviceSpecificConfigurationFields';

const deploymentOptions: BoxSelectorOption<FormValuesEdgeConfigurationType>[] =
  [
    {
      id: FormValuesEdgeConfigurationType.General,
      icon: FileCode2,
      label: 'General configuration',
      description: 'This type of configuration apply to all devices',
      value: FormValuesEdgeConfigurationType.General,
      iconType: 'badge',
    },
    {
      id: FormValuesEdgeConfigurationType.DeviceSpecific,
      icon: Lock,
      label: 'Device specific configuration',
      description: 'This type of configuration apply to specific devices',
      value: FormValuesEdgeConfigurationType.DeviceSpecific,
      iconType: 'badge',
    },
  ];

export function EdgeConfigurationTypeSelectorField() {
  const [{ value: configurationType }, , { setValue }] =
    useField<FormValues['type']>('type');

  return (
    <>
      <div className="col-sm-12 form-section-title">Type</div>
      <BoxSelector
        radioName="configurationType"
        value={configurationType}
        options={deploymentOptions}
        onChange={(v) => setValue(v)}
        slim
      />
      {configurationType === FormValuesEdgeConfigurationType.DeviceSpecific && (
        <DeviceSpecificConfigurationFields />
      )}
    </>
  );
}
