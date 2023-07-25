import { useState } from 'react';

import { readFileAsText } from '@/portainer/services/fileUploadReact';

import {
  FileUploadField,
  FileUploadProps,
} from '@@/form-components/FileUpload';

type Props = Omit<FileUploadProps, 'onChange' | 'value'> & {
  onChange: (value: string, fileName: string) => void;
};

export function LoadFromFileButton({ onChange, ...props }: Props) {
  const [file, setFile] = useState<File | null>(null);

  return (
    <FileUploadField
      onChange={handleChange}
      value={file}
      // eslint-disable-next-line react/jsx-props-no-spreading
      {...props}
    />
  );

  async function handleChange(file: File) {
    setFile(file);
    if (!file) {
      return;
    }

    const text = await readFileAsText(file);
    if (!text) {
      return;
    }
    onChange(text, file.name);
  }
}
