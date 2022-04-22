import _ from 'lodash';
import { isUnassociatedEdgeEnvironment } from '@/portainer/environments/utils';

export default class WizardEndpointsController {
  /* @ngInject */
  constructor($async, $state, EndpointService, Notifications, $analytics, StateManager) {
    this.$async = $async;
    this.$state = $state;
    this.EndpointService = EndpointService;
    this.$analytics = $analytics;
    this.Notifications = Notifications;

    this.endpoints = [];
    this.agentVersion = StateManager.getState().application.version;

    this.reloadEndpoints = this.reloadEndpoints.bind(this);
    this.addAnalytics = this.addAnalytics.bind(this);
  }
  /**
   * WIZARD ENDPOINT SECTION
   */

  async reloadEndpoints() {
    // TODO: when converting to react, should use `useEnvironmentsList` hook like in HomeView, and reload if has unassociated edge environment
    try {
      const { value } = await this.EndpointService.endpoints();
      this.endpoints = value;

      if (this.endpoints.some((env) => isUnassociatedEdgeEnvironment(env))) {
        this.startReloadLoop();
      }
    } catch (e) {
      this.Notifications.error('Failure', e, 'Unable to retrieve endpoints');
    }
  }

  startReloadLoop() {
    this.clearReloadLoop();
    this.reloadInterval = setTimeout(this.reloadEndpoints, 5000);
  }

  clearReloadLoop() {
    if (this.reloadInterval) {
      clearInterval(this.reloadInterval);
    }
  }

  startWizard() {
    const options = this.state.options;
    this.state.selections = options.filter((item) => item.selected === true);
    this.state.maxStep = this.state.selections.length;

    if (this.state.selections.length !== 0) {
      this.state.section = this.state.selections[this.state.currentStep].endpoint;
      this.state.sectionTitle = this.state.selections[this.state.currentStep].title;
      this.state.selections[this.state.currentStep].stage = 'active';
    }

    if (this.state.currentStep === this.state.maxStep - 1) {
      this.state.nextStep = 'Finish';
    }

    this.$analytics.eventTrack('endpoint-wizard-endpoint-select', {
      category: 'portainer',
      metadata: {
        environment: _.compact([this.state.analytics.docker, this.state.analytics.kubernetes, this.state.analytics.aci, this.state.analytics.nomad]).join('/'),
      },
    });
    this.state.currentStep++;
  }

  previousStep() {
    this.state.section = this.state.selections[this.state.currentStep - 2].endpoint;
    this.state.sectionTitle = this.state.selections[this.state.currentStep - 2].title;
    this.state.selections[this.state.currentStep - 2].stage = 'active';
    this.state.selections[this.state.currentStep - 1].stage = '';
    this.state.nextStep = 'Next Step';
    this.state.currentStep--;
  }

  async nextStep() {
    if (this.state.currentStep >= this.state.maxStep - 1) {
      this.state.nextStep = 'Finish';
    }
    if (this.state.currentStep === this.state.maxStep) {
      // the Local Endpoint Counter from endpoints array due to including Local Endpoint been added Automatic before Wizard start
      const endpointsAdded = await this.EndpointService.endpoints();
      const endpointsArray = endpointsAdded.value;
      const filter = endpointsArray.filter((item) => item.Type === 1 || item.Type === 5);

      // NOTICE: This is the temporary fix for excluded docker api endpoint been counted as local endpoint
      this.state.counter.localEndpoint = filter.length - this.state.counter.dockerApi;

      this.$analytics.eventTrack('endpoint-wizard-environment-add-finish', {
        category: 'portainer',
        metadata: {
          'docker-agent': this.state.counter.dockerAgent,
          'docker-api': this.state.counter.dockerApi,
          'kubernetes-agent': this.state.counter.kubernetesAgent,
          'kaas-agent': this.state.counter.kaasAgent,
          'aci-api': this.state.counter.aciApi,
          'local-endpoint': this.state.counter.localEndpoint,
        },
      });
      this.$state.go('portainer.home');
    } else {
      this.state.section = this.state.selections[this.state.currentStep].endpoint;
      this.state.sectionTitle = this.state.selections[this.state.currentStep].title;
      this.state.selections[this.state.currentStep].stage = 'active';
      this.state.selections[this.state.currentStep - 1].stage = 'completed';
      this.state.currentStep++;
    }
  }

  addAnalytics(endpoint) {
    switch (endpoint) {
      case 'docker-agent':
        this.state.counter.dockerAgent++;
        break;
      case 'docker-api':
        this.state.counter.dockerApi++;
        break;
      case 'kubernetes-agent':
        this.state.counter.kubernetesAgent++;
        break;
      case 'kaas-agent':
        this.state.counter.kaasAgent++;
        break;
      case 'aci-api':
        this.state.counter.aciApi++;
        break;
    }
  }

  endpointSelect(endpoint) {
    switch (endpoint) {
      case 'docker':
        if (this.state.options[0].selected) {
          this.state.options[0].selected = false;
          this.state.dockerActive = '';
          this.state.analytics.docker = '';
        } else {
          this.state.options[0].selected = true;
          this.state.dockerActive = 'wizard-active';
          this.state.analytics.docker = 'Docker';
        }
        break;
      case 'kubernetes':
        if (this.state.options[1].selected) {
          this.state.options[1].selected = false;
          this.state.kubernetesActive = '';
          this.state.analytics.kubernetes = '';
        } else {
          this.state.options[1].selected = true;
          this.state.kubernetesActive = 'wizard-active';
          this.state.analytics.kubernetes = 'Kubernetes';
        }
        break;
      case 'kaas':
        if (this.state.options[2].selected) {
          this.state.options[2].selected = false;
          this.state.kaasActive = '';
          this.state.analytics.kaas = '';
        } else {
          this.state.options[2].selected = true;
          this.state.kaasActive = 'wizard-active';
          this.state.analytics.kaas = 'Kaas/';
        }
        break;
      case 'aci':
        if (this.state.options[3].selected) {
          this.state.options[3].selected = false;
          this.state.aciActive = '';
          this.state.analytics.aci = '';
        } else {
          this.state.options[3].selected = true;
          this.state.aciActive = 'wizard-active';
          this.state.analytics.aci = 'ACI';
        }
        break;
      case 'nomad':
        if (this.state.options[4].selected) {
          this.state.options[4].selected = false;
          this.state.nomadActive = '';
          this.state.analytics.nomad = '';
        } else {
          this.state.options[4].selected = true;
          this.state.nomadActive = 'wizard-active';
          this.state.analytics.nomad = 'Nomad';
        }
        break;
    }
    const options = this.state.options;
    this.state.selections = options.filter((item) => item.selected === true);
  }

  $onInit() {
    return this.$async(async () => {
      this.state = {
        currentStep: 0,
        section: '',
        sectionTitle: '',
        dockerActive: '',
        kubernetesActive: '',
        kaasActive: '',
        aciActive: '',
        nomadActive: '',
        maxStep: '',
        previousStep: 'Previous',
        nextStep: 'Next Step',
        selections: [],
        analytics: {
          docker: '',
          kubernetes: '',
          kaas: '',
          aci: '',
          nomad: '',
        },
        counter: {
          dockerAgent: 0,
          dockerApi: 0,
          kubernetesAgent: 0,
          kaasAgent: 0,
          aciApi: 0,
          localEndpoint: 0,
        },
        options: [
          {
            endpoint: 'docker',
            title: 'Docker',
            selected: false,
            stage: '',
            nameClass: 'docker',
            icon: 'fab fa-docker',
          },
          {
            endpoint: 'kubernetes',
            title: 'Kubernetes',
            selected: false,
            stage: '',
            nameClass: 'kubernetes',
            icon: 'fas fa-dharmachakra',
          },
          {
            endpoint: 'kaas',
            title: 'KaaS',
            selected: false,
            stage: '',
            nameClass: 'kaas',
            icon: 'fas fa-dharmachakra',
          },
          {
            endpoint: 'aci',
            title: 'ACI',
            selected: false,
            stage: '',
            nameClass: 'aci',
            icon: 'fab fa-microsoft',
          },
          {
            endpoint: 'nomad',
            title: 'Nomad',
            selected: false,
            stage: '',
            nameClass: 'nomad',
            icon: 'nomad-icon',
          },
        ],
        selectOption: '',
      };

      this.reloadEndpoints();
    });
  }

  $onDestroy() {
    this.clearReloadLoop();
  }
}
