import angular from 'angular';

import { AccessControlFormData } from '@/portainer/components/accessControlForm/porAccessControlFormModel';
import { STACK_NAME_VALIDATION_REGEX } from '@/react/constants';
import { RepositoryMechanismTypes } from '@/kubernetes/models/deploy';
import { isTemplateVariablesEnabled, renderTemplate } from '@/react/portainer/custom-templates/components/utils';
import { editor, upload, git, customTemplate } from '@@/BoxSelector/common-options/build-methods';
import { confirmWebEditorDiscard } from '@@/modals/confirm';
import { parseAutoUpdateResponse, transformAutoUpdateViewModel } from '@/react/portainer/gitops/AutoUpdateFieldset/utils';
import { baseStackWebhookUrl, createWebhookId } from '@/portainer/helpers/webhookHelper';
import { getVariablesFieldDefaultValues } from '@/react/portainer/custom-templates/components/CustomTemplatesVariablesField';

angular
  .module('portainer.app')
  .controller(
    'CreateStackController',
    function (
      $scope,
      $state,
      $async,
      $window,
      StackService,
      Authentication,
      Notifications,
      FormValidator,
      ResourceControlService,
      FormHelper,
      StackHelper,
      ContainerHelper,
      ContainerService,
      UserService,
      endpoint
    ) {
      $scope.onChangeTemplateId = onChangeTemplateId;
      $scope.onChangeTemplateVariables = onChangeTemplateVariables;
      $scope.isTemplateVariablesEnabled = isTemplateVariablesEnabled;
      $scope.buildAnalyticsProperties = buildAnalyticsProperties;
      $scope.buildMethods = [editor, upload, git, customTemplate];
      $scope.STACK_NAME_VALIDATION_REGEX = STACK_NAME_VALIDATION_REGEX;
      $scope.isAdmin = Authentication.isAdmin();
      $scope.isDockerStandalone = $scope.applicationState.endpoint.mode.provider === 'DOCKER_STANDALONE';

      $scope.formValues = {
        Name: '',
        StackFileContent: $state.params.yaml || '',
        StackFile: null,
        RepositoryURL: '',
        RepositoryURLValid: false,
        RepositoryReferenceName: 'refs/heads/main',
        RepositoryAuthentication: false,
        RepositoryUsername: '',
        RepositoryPassword: '',
        SaveCredential: true,
        RepositoryGitCredentialID: 0,
        NewCredentialName: '',
        Env: [],
        SupportRelativePath: false,
        FilesystemPath: '',
        AdditionalFiles: [],
        ComposeFilePathInRepository: 'docker-compose.yml',
        AccessControlData: new AccessControlFormData(),
        EnableWebhook: false,
        Variables: [],
        AutoUpdate: parseAutoUpdateResponse(),
        TLSSkipVerify: false,
      };

      $scope.state = {
        Method: 'editor',
        formValidationError: '',
        actionInProgress: false,
        StackType: null,
        editorYamlValidationError: '',
        uploadYamlValidationError: '',
        isEditorDirty: false,
        selectedTemplate: null,
        selectedTemplateId: null,
        baseWebhookUrl: baseStackWebhookUrl(),
        webhookId: createWebhookId(),
        selectedTemplateFailedUpdate: false,
        templateLoadFailed: false,
        isEditorReadOnly: false,
        versions: [1],
      };

      $window.onbeforeunload = () => {
        if ($scope.state.Method === 'editor' && $scope.formValues.StackFileContent && $scope.state.isEditorDirty) {
          return '';
        }
      };

      $scope.$on('$destroy', function () {
        $scope.state.isEditorDirty = false;
      });

      $scope.onChangeFormValues = onChangeFormValues;
      $scope.onBuildMethodChange = onBuildMethodChange;

      function onBuildMethodChange(value) {
        $scope.$evalAsync(() => {
          $scope.state.Method = value;
        });
      }

      $scope.onEnableWebhookChange = function (enable) {
        $scope.$evalAsync(() => {
          $scope.formValues.EnableWebhook = enable;
        });
      };

      $scope.onEnableRelativePathsChange = function (enable) {
        $scope.$evalAsync(() => {
          $scope.formValues.SupportRelativePath = enable;
        });
      };

      $scope.hasAuthorizations = function (authorizations) {
        return $scope.isAdmin || Authentication.hasAuthorizations(authorizations);
      };

      function buildAnalyticsProperties() {
        const metadata = { type: methodLabel($scope.state.Method) };

        if ($scope.state.Method === 'repository') {
          metadata.automaticUpdates = 'off';
          if ($scope.formValues.RepositoryAutomaticUpdates) {
            metadata.automaticUpdates = autoSyncLabel($scope.formValues.RepositoryMechanism);
          }
          metadata.auth = $scope.formValues.RepositoryAuthentication;
        }

        if ($scope.state.Method === 'template') {
          metadata.templateName = $scope.state.selectedTemplate.Title;
        }

        return { metadata };

        function methodLabel(method) {
          switch (method) {
            case 'editor':
              return 'web-editor';
            case 'repository':
              return 'git';
            case 'upload':
              return 'file-upload';
            case 'template':
              return 'custom-template';
          }
        }

        function autoSyncLabel(type) {
          switch (type) {
            case RepositoryMechanismTypes.INTERVAL:
              return 'polling';
            case RepositoryMechanismTypes.WEBHOOK:
              return 'webhook';
          }
          return 'off';
        }
      }

      function validateForm(accessControlData, isAdmin) {
        $scope.state.formValidationError = '';
        var error = '';
        error = FormValidator.validateAccessControl(accessControlData, isAdmin);

        if (error) {
          $scope.state.formValidationError = error;
          return false;
        }
        return true;
      }

      function createSwarmStack(name, method) {
        var env = FormHelper.removeInvalidEnvVars($scope.formValues.Env);
        const endpointId = +$state.params.endpointId;
        const webhook = $scope.formValues.EnableWebhook ? $scope.state.webhookId : '';
        if (method === 'template' || method === 'editor') {
          var stackFileContent = $scope.formValues.StackFileContent;
          return StackService.createSwarmStackFromFileContent(name, stackFileContent, env, endpointId, webhook);
        }

        if (method === 'upload') {
          var stackFile = $scope.formValues.StackFile;
          return StackService.createSwarmStackFromFileUpload(name, stackFile, env, endpointId, webhook);
        }

        if (method === 'repository') {
          var repositoryOptions = {
            AdditionalFiles: $scope.formValues.AdditionalFiles,
            RepositoryURL: $scope.formValues.RepositoryURL,
            RepositoryReferenceName: $scope.formValues.RepositoryReferenceName,
            ComposeFilePathInRepository: $scope.formValues.ComposeFilePathInRepository,
            RepositoryAuthentication: $scope.formValues.RepositoryAuthentication,
            RepositoryUsername: $scope.formValues.RepositoryGitCredentialID === 0 ? $scope.formValues.RepositoryUsername : '',
            RepositoryPassword: $scope.formValues.RepositoryGitCredentialID === 0 ? $scope.formValues.RepositoryPassword : '',
            RepositoryGitCredentialID: $scope.formValues.RepositoryGitCredentialID,
            ForcePullImage: $scope.formValues.ForcePullImage,
            SupportRelativePath: $scope.formValues.SupportRelativePath,
            FilesystemPath: $scope.formValues.FilesystemPath,
            AutoUpdate: transformAutoUpdateViewModel($scope.formValues.AutoUpdate, $scope.state.webhookId),
            TLSSkipVerify: $scope.formValues.TLSSkipVerify,
          };

          return StackService.createSwarmStackFromGitRepository(name, repositoryOptions, env, endpointId);
        }
      }

      function createComposeStack(name, method) {
        var env = FormHelper.removeInvalidEnvVars($scope.formValues.Env);
        const endpointId = +$state.params.endpointId;
        const webhook = $scope.formValues.EnableWebhook ? $scope.state.webhookId : '';

        if (method === 'editor' || method === 'template') {
          var stackFileContent = $scope.formValues.StackFileContent;
          return StackService.createComposeStackFromFileContent(name, stackFileContent, env, endpointId, webhook);
        } else if (method === 'upload') {
          var stackFile = $scope.formValues.StackFile;
          return StackService.createComposeStackFromFileUpload(name, stackFile, env, endpointId, webhook);
        } else if (method === 'repository') {
          var repositoryOptions = {
            AdditionalFiles: $scope.formValues.AdditionalFiles,
            RepositoryURL: $scope.formValues.RepositoryURL,
            RepositoryReferenceName: $scope.formValues.RepositoryReferenceName,
            ComposeFilePathInRepository: $scope.formValues.ComposeFilePathInRepository,
            RepositoryAuthentication: $scope.formValues.RepositoryAuthentication,
            RepositoryUsername: $scope.formValues.RepositoryGitCredentialID === 0 ? $scope.formValues.RepositoryUsername : '',
            RepositoryPassword: $scope.formValues.RepositoryGitCredentialID === 0 ? $scope.formValues.RepositoryPassword : '',
            RepositoryGitCredentialID: $scope.formValues.RepositoryGitCredentialID,
            ForcePullImage: $scope.formValues.ForcePullImage,
            SupportRelativePath: $scope.formValues.SupportRelativePath,
            FilesystemPath: $scope.formValues.FilesystemPath,
            AutoUpdate: transformAutoUpdateViewModel($scope.formValues.AutoUpdate, $scope.state.webhookId),
            TLSSkipVerify: $scope.formValues.TLSSkipVerify,
          };

          return StackService.createComposeStackFromGitRepository(name, repositoryOptions, env, endpointId);
        }
      }

      $scope.handleEnvVarChange = handleEnvVarChange;
      function handleEnvVarChange(value) {
        $scope.formValues.Env = value;
      }

      $scope.deployStack = async function () {
        var name = $scope.formValues.Name;
        var method = $scope.state.Method;

        var accessControlData = $scope.formValues.AccessControlData;
        var userDetails = Authentication.getUserDetails();
        var isAdmin = Authentication.isAdmin();

        if (method === 'editor' && $scope.formValues.StackFileContent === '') {
          $scope.state.formValidationError = 'Stack file content must not be empty';
          return;
        }

        if (!validateForm(accessControlData, isAdmin)) {
          return;
        }

        var type = $scope.state.StackType;
        var action = createSwarmStack;

        if (type === 2) {
          action = createComposeStack;
        }
        $scope.state.actionInProgress = true;

        // save git credential
        if ($scope.formValues.SaveCredential && $scope.formValues.NewCredentialName) {
          await UserService.saveGitCredential(userDetails.ID, $scope.formValues.NewCredentialName, $scope.formValues.RepositoryUsername, $scope.formValues.RepositoryPassword).then(
            function success(data) {
              $scope.formValues.RepositoryGitCredentialID = data.gitCredential.id;
            }
          );
        }

        action(name, method)
          .then(function success(data) {
            if (data.data) {
              data = data.data;
            }
            const userId = userDetails.ID;
            const resourceControl = data.ResourceControl;
            return ResourceControlService.applyResourceControl(userId, accessControlData, resourceControl);
          })
          .then(function success() {
            Notifications.success('Success', 'Stack successfully deployed');
            $scope.state.isEditorDirty = false;
            $state.go('docker.stacks');
          })
          .catch(function error(err) {
            Notifications.error('Deployment error', err, 'Unable to deploy stack');
            if ($scope.formValues.SaveCredential && $scope.formValues.NewCredentialName && $scope.formValues.RepositoryGitCredentialID) {
              UserService.deleteGitCredential(userDetails.ID, $scope.formValues.RepositoryGitCredentialID);
            }
          })
          .finally(function final() {
            $scope.state.actionInProgress = false;
          });
      };

      $scope.onChangeFileContent = onChangeFileContent;
      function onChangeFileContent(value) {
        $scope.formValues.StackFileContent = value;
        $scope.state.editorYamlValidationError = StackHelper.validateYAML($scope.formValues.StackFileContent, $scope.containerNames);
        $scope.state.isEditorDirty = true;
      }

      async function onFileLoadAsync(event) {
        $scope.state.uploadYamlValidationError = StackHelper.validateYAML(event.target.result, $scope.containerNames);
      }

      function onFileLoad(event) {
        return $async(onFileLoadAsync, event);
      }

      $scope.uploadFile = function (file) {
        $scope.formValues.StackFile = file;

        if (file) {
          const temporaryFileReader = new FileReader();
          temporaryFileReader.fileName = file.name;
          temporaryFileReader.onload = onFileLoad;
          temporaryFileReader.readAsText(file);
        }
      };

      function onChangeTemplateId(templateId, template) {
        return $async(async () => {
          if (!template || ($scope.state.selectedTemplateId === templateId && $scope.state.selectedTemplate === template)) {
            return;
          }

          try {
            $scope.state.selectedTemplateId = templateId;
            $scope.state.selectedTemplate = template;

            try {
              $scope.state.templateContent = await this.CustomTemplateService.customTemplateFile(templateId, template.GitConfig !== null);
              onChangeFileContent($scope.state.templateContent);

              $scope.state.isEditorReadOnly = true;
            } catch (err) {
              $scope.state.templateLoadFailed = true;
              throw err;
            }

            if (template.Variables && template.Variables.length > 0) {
              const variables = getVariablesFieldDefaultValues(template.Variables);
              onChangeTemplateVariables(variables);
            }
          } catch (err) {
            Notifications.error('Failure', err, 'Unable to retrieve Custom Template file');
          }
        });
      }

      function onChangeTemplateVariables(value) {
        onChangeFormValues({ Variables: value });

        if (!$scope.isTemplateVariablesEnabled) {
          return;
        }
        const rendered = renderTemplate($scope.state.templateContent, $scope.formValues.Variables, $scope.state.selectedTemplate.Variables);
        onChangeFormValues({ StackFileContent: rendered });
      }

      async function initView() {
        var endpointMode = $scope.applicationState.endpoint.mode;
        $scope.state.StackType = 2;
        $scope.isDockerStandalone = endpointMode.provider === 'DOCKER_STANDALONE';
        if (endpointMode.provider === 'DOCKER_SWARM_MODE' && endpointMode.role === 'MANAGER') {
          $scope.state.StackType = 1;
        }

        $scope.composeSyntaxMaxVersion = endpoint.ComposeSyntaxMaxVersion;
        try {
          const containers = await ContainerService.containers(endpoint.Id, true);
          $scope.containerNames = ContainerHelper.getContainerNames(containers);
        } catch (err) {
          Notifications.error('Failure', err, 'Unable to retrieve Containers');
        }
      }

      this.uiCanExit = async function () {
        if ($scope.state.Method === 'editor' && $scope.formValues.StackFileContent && $scope.state.isEditorDirty) {
          return confirmWebEditorDiscard();
        }
      };

      initView();

      function onChangeFormValues(newValues) {
        return $async(async () => {
          $scope.formValues = {
            ...$scope.formValues,
            ...newValues,
          };
        });
      }
    }
  );
