import axios from '@/portainer/services/axios';

angular.module('portainer.nomad').controller('LogsController', [
  '$scope',
  '$async',
  '$transition$',
  'Notifications',
  function ($scope, $async, $transition$, Notifications) {
    let controller = new AbortController();

    $scope.stderrLog = [];
    $scope.stdoutLog = [];

    $scope.changeLogCollection = function (logCollectionStatus) {
      if (!logCollectionStatus) {
        controller.abort();
        controller = new AbortController();
      } else {
        loadLogs('stderr', $scope.jobID, $scope.taskName, $scope.namespace, $scope.endpointId, controller);
        loadLogs('stdout', $scope.jobID, $scope.taskName, $scope.namespace, $scope.endpointId, controller);
      }
    };

    function stripEscapeCodes(logs) {
      return logs.replace(/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g, '');
    }

    function formatLogs(logs, splitter = '\\n') {
      if (!logs) {
        return [];
      }

      const formattedLogs = [];
      const logInLines = logs.trim().split(splitter);

      for (const logInLine of logInLines) {
        const line = stripEscapeCodes(logInLine).replace('\n', '').replace(/[""]+/g, '');
        formattedLogs.push({ line, spans: [{ foregroundColor: null, backgroundColor: null, text: line }] });
      }

      return formattedLogs;
    }
    async function loadLogs(logType, jobID, taskName, namespace, endpointId, controller, refresh = true, offset = 50000) {
      axios
        .get(`/nomad/allocation/${$scope.allocationID}/logs`, {
          params: {
            jobID,
            taskName,
            namespace,
            endpointId,
            refresh,
            logType,
            offset,
          },
          signal: controller.signal,
          onDownloadProgress: (progressEvent) => {
            $scope[`${logType}Log`] = formatLogs(progressEvent.currentTarget.response);
            $scope.$apply();
          },
        })
        .then((response) => {
          $scope[`${logType}Log`] = formatLogs(response.data, '\n');
          $scope.$apply();
        })
        .catch((err) => {
          if (err.message !== 'canceled') Notifications.error('Failure', err, 'Unable to retrieve task logs');
        });
    }

    async function initView() {
      return $async(async () => {
        $scope.jobID = $transition$.params().jobID;
        $scope.taskName = $transition$.params().taskName;
        $scope.allocationID = $transition$.params().allocationID;
        $scope.namespace = $transition$.params().namespace;
        $scope.endpointId = $transition$.params().endpointId;

        loadLogs('stderr', $scope.jobID, $scope.taskName, $scope.namespace, $scope.endpointId, controller);
        loadLogs('stdout', $scope.jobID, $scope.taskName, $scope.namespace, $scope.endpointId, controller);
      });
    }

    $scope.$on('$destroy', function () {
      if (controller) {
        controller.abort();
      }
    });

    initView();
  },
]);
