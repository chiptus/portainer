export interface AddOnFormValue {
  name: string;
  arguments?: string;
  repository?: string;
}

export type K8sAddOnsForm = {
  addons: AddOnFormValue[];
  currentVersion: string;
};

export type AddonsArgumentType = 'required' | 'optional' | '';

export type AddOnOption = {
  label: string;
  name: string;
  repository?: string;

  arguments?: string;
  tooltip?: string;
  placeholder?: string;
  argumentsType?: AddonsArgumentType;
};
