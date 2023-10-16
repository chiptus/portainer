import { useField } from 'formik';
import { Trash2 } from 'lucide-react';

import { capitalize } from '@/react/common/string-utils';
import { EdgeConfigurationCategory } from '@/react/edge/edge-configurations/types';

import { FormSection } from '@@/form-components/FormSection';
import { Button } from '@@/buttons';
import { WebEditorForm } from '@@/WebEditorForm';
import { FileUploadField } from '@@/form-components/FileUpload';

import { FormValues, FormValuesFileMethod } from '../types';
import { InputField } from '../InputField';

// import { LoadFromFileButton } from './LoadFromFileButton';

export function ConfigurationFieldset({
  category,
}: {
  category: EdgeConfigurationCategory;
}) {
  const [{ value: file }, { error: fileError }, { setValue: setFile }] =
    useField<FormValues['file']>('file');

  return (
    <FormSection title={capitalize(category)}>
      <div className="form-group">
        <div className="col-sm-12">
          <div className="flex items-center gap-2">
            {/* <Button
              icon={Plus}
              color="light"
              className="!ml-0"
              onClick={() =>
                setFile({
                  name: '',
                  method: FormValuesFileMethod.File,
                  content: '',
                })
              }
            >
              Create configuration
            </Button>
            <LoadFromFileButton
              inputId="load-from-file"
              title="Upload from file"
              color="light"
              hideFilename
              onChange={(content, fileName) =>
                setFile({
                  name: fileName,
                  method: FormValuesFileMethod.File,
                  content,
                })
              }
            /> */}
            <FileUploadField
              inputId="load-package"
              title="Upload from package"
              color="light"
              hideFilename
              // accept="application/zip,application/gzip,application/x-tar,application/x-gtar"
              accept="application/zip"
              onChange={(archive) =>
                setFile({
                  name: archive.name,
                  method: FormValuesFileMethod.Archive,
                  content: archive,
                })
              }
              tooltip={`You can upload your ${category} file as a compressed package in the following format: ZIP. This format allows you to bundle multiple files together into a single package, making it easier to transfer your ${category} files.`}
              // tooltip="You can upload your configuration file as a compressed package in one of the following formats: ZIP, GZIP, TAR, TGZ. These formats allow you to bundle multiple files together into a single package, making it easier to transfer your configuration files."
            />
          </div>
        </div>
      </div>

      {file.method === 'file' && (
        <>
          <InputField fieldName="file.name" label="Name" required />
          <div className="form-group">
            <div className="col-sm-12">
              <WebEditorForm
                value={file.content}
                id="configuration-editor"
                placeholder={`Define or paste the content of your ${category} file here`}
                onChange={(v) =>
                  setFile({
                    name: file.name,
                    method: FormValuesFileMethod.File,
                    content: v,
                  })
                }
                error={(fileError as { content?: string })?.content}
              />
            </div>
          </div>
          <div className="form-group">
            <div className="col-sm-12">
              <Button
                icon={Trash2}
                color="dangerlight"
                onClick={() => setFile({ name: '' })}
              >
                Remove {category}
              </Button>
            </div>
          </div>
        </>
      )}
      {file.method === 'archive' && (
        <div className="form-group">
          <div className="col-sm-12">
            <div className="flex items-center gap-2 rounded-lg bg-[color:var(--bg-webeditor-color)] p-2.5">
              <span>Uploaded package:</span>
              <span>{file.name}</span>
              <Button
                icon={Trash2}
                color="dangerlight"
                onClick={() => setFile({ name: '' })}
              />
            </div>
          </div>
        </div>
      )}
    </FormSection>
  );
}
