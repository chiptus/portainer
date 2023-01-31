import { PortainerEndpointTypes } from '@/portainer/models/endpoint/models';
import { getValidEditorTypes } from '@/react/edge/edge-stacks/utils';
import { EditorType } from '@/react/edge/edge-stacks/types';

export class EditEdgeStackFormController {
  /* @ngInject */
  constructor($async, Notifications, EdgeStackService, RegistryService, $window, $scope) {
    Object.assign(this, { $async, Notifications, EdgeStackService, RegistryService, $window, $scope });

    this.state = {
      endpointTypes: [],
      readOnlyCompose: false,
    };

    this.fileContents = {
      0: '',
      1: '',
      2: '',
    };

    this.formValues = {
      RegistryOptions: [],
    };

    this.isActive = false;

    this.EditorType = EditorType;

    this.matchRegistry = this.matchRegistry.bind(this);
    this.selectedRegistry = this.selectedRegistry.bind(this);
    this.dryrunFromFileContent = this.dryrunFromFileContent.bind(this);
    this.clearRegistries = this.clearRegistries.bind(this);
    this.getRegistriesOptions = this.getRegistriesOptions.bind(this);

    this.onChangeGroups = this.onChangeGroups.bind(this);
    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.onChangeComposeConfig = this.onChangeComposeConfig.bind(this);
    this.onChangeNomadHcl = this.onChangeNomadHcl.bind(this);
    this.onChangeKubeManifest = this.onChangeKubeManifest.bind(this);
    this.hasDockerEndpoint = this.hasDockerEndpoint.bind(this);
    this.hasNomadEndpoint = this.hasNomadEndpoint.bind(this);
    this.hasKubeEndpoint = this.hasKubeEndpoint.bind(this);
    this.onChangeDeploymentType = this.onChangeDeploymentType.bind(this);
    this.removeLineBreaks = this.removeLineBreaks.bind(this);
    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.onChangeUseManifestNamespaces = this.onChangeUseManifestNamespaces.bind(this);
    this.onChangePrePullImage = this.onChangePrePullImage.bind(this);
    this.selectValidDeploymentType = this.selectValidDeploymentType.bind(this);
  }

  checkRegistries(registries) {
    return registries.every((value) => value === registries[0]);
  }

  onChangeUseManifestNamespaces(value) {
    this.$scope.$evalAsync(() => {
      this.model.UseManifestNamespaces = value;
    });
  }

  selectedRegistry(e) {
    return this.$async(async () => {
      const selectedRegistry = e;
      this.registryID = selectedRegistry.Id;
      this.model.Registries[0] = this.registryID;
    });
  }

  clearRegistries() {
    this.model.Registries = [];
    this.registryID = '';
  }

  matchRegistry() {
    return this.$async(async () => {
      this.registryID = '';
      this.errorMessage = '';
      this.dryrun = true;
      let response = '';

      this.dryrunName = this.model.Name + '-' + 'dryrun';

      response = await this.dryrunFromFileContent(this.dryrunName, this.dryrun);

      try {
        if (response.Registries.length !== 0) {
          const validRegistry = this.checkRegistries(response.Registries);
          if (validRegistry) {
            this.registryID = response.Registries[0];
            this.model.Registries[0] = this.registryID;
          } else {
            this.registryID = '';
            this.errorMessage = ' Images need to be from a single registry, please edit and reload';
          }
        } else {
          this.registryID = '';
        }
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve registries');
      } finally {
        this.dryrun = false;
      }
    });
  }

  dryrunFromFileContent(name, dryrun) {
    const { StackFileContent, Groups, DeploymentType } = this.formValues;
    return this.EdgeStackService.createStackFromFileContent(
      {
        name,
        StackFileContent,
        EdgeGroups: Groups,
        DeploymentType,
      },
      dryrun
    );
  }

  hasKubeEndpoint() {
    return this.state.endpointTypes.includes(PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment);
  }

  hasDockerEndpoint() {
    return this.state.endpointTypes.includes(PortainerEndpointTypes.EdgeAgentOnDockerEnvironment);
  }

  hasNomadEndpoint() {
    return this.state.endpointTypes.includes(PortainerEndpointTypes.EdgeAgentOnNomadEnvironment);
  }

  onChangeGroups(groups) {
    return this.$scope.$evalAsync(() => {
      this.model.EdgeGroups = groups;
      this.setEnvironmentTypesInSelection(groups);
      this.selectValidDeploymentType();
      this.state.readOnlyCompose = this.hasKubeEndpoint();
    });
  }

  isFormValid() {
    return this.model.EdgeGroups.length && this.model.StackFileContent && this.validateEndpointsForDeployment();
  }

  setEnvironmentTypesInSelection(groups) {
    const edgeGroups = groups.map((id) => this.edgeGroups.find((e) => e.Id === id));
    this.state.endpointTypes = edgeGroups.flatMap((group) => group.EndpointTypes);
  }

  selectValidDeploymentType() {
    const validTypes = getValidEditorTypes(this.state.endpointTypes, this.allowKubeToSelectCompose);

    if (!validTypes.includes(this.model.DeploymentType)) {
      this.onChangeDeploymentType(validTypes[0]);
    }
  }

  removeLineBreaks(value) {
    return value.replace(/(\r\n|\n|\r)/gm, '');
  }

  onChangeFileContent(type, value) {
    const oldValue = this.fileContents[type];
    if (this.removeLineBreaks(oldValue) !== this.removeLineBreaks(value)) {
      this.model.StackFileContent = value;
      this.fileContents[type] = value;
      this.isEditorDirty = true;
    }
  }

  onChangeNomadHcl(value) {
    this.onChangeFileContent(2, value);
  }

  onChangeKubeManifest(value) {
    this.onChangeFileContent(1, value);
    this.formValues.StackFileContent = value;
  }

  onChangeComposeConfig(value) {
    this.onChangeFileContent(0, value);
    this.formValues.StackFileContent = value;
  }

  onChangeDeploymentType(deploymentType) {
    this.$scope.$evalAsync(() => {
      this.model.DeploymentType = deploymentType;
      this.model.StackFileContent = this.fileContents[deploymentType];
    });
  }

  validateEndpointsForDeployment() {
    return this.model.DeploymentType == 0 || !this.hasDockerEndpoint();
  }

  getRegistriesOptions() {
    return this.$async(async () => {
      this.formValues.RegistryOptions = await this.RegistryService.registries();
    });
  }

  onChangePrePullImage(value) {
    return this.$scope.$evalAsync(() => {
      this.model.PrePullImage = value;
    });
  }

  $onInit() {
    this.fileContents[this.model.DeploymentType] = this.model.StackFileContent;
    this.setEnvironmentTypesInSelection(this.model.EdgeGroups);
    this.getRegistriesOptions();

    this.formValues.StackFileContent = this.model.StackFileContent;
    this.formValues.DeploymentType = this.model.DeploymentType;
    this.formValues.Groups = this.model.EdgeGroups;

    // allow kube to view compose if it's an existing kube compose stack
    const initiallyContainsKubeEnv = this.hasKubeEndpoint();
    const isComposeStack = this.model.DeploymentType === 0;
    this.allowKubeToSelectCompose = initiallyContainsKubeEnv && isComposeStack;
    this.state.readOnlyCompose = this.allowKubeToSelectCompose;
    this.selectValidDeploymentType();

    if (this.model.Registries !== null && this.model.Registries.length !== 0) {
      this.isActive = true;
    }
  }
}
