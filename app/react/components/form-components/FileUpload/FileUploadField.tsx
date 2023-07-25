import clsx from 'clsx';
import { ChangeEvent, ComponentProps, createRef } from 'react';
import { Upload, XCircle } from 'lucide-react';

import { Button } from '@@/buttons';
import { Icon } from '@@/Icon';
import { Tooltip } from '@@/Tip/Tooltip';

export interface Props {
  onChange(value: File): void;
  value?: File | null;
  accept?: string;
  title?: string;
  required?: boolean;
  inputId: string;
  dataCy?: string;
  className?: string;
  color?: ComponentProps<typeof Button>['color'];
  name?: string;
  hideFilename?: boolean;
  tooltip?: string;
}

export function FileUploadField({
  onChange,
  value,
  accept,
  title = 'Select a file',
  required = false,
  inputId,
  dataCy,
  className,
  color = 'primary',
  name,
  hideFilename,
  tooltip,
}: Props) {
  const fileRef = createRef<HTMLInputElement>();

  return (
    <div className="flex gap-2">
      <input
        id={inputId}
        ref={fileRef}
        type="file"
        accept={accept}
        required={required}
        className="!hidden"
        onChange={changeHandler}
        aria-label="file-input"
        name={name}
      />
      <Button
        size="small"
        color={color}
        onClick={handleButtonClick}
        className={clsx('!ml-0', className)}
        data-cy={dataCy}
        icon={Upload}
      >
        {title}
      </Button>
      {tooltip && <Tooltip message={tooltip} />}

      <span
        className={clsx(
          'vertical-center',
          hideFilename && !required ? 'hidden' : ''
        )}
      >
        {value
          ? !hideFilename && value.name
          : required && <Icon icon={XCircle} mode="danger" />}
      </span>
    </div>
  );

  function handleButtonClick() {
    if (fileRef && fileRef.current) {
      fileRef.current.click();
    }
  }

  function changeHandler(event: ChangeEvent<HTMLInputElement>) {
    if (event.target && event.target.files && event.target.files.length > 0) {
      onChange(event.target.files[0]);
    }
  }
}
