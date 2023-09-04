import { ChangeEvent, ReactNode } from 'react';
import { Trash2 } from 'lucide-react';
import clsx from 'clsx';

import { FormError } from '@@/form-components/FormError';
import { Button } from '@@/buttons';

import { ReadOnly } from './ReadOnly';
import { Annotation } from './types';

export type AnnotationErrors = Record<string, ReactNode>;

interface Props {
  annotations: Annotation[];
  handleAnnotationChange: (
    index: number,
    key: 'Key' | 'Value',
    val: string
  ) => void;
  removeAnnotation: (index: number) => void;
  errors: AnnotationErrors;
  placeholder: string[];
  disabled?: boolean;
  screen?: string;
}

export function AnnotationsForm({
  annotations,
  handleAnnotationChange,
  removeAnnotation,
  errors,
  placeholder,
  disabled,
  screen,
}: Props) {
  if (disabled && screen !== 'ingress') {
    return <ReadOnly annotations={annotations} />;
  }

  return (
    <>
      {annotations.map((annotation, i) => (
        <div className="row mb-4" key={annotation.ID}>
          <div className="form-group col-sm-4 !m-0 !pl-0">
            <div className="input-group input-group-sm">
              <span
                className={clsx(
                  'input-group-addon',
                  disabled ? '' : 'required'
                )}
              >
                Key
              </span>
              <input
                name={`annotation_key_${i}`}
                type="text"
                className="form-control form-control-sm"
                placeholder={placeholder[0]}
                defaultValue={annotation.Key}
                onChange={(e: ChangeEvent<HTMLInputElement>) =>
                  handleAnnotationChange(i, 'Key', e.target.value)
                }
                disabled={disabled}
              />
            </div>
            {errors[`annotations.key[${i}]`] && (
              <FormError className="!mb-0 mt-1">
                {errors[`annotations.key[${i}]`]}
              </FormError>
            )}
          </div>
          <div className="form-group col-sm-4 !m-0 !pl-0">
            <div className="input-group input-group-sm">
              <span
                className={clsx(
                  'input-group-addon',
                  disabled ? '' : 'required'
                )}
              >
                Value
              </span>
              <input
                name={`annotation_value_${i}`}
                type="text"
                className="form-control form-control-sm"
                placeholder={placeholder[1]}
                defaultValue={annotation.Value}
                onChange={(e: ChangeEvent<HTMLInputElement>) =>
                  handleAnnotationChange(i, 'Value', e.target.value)
                }
                disabled={disabled}
              />
            </div>
            {errors[`annotations.value[${i}]`] && (
              <FormError className="!mb-0 mt-1">
                {errors[`annotations.value[${i}]`]}
              </FormError>
            )}
          </div>
          {!disabled && (
            <div className="col-sm-3 !m-0 !pl-0">
              <Button
                size="medium"
                color="dangerlight"
                className="btn-only-icon !ml-0"
                type="button"
                onClick={() => removeAnnotation(i)}
                icon={Trash2}
              />
            </div>
          )}
        </div>
      ))}
    </>
  );
}
