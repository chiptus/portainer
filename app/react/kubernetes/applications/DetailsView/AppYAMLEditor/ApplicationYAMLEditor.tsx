import { useCurrentStateAndParams } from '@uirouter/react';

import { InlineLoader } from '@@/InlineLoader';
import { Widget } from '@@/Widget/Widget';
import { WidgetBody } from '@@/Widget';

import { YAMLInspector } from '../../../components/YAMLInspector';
import { isSystemNamespace } from '../../../namespaces/utils';

import { useApplicationYAML } from './useApplicationYAML';

// the yaml currently has the yaml from the app, related services and horizontal pod autoscalers
// TODO: this could be extended to include other related resources like ingresses, etc.
export function ApplicationYAMLEditor() {
  const {
    params: { namespace },
  } = useCurrentStateAndParams();
  const isInSystemNamespace = namespace ? isSystemNamespace(namespace) : false;
  const { fullApplicationYaml, isApplicationYAMLLoading } =
    useApplicationYAML();

  if (isApplicationYAMLLoading) {
    return (
      <div className="row">
        <div className="col-sm-12">
          <Widget>
            <WidgetBody>
              <InlineLoader>Loading application YAML...</InlineLoader>
            </WidgetBody>
          </Widget>
        </div>
      </div>
    );
  }

  return (
    <div className="row">
      <div className="col-sm-12">
        <Widget>
          <WidgetBody>
            <YAMLInspector
              identifier="application-yaml"
              data={fullApplicationYaml}
              authorised
              system={isInSystemNamespace}
              hideMessage
            />
          </WidgetBody>
        </Widget>
      </div>
    </div>
  );
}
