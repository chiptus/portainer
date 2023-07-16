export class EditEdgeStackViewController {
  /* @ngInject */
  constructor($async, $state, EdgeStackService, Notifications) {
    this.$async = $async;
    this.$state = $state;
    this.EdgeStackService = EdgeStackService;
    this.Notifications = Notifications;

    this.stack = null;

    this.state = {
      activeTab: 0,
    };
  }

  async $onInit() {
    return this.$async(async () => {
      const { stackId, tab } = this.$state.params;
      this.state.activeTab = tab ? parseInt(tab, 10) : 0;
      try {
        this.stack = await this.EdgeStackService.stack(stackId);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve stack data');
      }
    });
  }
}
