import angular from 'angular';
import PortainerError from 'Portainer/error';

class KubernetesReplicaSetService {
  /* @ngInject */
  constructor($async, $state, KubernetesReplicaSets, Notifications, KubernetesNamespaceService, Authentication) {
    this.$async = $async;
    this.$state = $state;
    this.Notifications = Notifications;
    this.Authentication = Authentication;
    this.KubernetesNamespaceService = KubernetesNamespaceService;
    this.KubernetesReplicaSets = KubernetesReplicaSets;

    this.getAllAsync = this.getAllAsync.bind(this);
  }

  /**
   * GET
   */
  async getAllAsync(namespace) {
    try {
      const data = await this.KubernetesReplicaSets(namespace).get().$promise;
      return data.items;
    } catch (err) {
      if (err.status === 403) {
        this.Notifications.error('Failure', new Error('Reloading page, as your permissions for namespace ' + namespace + ' appear to have been revoked.'));
        await this.KubernetesNamespaceService.refreshCacheAsync().catch(() => {
          this.Authentication.logout();
          this.$state.go('portainer.logout');
        });
        this.$state.reload();
        return;
      }
      throw new PortainerError('Unable to retrieve ReplicaSets', err);
    }
  }

  get(namespace) {
    return this.$async(this.getAllAsync, namespace);
  }
}

export default KubernetesReplicaSetService;
angular.module('portainer.kubernetes').service('KubernetesReplicaSetService', KubernetesReplicaSetService);
