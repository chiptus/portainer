import { useCurrentStateAndParams } from '@uirouter/react';
import { Layers } from 'lucide-react';

import { useParamState } from '@/react/hooks/useParamState';

import { NavTabs } from '@@/NavTabs';
import { PageHeader } from '@@/PageHeader';
import { Widget } from '@@/Widget';
import { Icon } from '@@/Icon';

import { useEdgeStack } from '../queries/useEdgeStack';

import { EditEdgeStackForm } from './EditEdgeStackForm/EditEdgeStackForm';

export function ItemView() {
  const {
    params: { stackId },
  } = useCurrentStateAndParams();
  const [tab, setTab] = useParamState('tab', (param) =>
    param ? parseInt(param, 10) : 0
  );
  const stackQuery = useEdgeStack(stackId);

  if (!stackQuery.data) {
    return null;
  }

  const stack = stackQuery.data;

  return (
    <>
      <PageHeader
        title="Edit Edge stack"
        breadcrumbs={[
          { label: 'Edge Stacks', link: 'edge.stacks' },
          stack.Name,
        ]}
        reload
      />
      <div className="mx-4">
        <Widget>
          <Widget.Body>
            <NavTabs<number>
              type="pills"
              justified
              selectedId={tab}
              onSelect={setTab}
              options={[
                {
                  id: 0,
                  label: (
                    <div className="vertical-center">
                      <Icon icon={Layers} /> Stack
                    </div>
                  ),
                  children: <EditEdgeStackForm />,
                },
              ]}
            />
          </Widget.Body>
        </Widget>
      </div>
    </>
  );
}
