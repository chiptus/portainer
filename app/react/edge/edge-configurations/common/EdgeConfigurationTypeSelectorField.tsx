import { FileCode2, Lock } from 'lucide-react';
import { useField } from 'formik';

import { BoxSelector } from '@@/BoxSelector';
import { BoxSelectorOption } from '@@/BoxSelector/types';
import { TextTip } from '@@/Tip/TextTip';

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

  const textTipForGeneral =
    'Selecting this option will result in the uploaded configuration ' +
    'being sent to every edge device. The configuration will be placed in the defined directory of ' +
    'each edge device. This choice ensures uniformity across all devices, as they will share the ' +
    'same configuration.';
  const textTipForDeviceSpecific =
    'Opting for this approach means that only ' +
    'configurations matching specific criteria will be transmitted to certain edge devices. The ' +
    'uploaded configuration will be placed in the defined directory of only those devices that ' +
    'meet the criteria. This offers targeted customization, tailoring configurations to the ' +
    'unique requirements of individual devices.';
  const textTip =
    configurationType === FormValuesEdgeConfigurationType.General
      ? textTipForGeneral
      : textTipForDeviceSpecific;

  return (
    <>
      <div className="col-sm-12 form-section-title">
        <div>Type</div>
        <TextTip color="blue" inline={false} className="mt-2">
          {textTip}
        </TextTip>
      </div>

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
