import { PortainerEndpointCreationTypes, PortainerEndpointTypes } from 'Portainer/models/endpoint/models';
import { getAgentShortVersion } from 'Portainer/views/endpoints/helpers';
import { baseHref } from '@/portainer/helpers/pathHelper';
import { EDGE_ASYNC_INTERVAL_USE_DEFAULT } from '@/edge/components/EdgeAsyncIntervalsForm';
import { EndpointSecurityFormData } from '../../../components/endpointSecurity/porEndpointSecurityModel';

angular
  .module('portainer.app')
  .controller(
    'CreateEndpointController',
    function CreateEndpointController(
      $async,
      $analytics,
      $q,
      $scope,
      $state,
      $filter,
      clipboard,
      EndpointService,
      GroupService,
      SettingsService,
      Notifications,
      Authentication,
      StateManager
    ) {
      $scope.onChangeCheckInInterval = onChangeCheckInInterval;
      $scope.setFieldValue = setFieldValue;
      $scope.onChangeEdgeSettings = onChangeEdgeSettings;

      $scope.state = {
        EnvironmentType: $state.params.isEdgeDevice ? 'edge_agent' : 'agent',
        PlatformType: 'linux',
        actionInProgress: false,
        deploymentTab: 0,
        allowCreateTag: Authentication.isAdmin(),
        isEdgeDevice: $state.params.isEdgeDevice,
        edgeAsyncMode: false,
        kubeconfigError: true,
        kubeconfigFile: '',
      };

      $scope.agentVersion = StateManager.getState().application.version;
      $scope.agentShortVersion = getAgentShortVersion($scope.agentVersion);
      $scope.agentSecret = '';

      $scope.formValues = {
        Name: '',
        URL: '',
        PublicURL: '',
        GroupId: 1,
        SecurityFormData: new EndpointSecurityFormData(),
        AzureApplicationId: '',
        AzureTenantId: '',
        AzureAuthenticationKey: '',
        TagIds: [],
        CheckinInterval: 0,
        KubeConfig: '',
        Edge: {
          PingInterval: EDGE_ASYNC_INTERVAL_USE_DEFAULT,
          SnapshotInterval: EDGE_ASYNC_INTERVAL_USE_DEFAULT,
          CommandInterval: EDGE_ASYNC_INTERVAL_USE_DEFAULT,
        },
      };

      $scope.deployCommands = {
        kubeLoadBalancer: `curl -L https://downloads.portainer.io/ee${$scope.agentShortVersion}/portainer-agent-k8s-lb.yaml -o portainer-agent-k8s.yaml; kubectl apply -f portainer-agent-k8s.yaml`,
        kubeNodePort: `curl -L https://downloads.portainer.io/ee${$scope.agentShortVersion}/portainer-agent-k8s-nodeport.yaml -o portainer-agent-k8s.yaml; kubectl apply -f portainer-agent-k8s.yaml`,
        agentLinux: agentLinuxSwarmCommand,
        agentWindows: agentWindowsSwarmCommand,
      };

      $scope.copyAgentCommand = function () {
        if ($scope.state.deploymentTab === 2 && $scope.state.PlatformType === 'linux') {
          clipboard.copyText($scope.deployCommands.agentLinux($scope.agentVersion, $scope.agentSecret));
        } else if ($scope.state.deploymentTab === 2 && $scope.state.PlatformType === 'windows') {
          clipboard.copyText($scope.deployCommands.agentWindows($scope.agentVersion, $scope.agentSecret));
        } else if ($scope.state.deploymentTab === 1) {
          clipboard.copyText($scope.deployCommands.kubeNodePort);
        } else {
          clipboard.copyText($scope.deployCommands.kubeLoadBalancer);
        }
        $('#copyNotification').show().fadeOut(2500);
      };

      $scope.setDefaultPortainerInstanceURL = function () {
        const baseHREF = baseHref();
        $scope.formValues.URL = window.location.origin + (baseHREF !== '/' ? baseHREF : '');
      };

      $scope.resetEndpointURL = function () {
        $scope.formValues.URL = '';
      };

      $scope.onChangeTags = function onChangeTags(value) {
        return $scope.$evalAsync(() => {
          $scope.formValues.TagIds = value;
        });
      };

      function onChangeCheckInInterval(value) {
        setFieldValue('EdgeCheckinInterval', value);
      }

      function onChangeEdgeSettings(value) {
        setFieldValue('Edge', value);
      }

      function setFieldValue(name, value) {
        return $scope.$evalAsync(() => {
          $scope.formValues = {
            ...$scope.formValues,
            [name]: value,
          };
        });
      }

      $scope.onCreateKaaSEnvironment = function onCreateKaaSEnvironment() {
        $state.go('portainer.endpoints');
      };

      async function readFileContent(file) {
        return new Promise((resolve, reject) => {
          const fileReader = new FileReader();
          fileReader.onload = (e) => {
            if (e.target == null || e.target.result == null) {
              resolve('');
              return;
            }
            const base64 = e.target.result.toString();
            const index = base64.indexOf('base64,');
            const cert = base64.substring(index + 7, base64.length);
            resolve(cert);
          };
          fileReader.onerror = () => {
            reject(new Error('error reading provisioning certificate file'));
          };
          fileReader.readAsDataURL(file);
        });
      }

      $scope.handleKubeconfigUpload = async function handleKubeconfigUpload(file) {
        $scope.state.kubeconfigError = false;
        $scope.state.kubeconfigFile = file;
        const fileContent = await readFileContent(file);
        $scope.formValues = {
          ...$scope.formValues,
          KubeConfig: fileContent,
        };
        $scope.$apply();
      };

      $scope.addDockerEndpoint = function () {
        var name = $scope.formValues.Name;
        var URL = $filter('stripprotocol')($scope.formValues.URL);
        var publicURL = $scope.formValues.PublicURL;
        var groupId = $scope.formValues.GroupId;
        var tagIds = $scope.formValues.TagIds;

        if ($scope.formValues.ConnectSocket) {
          URL = $scope.formValues.SocketPath;
          $scope.state.actionInProgress = true;
          EndpointService.createLocalEndpoint(name, URL, publicURL, groupId, tagIds)
            .then(function success() {
              Notifications.success('Environment created', name);
              $state.go('portainer.endpoints', {}, { reload: true });
            })
            .catch(function error(err) {
              Notifications.error('Failure', err, 'Unable to create environment');
            })
            .finally(function final() {
              $scope.state.actionInProgress = false;
            });
        } else {
          if (publicURL === '') {
            publicURL = URL.split(':')[0];
          }
          var securityData = $scope.formValues.SecurityFormData;
          var TLS = securityData.TLS;
          var TLSMode = securityData.TLSMode;
          var TLSSkipVerify = TLS && (TLSMode === 'tls_client_noca' || TLSMode === 'tls_only');
          var TLSSkipClientVerify = TLS && (TLSMode === 'tls_ca' || TLSMode === 'tls_only');
          var TLSCAFile = TLSSkipVerify ? null : securityData.TLSCACert;
          var TLSCertFile = TLSSkipClientVerify ? null : securityData.TLSCert;
          var TLSKeyFile = TLSSkipClientVerify ? null : securityData.TLSKey;

          addEndpoint(
            name,
            PortainerEndpointCreationTypes.LocalDockerEnvironment,
            URL,
            publicURL,
            groupId,
            tagIds,
            TLS,
            TLSSkipVerify,
            TLSSkipClientVerify,
            TLSCAFile,
            TLSCertFile,
            TLSKeyFile
          );
        }
      };

      $scope.addKubernetesEndpoint = function () {
        var name = $scope.formValues.Name;
        var tagIds = $scope.formValues.TagIds;
        $scope.state.actionInProgress = true;
        EndpointService.createLocalKubernetesEndpoint(name, tagIds)
          .then(function success(result) {
            Notifications.success('Environment created', name);
            $state.go('portainer.k8sendpoint.kubernetesConfig', { id: result.Id });
          })
          .catch(function error(err) {
            Notifications.error('Failure', err, 'Unable to create environment');
          })
          .finally(function final() {
            $scope.state.actionInProgress = false;
          });
      };

      $scope.addAgentEndpoint = addAgentEndpoint;
      async function addAgentEndpoint() {
        return $async(async () => {
          const name = $scope.formValues.Name;
          const URL = $scope.formValues.URL;
          const publicURL = $scope.formValues.PublicURL === '' ? URL.split(':')[0] : $scope.formValues.PublicURL;
          const groupId = $scope.formValues.GroupId;
          const tagIds = $scope.formValues.TagIds;

          const endpoint = await addEndpoint(name, PortainerEndpointCreationTypes.AgentEnvironment, URL, publicURL, groupId, tagIds, true, true, true, null, null, null);
          $analytics.eventTrack('portainer-endpoint-creation', { category: 'portainer', metadata: { type: 'agent', platform: platformLabel(endpoint.Type) } });
        });

        function platformLabel(type) {
          switch (type) {
            case PortainerEndpointTypes.DockerEnvironment:
            case PortainerEndpointTypes.AgentOnDockerEnvironment:
            case PortainerEndpointTypes.EdgeAgentOnDockerEnvironment:
              return 'docker';
            case PortainerEndpointTypes.KubernetesLocalEnvironment:
            case PortainerEndpointTypes.AgentOnKubernetesEnvironment:
            case PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment:
              return 'kubernetes';
          }
        }
      }

      $scope.addEdgeAgentEndpoint = function () {
        var name = $scope.formValues.Name;
        var groupId = $scope.formValues.GroupId;
        var tagIds = $scope.formValues.TagIds;
        var URL = $scope.formValues.URL;

        addEndpoint(
          name,
          PortainerEndpointCreationTypes.EdgeAgentEnvironment,
          URL,
          '',
          groupId,
          tagIds,
          false,
          false,
          false,
          null,
          null,
          null,
          $scope.formValues.CheckinInterval,
          $scope.formValues.Edge.PingInterval,
          $scope.formValues.Edge.SnapshotInterval,
          $scope.formValues.Edge.CommandInterval
        );
      };

      $scope.addAzureEndpoint = function () {
        var name = $scope.formValues.Name;
        var applicationId = $scope.formValues.AzureApplicationId;
        var tenantId = $scope.formValues.AzureTenantId;
        var authenticationKey = $scope.formValues.AzureAuthenticationKey;
        var groupId = $scope.formValues.GroupId;
        var tagIds = $scope.formValues.TagIds;

        createAzureEndpoint(name, applicationId, tenantId, authenticationKey, groupId, tagIds);
      };

      function createAzureEndpoint(name, applicationId, tenantId, authenticationKey, groupId, tagIds) {
        $scope.state.actionInProgress = true;
        EndpointService.createAzureEndpoint(name, applicationId, tenantId, authenticationKey, groupId, tagIds)
          .then(function success() {
            Notifications.success('Environment created', name);
            $state.go('portainer.endpoints', {}, { reload: true });
          })
          .catch(function error(err) {
            Notifications.error('Failure', err, 'Unable to create environment');
          })
          .finally(function final() {
            $scope.state.actionInProgress = false;
          });
      }

      $scope.addKubeconfigEndpoint = function () {
        var name = $scope.formValues.Name;
        var kubeConfig = $scope.formValues.KubeConfig;
        var groupId = $scope.formValues.GroupId;
        var tagIds = $scope.formValues.TagIds;

        createKubeconfigEndpoint(name, kubeConfig, groupId, tagIds);
      };

      function createKubeconfigEndpoint(name, kubeConfig, groupId, tagIds) {
        $scope.state.actionInProgress = true;
        EndpointService.createKubeConfigEndpoint(name, kubeConfig, groupId, tagIds)
          .then(function success() {
            Notifications.success('Kubeconfig import started', name);
            $state.go('portainer.endpoints', {}, { reload: true });
          })
          .catch(function error(err) {
            Notifications.error('Failure', err, 'Unable to create environment');
          })
          .finally(function final() {
            $scope.state.actionInProgress = false;
          });
      }

      async function addEndpoint(
        name,
        creationType,
        URL,
        PublicURL,
        groupId,
        tagIds,
        TLS,
        TLSSkipVerify,
        TLSSkipClientVerify,
        TLSCAFile,
        TLSCertFile,
        TLSKeyFile,
        CheckinInterval,
        EdgePingInterval,
        EdgeSnapshotInterval,
        EdgeCommandInterval
      ) {
        return $async(async () => {
          $scope.state.actionInProgress = true;
          try {
            const endpoint = await EndpointService.createRemoteEndpoint(
              name,
              creationType,
              URL,
              PublicURL,
              groupId,
              tagIds,
              TLS,
              TLSSkipVerify,
              TLSSkipClientVerify,
              TLSCAFile,
              TLSCertFile,
              TLSKeyFile,
              CheckinInterval,
              $scope.state.isEdgeDevice,
              EdgePingInterval,
              EdgeSnapshotInterval,
              EdgeCommandInterval
            );

            Notifications.success('Environment created', name);
            switch (endpoint.Type) {
              case PortainerEndpointTypes.EdgeAgentOnDockerEnvironment:
              case PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment:
                $state.go('portainer.endpoints.endpoint', { id: endpoint.Id });
                break;
              case PortainerEndpointTypes.AgentOnKubernetesEnvironment:
                $state.go('portainer.k8sendpoint.kubernetesConfig', { id: endpoint.Id });
                break;
              default:
                $state.go('portainer.endpoints', {}, { reload: true });
                break;
            }

            return endpoint;
          } catch (err) {
            Notifications.error('Failure', err, 'Unable to create environment');
          } finally {
            $scope.state.actionInProgress = false;
          }
        });
      }

      function initView() {
        $q.all({
          groups: GroupService.groups(),
          settings: SettingsService.settings(),
        })
          .then(function success(data) {
            $scope.groups = data.groups;

            const settings = data.settings;

            $scope.agentSecret = settings.AgentSecret;
            $scope.state.edgeAsyncMode = settings.Edge.AsyncMode;
          })
          .catch(function error(err) {
            Notifications.error('Failure', err, 'Unable to load groups');
          });
      }

      function agentLinuxSwarmCommand(agentVersion, agentSecret) {
        let secret = agentSecret == '' ? '' : `\\\n  -e AGENT_SECRET=${agentSecret} `;
        return `
docker network create \\
  --driver overlay \\
  portainer_agent_network

docker service create \\
  --name portainer_agent \\
  --network portainer_agent_network \\
  -p 9001:9001/tcp ${secret}\\
  --mode global \\
  --constraint 'node.platform.os == linux' \\
  --mount type=bind,src=//var/run/docker.sock,dst=/var/run/docker.sock \\
  --mount type=bind,src=//var/lib/docker/volumes,dst=/var/lib/docker/volumes \\
  portainer/agent:${agentVersion}
  `;
      }

      function agentWindowsSwarmCommand(agentVersion, agentSecret) {
        let secret = agentSecret == '' ? '' : `\\\n  -e AGENT_SECRET=${agentSecret} `;
        return `
docker network create \\
  --driver overlay \\
  portainer_agent_network && \\
docker service create \\
  --name portainer_agent \\
  --network portainer_agent_network \\
  -p 9001:9001/tcp  ${secret}\\
  --mode global \\
  --constraint 'node.platform.os == windows' \\
  --mount type=npipe,src=\\\\.\\pipe\\docker_engine,dst=\\\\.\\pipe\\docker_engine \\
  --mount type=bind,src=C:\\ProgramData\\docker\\volumes,dst=C:\\ProgramData\\docker\\volumes \\
  portainer/agent:${agentVersion}
  `;
      }

      initView();
    }
  );
