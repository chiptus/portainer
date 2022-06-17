import { ReactNode } from 'react';

import { Widget, WidgetBody } from '@@/Widget';

interface Props {
  value?: number;
  icon: string;
  type: string;
  children?: ReactNode;
}

export function DashboardItem({ value, icon, type, children }: Props) {
  return (
    <div className="col-sm-12 col-md-6" aria-label={type}>
      <Widget>
        <WidgetBody>
          <div className="widget-icon blue pull-left">
            <i className={icon} aria-hidden="true" aria-label="icon" />
          </div>
          <div className="pull-right">{children}</div>
          <div className="title" aria-label="value">
            {typeof value !== 'undefined' ? value : '-'}
          </div>
          <div className="comment" aria-label="resourceType">
            {type}
          </div>
        </WidgetBody>
      </Widget>
    </div>
  );
}
