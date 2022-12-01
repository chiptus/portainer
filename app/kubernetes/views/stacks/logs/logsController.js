import _ from 'lodash-es';
import angular from 'angular';
import $allSettled from 'Portainer/services/allSettled';

const colors = ['red', 'orange', 'lime', 'green', 'darkgreen', 'cyan', 'turquoise', 'teal', 'deepskyblue', 'blue', 'darkblue', 'slateblue', 'magenta', 'darkviolet'];

class KubernetesStackLogsController {
  /* @ngInject */
  constructor(Notifications, KubernetesApplicationService, KubernetesPodService) {
    this.Notifications = Notifications;
    this.KubernetesApplicationService = KubernetesApplicationService;
    this.KubernetesPodService = KubernetesPodService;

    this.generateLogsPromise = this.generateLogsPromise.bind(this);
    this.generateAppPromise = this.generateAppPromise.bind(this);
    this.getStackLogsAsync = this.getStackLogsAsync.bind(this);
  }

  async generateLogsPromise(pod, container, params) {
    const res = {
      Pod: pod,
      Logs: [],
    };
    res.Logs = await this.KubernetesPodService.logs(pod.Namespace, pod.Name, container.Name, params);
    return res;
  }

  generateAppPromise(params) {
    return async (app) => {
      const res = {
        Application: app,
        Pods: [],
      };

      const promises = _.flatMap(_.map(app.Pods, (pod) => _.map(pod.Containers, (container) => this.generateLogsPromise(pod, container, params))));
      const result = await $allSettled(promises);
      res.Pods = result.fulfilled;
      return res;
    }
  }

  async getStackLogsAsync(params) {
    let colorIndex = -1;
    try {
      const applications = await this.KubernetesApplicationService.get(this.$transition$.params().namespace);
      const filteredApplications = _.filter(applications, (app) => app.StackName === this.$transition$.params().name);
      const logsPromises = _.map(filteredApplications, this.generateAppPromise(params));
      const data = await Promise.all(logsPromises);
      const logs = _.flatMap(data, (app) => {
        return _.flatMap(app.Pods, (pod) => {
          colorIndex += 1;
          return {
            sectionName: pod.Pod.Name,
            sectionNameColor: colors[colorIndex % colors.length],
            logs: pod.Logs.logs,
          }
        });
      });

      return logs
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve application logs');
    }
  }
}

export default KubernetesStackLogsController;
angular.module('portainer.kubernetes').controller('KubernetesStackLogsController', KubernetesStackLogsController);
