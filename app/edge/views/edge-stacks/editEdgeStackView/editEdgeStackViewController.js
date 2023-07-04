import _ from 'lodash';
import { EditorType } from '@/react/edge/edge-stacks/types';
import { confirmWebEditorDiscard } from '@@/modals/confirm';
import { confirmStackUpdate } from '@/react/common/stacks/common/confirm-stack-update';
import { EnvironmentType } from '@/react/portainer/environments/types';
import { createWebhookId } from '@/portainer/helpers/webhookHelper';

export class EditEdgeStackViewController {
  /* @ngInject */
  constructor($async, $state, $window, EdgeGroupService, EdgeStackService, Notifications) {
    this.$async = $async;
    this.$state = $state;
    this.$window = $window;
    this.EdgeGroupService = EdgeGroupService;
    this.EdgeStackService = EdgeStackService;
    this.Notifications = Notifications;

    this.stack = null;
    this.edgeGroups = null;
    this.logEnvironmentIds = null;

    this.state = {
      actionInProgress: false,
      activeTab: 0,
      isStackDeployed: false,
    };

    this.formValues = {
      content: null,
    };

    this.deployStack = this.deployStack.bind(this);
    this.deployStackAsync = this.deployStackAsync.bind(this);
    this.onEditorChange = this.onEditorChange.bind(this);
    this.isEditorDirty = this.isEditorDirty.bind(this);
  }

  async $onInit() {
    return this.$async(async () => {
      const { stackId, tab } = this.$state.params;
      this.state.activeTab = tab ? parseInt(tab, 10) : 0;
      try {
        const [edgeGroups, model, file] = await Promise.all([
          this.EdgeGroupService.groups(),
          this.EdgeStackService.stack(stackId),
          this.EdgeStackService.stackFile(stackId).catch(() => ''),
        ]);

        this.edgeGroups = edgeGroups;
        this.stack = model;
        this.originalFileContent = file;
        this.formValues = {
          content: file,
        };

        const stackEdgeGroups = model.EdgeGroups.map((id) => this.edgeGroups.find((e) => e.Id === id));
        const endpointTypes = stackEdgeGroups.flatMap((group) => group.EndpointTypes);
        const initiallyContainsKubeEnv = endpointTypes.includes(EnvironmentType.EdgeAgentOnKubernetes);
        const isComposeStack = this.stack.DeploymentType === 0;

        this.allowKubeToSelectCompose = initiallyContainsKubeEnv && isComposeStack;
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve stack data');
      }

      this.$window.onbeforeunload = () => {
        if (this.isEditorDirty()) {
          return '';
        }
      };
    });
  }

  $onDestroy() {
    this.$window.onbeforeunload = undefined;
  }

  async uiCanExit() {
    if (this.isEditorDirty()) {
      return confirmWebEditorDiscard();
    }
  }

  onEditorChange(content) {
    this.formValues.content = content;
  }

  isEditorDirty() {
    return !this.state.isStackDeployed && this.formValues.content.replace(/(\r\n|\n|\r)/gm, '') !== this.originalFileContent.replace(/(\r\n|\n|\r)/gm, '');
  }

  deployStack(values) {
    return this.$async(async () => {
      let pullImage = false;
      if (this.formValues.DeploymentType == EditorType.Compose) {
        const defaultToggle = values.PrePullImage;
        const result = await confirmStackUpdate('Do you want to force an update of the stack?', defaultToggle);
        if (!result) {
          return;
        }
      }

      this.deployStackAsync(values, pullImage);
    });
  }

  async deployStackAsync(values, rePullImage) {
    this.state.actionInProgress = true;
    try {
      const updateVersion = !!(
        this.originalFileContent != values.content ||
        values.privateRegistryId !== this.stack.Registries[0] ||
        values.useManifestNamespaces !== this.stack.UseManifestNamespaces ||
        values.prePullImage !== this.stack.PrePullImage ||
        values.retryDeploy !== this.stack.RetryDeploy ||
        _.differenceWith(values.envVars, this.stack.Env, _.isEqual).length > 0 ||
        rePullImage
      );

      await this.EdgeStackService.updateStack(this.stack.Id, {
        stackFileContent: values.content,
        edgeGroups: values.edgeGroups,
        deploymentType: values.deploymentType,
        registries: [values.privateRegistryId],
        useManifestNamespaces: values.useManifestNamespaces,
        prePullImage: values.prePullImage,
        updateVersion,
        rePullImage,
        webhook: values.webhookEnabled ? this.stack.Webhook || createWebhookId() : '',
        envVars: values.envVars,
      });
      this.Notifications.success('Success', 'Stack successfully deployed');
      this.state.isStackDeployed = true;
      this.$state.go('edge.stacks');
    } catch (err) {
      this.Notifications.error('Deployment error', err, 'Unable to deploy stack');
    } finally {
      this.state.actionInProgress = false;
    }
  }
}
