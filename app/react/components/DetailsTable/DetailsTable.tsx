import clsx from 'clsx';
import { PropsWithChildren } from 'react';

type Props = {
  headers?: string[];
  dataCy?: string;
  className?: string;
};

export function DetailsTable({
  headers = [],
  dataCy,
  className,
  children,
}: PropsWithChildren<Props>) {
  return (
    <table className={clsx('table', className)} data-cy={dataCy}>
      {headers.length > 0 && (
        <thead>
          <tr>
            {headers.map((header) => (
              <th key={header}>{header}</th>
            ))}
          </tr>
        </thead>
      )}
      <tbody>{children}</tbody>
    </table>
  );
}
