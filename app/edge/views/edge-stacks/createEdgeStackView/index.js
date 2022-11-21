import angular from 'angular';

import { createEdgeStackView } from './create-edge-stack-view';
import { edgeStacksDockerComposeForm } from './docker-compose-form';
import { kubeManifestForm } from './kube-manifest-form';
import { NomadHclForm } from './nomad-hcl-form';
import { kubeDeployDescription } from './kube-deploy-description';

export default angular
  .module('portainer.edge.stacks.create', [])
  .component('createEdgeStackView', createEdgeStackView)
  .component('edgeStacksDockerComposeForm', edgeStacksDockerComposeForm)
  .component('edgeStacksKubeManifestForm', kubeManifestForm)
  .component('edgeStacksNomadHclForm', NomadHclForm)
  .component('kubeDeployDescription', kubeDeployDescription).name;
