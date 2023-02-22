import controller from './nomad-hcl-form.controller.js';

export const NomadHclForm = {
  templateUrl: './nomad-hcl-form.html',
  controller,

  bindings: {
    formValues: '=',
    state: '=',
  },
};
