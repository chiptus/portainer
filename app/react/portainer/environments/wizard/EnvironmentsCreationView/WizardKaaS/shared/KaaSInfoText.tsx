import { TextTip } from '@@/Tip/TextTip';

export function KaaSInfoText() {
  return (
    <TextTip color="blue">
      This will allow you to create a Kubernetes environment (cluster) using a
      cloud provider&apos;s Kubernetes as a Service (KaaS) offering, and will
      then deploy the Portainer agent to it.
    </TextTip>
  );
}
