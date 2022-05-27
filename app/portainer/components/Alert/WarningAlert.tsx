import { PropsWithChildren } from 'react';

interface Props {
  showIcon?: boolean;
  iconClasses?: string;
  containerStyles?: Record<string, string>;
  iconStyles?: Record<string, string>;
}

export function WarningAlert({
  children,
  showIcon = true,
  iconClasses = 'fa fa-exclamation-triangle orange-icon space-right',
  containerStyles = {},
  iconStyles = {},
}: PropsWithChildren<Props>) {
  return (
    <div className="small text-warning" style={containerStyles}>
      {showIcon && (
        <i className={iconClasses} aria-hidden="true" style={iconStyles} />
      )}
      {children}
    </div>
  );
}
