import _ from 'lodash-es';
import { getEnvironments } from '@/react/portainer/environments/environment.service';
import { EditorType } from '@/react/edge/edge-stacks/types';
// import { environmentTypeIcon } from '@/portainer/filters/filters';

export class EditEdgeStackViewController {
  /* @ngInject */
  constructor($async, $state, $window, ModalService, EdgeGroupService, EdgeStackService, Notifications) {
    this.$async = $async;
    this.$state = $state;
    this.$window = $window;
    this.ModalService = ModalService;
    this.EdgeGroupService = EdgeGroupService;
    this.EdgeStackService = EdgeStackService;
    this.Notifications = Notifications;

    this.stack = null;
    this.edgeGroups = null;
    this.logEnvironmentIds = null;

    this.state = {
      actionInProgress: false,
      activeTab: 0,
      isEditorDirty: false,
    };

    this.deployStack = this.deployStack.bind(this);
    this.deployStackAsync = this.deployStackAsync.bind(this);
    this.getPaginatedEndpoints = this.getPaginatedEndpoints.bind(this);
    this.getPaginatedEndpointsAsync = this.getPaginatedEndpointsAsync.bind(this);
  }

  async $onInit() {
    return this.$async(async () => {
      const { stackId, tab } = this.$state.params;
      this.state.activeTab = tab;
      try {
        const [edgeGroups, model, file] = await Promise.all([this.EdgeGroupService.groups(), this.EdgeStackService.stack(stackId), this.EdgeStackService.stackFile(stackId)]);

        this.edgeGroups = edgeGroups;
        this.stack = model;
        this.stackEndpointIds = this.filterStackEndpoints(model.EdgeGroups, edgeGroups);
        this.originalFileContent = file;
        this.formValues = {
          StackFileContent: file,
          EdgeGroups: this.stack.EdgeGroups,
          DeploymentType: this.stack.DeploymentType,
          Name: this.stack.Name,
          Registries: this.stack.Registries,
          UseManifestNamespaces: this.stack.UseManifestNamespaces,
          PrePullImage: this.stack.PrePullImage,
        };
        this.oldFileContent = this.formValues.StackFileContent;
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve stack data');
      }

      this.$window.onbeforeunload = () => {
        if (this.formValues.StackFileContent !== this.oldFileContent && this.state.isEditorDirty) {
          return '';
        }
      };
    });
  }

  $onDestroy() {
    this.state.isEditorDirty = false;
  }

  async uiCanExit() {
    if (this.formValues.StackFileContent.replace(/(\r\n|\n|\r)/gm, '') !== this.oldFileContent.replace(/(\r\n|\n|\r)/gm, '') && this.state.isEditorDirty) {
      return this.ModalService.confirmWebEditorDiscard();
    }
  }

  filterStackEndpoints(groupIds, groups) {
    return _.flatten(
      _.map(groupIds, (Id) => {
        const group = _.find(groups, { Id });
        return group.Endpoints;
      })
    );
  }

  deployStack() {
    if (this.formValues.DeploymentType == EditorType.Compose) {
      const that = this;
      const defaultToggle = this.formValues.PrePullImage;
      this.ModalService.confirmStackUpdate('Do you want to force an update of the stack?', defaultToggle, null, function (result) {
        if (result) {
          that.formValues.RePullImage = !!result[0];
          return that.$async(that.deployStackAsync);
        }
      });
    } else {
      return this.$async(this.deployStackAsync);
    }
  }

  async deployStackAsync() {
    this.state.actionInProgress = true;
    try {
      if (
        this.originalFileContent != this.formValues.StackFileContent ||
        this.formValues.Registries[0] !== this.stack.Registries[0] ||
        this.formValues.UseManifestNamespaces !== this.stack.UseManifestNamespaces ||
        this.formValues.PrePullImage !== this.stack.PrePullImage ||
        this.formValues.RePullImage
      ) {
        this.formValues.Version = this.stack.Version + 1;
      }
      await this.EdgeStackService.updateStack(this.stack.Id, this.formValues);
      this.Notifications.success('Success', 'Stack successfully deployed');
      this.state.isEditorDirty = false;
      this.$state.go('edge.stacks');
    } catch (err) {
      this.Notifications.error('Deployment error', err, 'Unable to deploy stack');
    } finally {
      this.state.actionInProgress = false;
    }
  }

  getPaginatedEndpoints(...args) {
    return this.$async(this.getPaginatedEndpointsAsync, ...args);
  }

  async getPaginatedEndpointsAsync(lastId, limit, search) {
    try {
      if (this.stackEndpointIds.length === 0) {
        return { endpoints: [], totalCount: 0 };
      }

      const query = { search, endpointIds: this.stackEndpointIds };
      const { value, totalCount } = await getEnvironments({ start: lastId, limit, query });

      return { endpoints: value, totalCount };
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve environment information');
    }
  }
}
