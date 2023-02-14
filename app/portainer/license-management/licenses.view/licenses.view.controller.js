import moment from 'moment';
import { confirmDelete } from '@@/modals/confirm';

export default class LicensesViewController {
  /* @ngInject */
  constructor($async, $state, StatusService, LicenseService, Notifications, clipboard) {
    this.$async = $async;
    this.$state = $state;
    this.StatusService = StatusService;
    this.LicenseService = LicenseService;
    this.Notifications = Notifications;
    this.clipboard = clipboard;

    this.info = null;
    this.licenses = null;
    this.usedNodes = 0;
    this.template = 'info';

    this.removeAction = this.removeAction.bind(this);
    this.copyLicenseKey = this.copyLicenseKey.bind(this);
  }

  copyLicenseKey(item) {
    this.clipboard.copyText(item.licenseKey);
  }

  removeAction(licenses) {
    return this.$async(async () => {
      try {
        const validLicensesToRemove = licenses.filter((l) => l.valid);
        if (validLicensesToRemove.length) {
          const validLicenses = this.licenses.filter((l) => l.valid);
          if (validLicenses.length === validLicensesToRemove.length) {
            this.Notifications.warning('At least one valid license is required');
            return;
          }
        }

        if (!(await confirmDelete('Are you sure you want to remove these licenses?'))) {
          return;
        }

        const response = await this.LicenseService.remove(licenses.map((license) => license.licenseKey));
        if (response.failedKeys && Object.keys(response.failedKeys).length) {
          throw new Error('Failed removing licenses');
        }
      } catch (err) {
        this.Notifications.error('Failure', err, 'Failed removing licenses');
      }
      this.$state.reload();
    });
  }

  async $onInit() {
    return this.$async(async () => {
      try {
        const licenses = await this.LicenseService.licenses();
        this.licenses = licenses.map((license) => {
          const expiresAt = moment.unix(license.expiresAt);
          const valid = !license.revoked && moment().isBefore(expiresAt);
          return {
            ...license,
            showExpiresAt: expiresAt.format('YYYY-MM-DD HH:mm'),
            valid,
          };
        });
      } catch (err) {
        this.Notifications.error('Failure', err, 'Failed loading licenses');
      }

      try {
        this.usedNodes = await this.StatusService.nodesCount();
      } catch (err) {
        this.Notifications.error('Failure', err, 'Failed to get nodes count');
      }

      try {
        this.info = await this.LicenseService.info();
        if (this.usedNodes > this.info.nodes) {
          this.template = 'alert';
        }

        this.LicenseService.subscribe((info) => {
          this.info = info;
        });
      } catch (err) {
        this.Notifications.error('Failure', err, 'Failed loading license info');
      }
    });
  }
}
