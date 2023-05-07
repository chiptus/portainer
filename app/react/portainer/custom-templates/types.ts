import { VariableDefinition } from './components/CustomTemplatesVariablesDefinitionField';

export type CustomTemplate = {
  Id: number;
  Title: string;
  Type: number;
  Variables: VariableDefinition[];
  Path: string;
};

export type CustomTemplateFileContent = {
  FileContent: string;
};

export const CustomTemplateKubernetesType = 3;
