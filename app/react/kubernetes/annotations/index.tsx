import { useEffect, useState, useRef } from 'react';
import { Plus } from 'lucide-react';
import { v4 as uuidv4 } from 'uuid';
import { debounce } from 'lodash';

import { Button } from '@@/buttons';
import { Tooltip } from '@@/Tip/Tooltip';

import { Annotation, AnnotationErrors } from './types';
import { IngressActions } from './IngressActions';
import { AnnotationsForm } from './AnnotationsForm';

interface Props {
  index?: number;
  initialAnnotations: Annotation[];
  errors?: AnnotationErrors;
  placeholder?: string[];
  handleUpdateAnnotations?: (
    annotations: Annotation[],
    index: number | undefined
  ) => void;

  hideForm?: boolean;
  ingressType?: string;

  screen?: string;
}

export function Annotations({
  initialAnnotations,
  hideForm,
  errors = [],
  placeholder = ['e.g. app.kubernetes.io/name', 'e.g. examplename'],
  ingressType,
  handleUpdateAnnotations = () => {},
  screen,
  index = undefined,
}: Props) {
  const [annotations, setAnnotations] =
    useState<Annotation[]>(initialAnnotations);

  const debouncedHandleUpdateAnnotations = useRef(
    debounce(handleUpdateAnnotations, 300)
  );

  useEffect(() => {
    debouncedHandleUpdateAnnotations.current(annotations, index);
  }, [annotations, index]);

  return (
    <>
      <div className="col-sm-12 text-muted vertical-center mb-2 block px-0">
        <div className="control-label !mb-2 text-left font-medium">
          Annotations
          {!hideForm && (
            <Tooltip
              message={
                <div className="vertical-center">
                  <span>
                    Allows specifying of{' '}
                    <a
                      href="https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/"
                      target="_black"
                    >
                      annotations
                    </a>{' '}
                    for the object. See further Kubernetes documentation on{' '}
                    <a
                      href="https://kubernetes.io/docs/reference/labels-annotations-taints/"
                      target="_black"
                    >
                      well-known annotations
                    </a>
                    .
                  </span>
                </div>
              }
            />
          )}
        </div>
      </div>

      {annotations && (
        <AnnotationsForm
          placeholder={placeholder}
          annotations={annotations}
          handleAnnotationChange={handleAnnotationChange}
          removeAnnotation={removeAnnotation}
          errors={errors}
          disabled={hideForm}
          screen={screen}
        />
      )}

      {!hideForm && screen === 'ingress' && (
        <IngressActions
          addNewAnnotation={addNewAnnotation}
          ingressType={ingressType}
          hideForm={hideForm}
        />
      )}

      {!hideForm && !screen && (
        <div className="col-sm-12 p-0">
          <Button
            className="btn btn-sm btn-light !ml-0 mb-2"
            onClick={() => addNewAnnotation()}
            icon={Plus}
          >
            Add annotation
          </Button>
        </div>
      )}
    </>
  );

  function addNewAnnotation(type?: 'rewrite' | 'regex' | 'ingressClass') {
    const newAnnotations = [...annotations];
    const annotation: Annotation = {
      Key: '',
      Value: '',
      ID: uuidv4(),
    };
    switch (type) {
      case 'rewrite':
        annotation.Key = 'nginx.ingress.kubernetes.io/rewrite-target';
        annotation.Value = '/$1';
        break;
      case 'regex':
        annotation.Key = 'nginx.ingress.kubernetes.io/use-regex';
        annotation.Value = 'true';
        break;
      case 'ingressClass':
        annotation.Key = 'kubernetes.io/ingress.class';
        annotation.Value = '';
        break;
      default:
        break;
    }
    newAnnotations?.push(annotation);
    setAnnotations(newAnnotations);
  }

  function removeAnnotation(index: number) {
    const newAnnotations = [...annotations];

    if (index > -1) {
      newAnnotations?.splice(index, 1);
    }

    setAnnotations(newAnnotations);
  }

  function handleAnnotationChange(
    index: number,
    key: 'Key' | 'Value',
    val: string
  ) {
    setAnnotations((prevAnnotations) => {
      const oldAnnotations = [...prevAnnotations];

      oldAnnotations[index] = oldAnnotations[index] || {
        Key: '',
        Value: '',
      };
      oldAnnotations[index][key] = val;

      return oldAnnotations;
    });
  }
}
