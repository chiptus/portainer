import RcSlider from 'rc-slider';
import clsx from 'clsx';

import styles from './Slider.module.css';
import 'rc-slider/assets/index.css';

export interface Props {
  min: number;
  max: number;
  step: number;
  value: number;
  onChange: (value: number) => void;
  // true if you want to always show the tooltip
  dataCy: string;
  visibleTooltip?: boolean;
  disabled?: boolean;
}

export function Slider({
  min,
  max,
  step,
  value,
  onChange,
  dataCy,
  visibleTooltip: visible,
  disabled,
}: Props) {
  const SliderWithTooltip = RcSlider.createSliderWithTooltip(RcSlider);
  // if the tooltip is always visible, hide the marks when tooltip value gets close to the edges
  const marks = {
    [min]: visible && value / max < 0.1 ? '' : translateMinValue(min),
    [max]: visible && value / max > 0.9 ? '' : max.toString(),
  };

  return (
    <div className={styles.root}>
      <SliderWithTooltip
        tipFormatter={translateMinValue}
        min={min}
        max={max}
        step={step}
        marks={marks}
        defaultValue={value}
        onAfterChange={onChange}
        className={clsx(
          styles.slider,
          disabled && 'opacity-30 th-highcontrast:opacity-75'
        )}
        tipProps={{ visible }}
        railStyle={{ height: 8 }}
        trackStyle={{ height: 8 }}
        dotStyle={{ visibility: 'hidden' }}
        disabled={disabled}
        data-cy={dataCy}
      />
    </div>
  );
}

function translateMinValue(value: number) {
  if (value === 0) {
    return 'unlimited';
  }
  return value.toString();
}
