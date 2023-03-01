import { Plus } from 'lucide-react';

import { Button } from '@@/buttons';
import { TooltipWithChildren } from '@@/Tip/TooltipWithChildren';

interface Props {
  addNewAnnotation: (type?: 'rewrite' | 'regex' | 'ingressClass') => void;
  ingressType?: string;
  hideForm?: boolean;
}

export function IngressActions({
  addNewAnnotation,
  ingressType,
  hideForm,
}: Props) {
  return (
    <div className="col-sm-12 anntation-actions p-0">
      <TooltipWithChildren message="Use annotations to configure options for an ingress. Review Nginx or Traefik documentation to find the annotations supported by your choice of ingress type.">
        <span>
          <Button
            className="btn btn-sm btn-light mb-2 !ml-0"
            onClick={() => addNewAnnotation()}
            icon={Plus}
          >
            {' '}
            Add annotation
          </Button>
        </span>
      </TooltipWithChildren>

      {ingressType === 'nginx' && (
        <>
          <TooltipWithChildren message="When the exposed URLs for your applications differ from the specified paths in the ingress, use the rewrite target annotation to denote the path to redirect to.">
            <span>
              <Button
                className="btn btn-sm btn-light mb-2 ml-2"
                onClick={() => addNewAnnotation('rewrite')}
                icon={Plus}
                data-cy="add-rewrite-annotation"
              >
                Add rewrite annotation
              </Button>
            </span>
          </TooltipWithChildren>

          <TooltipWithChildren message="Enable use of regular expressions in ingress paths (set in the ingress details of an application). Use this along with rewrite-target to specify the regex capturing group to be replaced, e.g. path regex of ^/foo/(,*) and rewrite-target of /bar/$1 rewrites example.com/foo/account to example.com/bar/account.">
            <span>
              <Button
                className="btn btn-sm btn-light mb-2 ml-2"
                onClick={() => addNewAnnotation('regex')}
                icon={Plus}
                data-cy="add-regex-annotation"
              >
                Add regular expression annotation
              </Button>
            </span>
          </TooltipWithChildren>
        </>
      )}

      {ingressType === 'custom' && (
        <Button
          className="btn btn-sm btn-light mb-2 ml-2"
          onClick={() => addNewAnnotation('ingressClass')}
          icon={Plus}
          data-cy="add-ingress-class-annotation"
          disabled={hideForm}
        >
          Add kubernetes.io/ingress.class annotation
        </Button>
      )}
    </div>
  );
}
