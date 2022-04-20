import { IComponentOptions } from 'angular';
import './Tooltip.css';

export const TooltipAngular: IComponentOptions = {
  bindings: {
    message: '@',
    position: '@',
    iconClassName: '@?',
  },
  template: `<span
    class="interactive"
    tooltip-append-to-body="true"
    tooltip-placement="{{$ctrl.position}}"
    tooltip-class="portainer-tooltip"
    uib-tooltip="{{$ctrl.message}}"
  >
    <i
      class="{{$ctrl.iconClassName || 'fa fa-question-circle blue-icon tooltip-icon'}}"
      aria-hidden="true"
    ></i>
  </span>`,
};
