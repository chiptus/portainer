import { FormikErrors } from 'formik';

import { FeatureId } from '@/react/portainer/feature-flags/enums';

import { FormControl } from '@@/form-components/FormControl';
import { FormSection } from '@@/form-components/FormSection';
import { Input } from '@@/form-components/Input/Input';
import { SwitchField } from '@@/form-components/SwitchField';
import { TextTip } from '@@/Tip/TextTip';

import { LoadBalancerQuotaFormValues } from './types';

interface Props {
  values: LoadBalancerQuotaFormValues;
  onChange: (value: LoadBalancerQuotaFormValues) => void;
  errors?: FormikErrors<LoadBalancerQuotaFormValues>;
}

export function LoadBalancerFormSection({
  values: loadBalancerQuota,
  onChange,
  errors,
}: Props) {
  return (
    <FormSection title="Load balancers">
      <TextTip color="blue">
        You can set a quota on the number of external load balancers that can be
        created inside this namespace. Set this quota to 0 to effectively
        disable the use of load balancers in this namespace.
      </TextTip>
      <SwitchField
        dataCy="k8sNamespaceCreate-loadBalancerQuotaToggle"
        label="Load balancer quota"
        labelClass="col-sm-3 col-lg-2"
        fieldClass="pt-2"
        checked={loadBalancerQuota.enabled}
        onChange={(enabled) => {
          if (enabled) {
            onChange({ limit: 0, enabled });
            return;
          }
          onChange({ ...loadBalancerQuota, enabled });
        }}
        featureId={FeatureId.K8S_ANNOTATIONS}
      />
      {loadBalancerQuota.enabled && (
        <div className="pt-5">
          <FormControl
            label="Max load balancers"
            inputId="loadbalancers"
            errors={errors?.limit}
            required
          >
            <Input
              required
              type="number"
              min={0}
              value={loadBalancerQuota.limit}
              onChange={(event) =>
                onChange({
                  ...loadBalancerQuota,
                  limit: event.target.valueAsNumber,
                })
              }
              data-cy="k8sNamespaceCreate-maxLoadBalancerInput"
              className="w-1/4"
            />
          </FormControl>
        </div>
      )}
    </FormSection>
  );
}
