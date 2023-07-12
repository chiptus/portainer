import clsx from 'clsx';
import { ComponentProps } from 'react';

import { FeatureId } from '@/react/portainer/feature-flags/enums';

import { Tooltip } from '@@/Tip/Tooltip';

import styles from './SwitchField.module.css';
import { Switch } from './Switch';

export interface Props {
  label: string;
  checked: boolean;
  onChange(value: boolean, index?: number): void;

  index?: number;
  name?: string;
  tooltip?: ComponentProps<typeof Tooltip>['message'];
  setTooltipHtmlMessage?: ComponentProps<typeof Tooltip>['setHtmlMessage'];
  labelClass?: string;
  switchClass?: string;
  fieldClass?: string;
  dataCy?: string;
  disabled?: boolean;
  featureId?: FeatureId;
}

export function SwitchField({
  tooltip,
  checked,
  label,
  index,
  name,
  labelClass,
  fieldClass,
  dataCy,
  disabled,
  onChange,
  featureId,
  switchClass,
  setTooltipHtmlMessage,
}: Props) {
  const toggleName = name ? `toggle_${name}` : '';
  return (
    <label className={clsx(styles.root, fieldClass)}>
      <span
        className={clsx('space-right control-label !p-0 text-left', labelClass)}
      >
        {label}
        {tooltip && (
          <Tooltip message={tooltip} setHtmlMessage={setTooltipHtmlMessage} />
        )}
      </span>
      <Switch
        className={clsx('space-right shrink-0', switchClass)}
        name={toggleName}
        id={toggleName}
        checked={checked}
        disabled={disabled}
        onChange={onChange}
        featureId={featureId}
        index={index}
        dataCy={dataCy}
      />
    </label>
  );
}
