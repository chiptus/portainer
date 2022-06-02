import clsx from 'clsx';

interface Props {
  labels: string[];
  icons: string[];
  active: number;
  onClick: (buttonIndex: number) => void;
}

export function InputToggleSelector({ labels, icons, onClick, active }: Props) {
  return (
    <ul className="nav nav-pills nav-justified flex">
      {labels.map((label, i) => (
        <li key={i} className="interactive flex-1">
          <button
            type="button"
            data-cy={`${label}-button`}
            className={clsx(
              'text-blue-9 rounded text-xs border-0 py-1.5 px-2 !ml-0 w-full',
              active === i
                ? 'text-white bg-blue-2'
                : 'text-blue-9 hover:bg-grey-3'
            )}
            onClick={() => onClick(i)}
          >
            <span>
              {icons[i] && <i className={`${icons[i]} mr-1`} />}
              {label}
            </span>
          </button>
        </li>
      ))}
    </ul>
  );
}
