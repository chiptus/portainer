import { Activity, Loader2 } from 'lucide-react';
import { useCurrentStateAndParams } from '@uirouter/react';
import { useEffect } from 'react';

import { notifyError } from '@/portainer/services/notifications';
import { useEnvironmentId } from '@/react/hooks/useEnvironmentId';

import { PageHeader } from '@@/PageHeader';
import { Widget, WidgetBody, WidgetTitle } from '@@/Widget';
import { Card } from '@@/Card';
import { Icon } from '@@/Icon';

import { useNodeQuery } from '../HomeView/nodes.service';
import { getInternalNodeIpAddress } from '../HomeView/NodesDatatable/utils';

import { useNodeStatusQuery } from './nodeStatus.service';

export function Microk8sNodeStatusView() {
  const {
    params: { nodeName },
  } = useCurrentStateAndParams();
  const environmentId = useEnvironmentId();

  const { data: node, ...nodeQuery } = useNodeQuery(environmentId, nodeName);
  const nodeIP = getInternalNodeIpAddress(node);
  const { data: nodeStatus, ...nodeStatusQuery } = useNodeStatusQuery(
    environmentId,
    nodeName,
    nodeIP
  );

  // wrap the error in useEffect to prevent showing the error multiple times
  useEffect(() => {
    if (!nodeName) {
      notifyError('Node name not found from the url path');
    }
  }, [nodeName]);

  return (
    <>
      <PageHeader
        title="Node status"
        breadcrumbs={[
          { label: 'Cluster information', link: 'kubernetes.cluster' },
          { label: 'Node status' },
        ]}
      />
      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetTitle title="Node status" icon={Activity} />
            <WidgetBody>
              <Card>
                {nodeQuery.isLoading ||
                  (nodeStatusQuery.isLoading && (
                    <div className="text-muted vertical-center text-sm">
                      <Icon icon={Loader2} className="animate-spin-slow" />
                      Loading MicroK8s node status
                    </div>
                  ))}
                {nodeStatus && (
                  <code className="whitespace-pre-wrap break-words bg-inherit p-0">
                    {nodeStatus}
                  </code>
                )}
              </Card>
            </WidgetBody>
          </Widget>
        </div>
      </div>
    </>
  );
}
