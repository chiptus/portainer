import { SchemaOf, array, object, string } from 'yup';

import { AddonOptionInfo } from '@/react/portainer/environments/wizard/EnvironmentsCreationView/WizardK8sInstall/types';

import { AddOnFormValue, K8sAddOnsForm } from './types';

export function addonsValidation(
  addonOptionsInfo: AddonOptionInfo[],
  initialValues?: K8sAddOnsForm
): SchemaOf<AddOnFormValue[]> {
  return array().of(
    object({
      name: string().required('Addon name is required'),
      arguments: string().test(
        'argument-required',
        'Argument is required',
        // eslint-disable-next-line func-names
        function (value) {
          // if the addon argument is already set, then it's valid
          if (value) {
            return true;
          }
          // If the the addon is already existing with no initial arguments, then don't required arguments at all.
          // This is because a user might have added the addon with the CLI outside of portainer and we cannot read the arguments.
          const existingAddon = initialValues?.addons.find(
            (addonOption) => addonOption.name === this.parent.name
          );
          if (existingAddon && !existingAddon.arguments) {
            return true;
          }
          // otherwise, required should depend on if the matching addon option has required arguments
          const matchingAddonOption = addonOptionsInfo.find(
            (addonOption) => addonOption.label === this.parent.name
          );
          if (!matchingAddonOption) {
            return true;
          }
          if (matchingAddonOption.argumentsType === 'required' && !value) {
            return false;
          }
          return true;
        }
      ),
      repository: string(),
    })
  );
}
