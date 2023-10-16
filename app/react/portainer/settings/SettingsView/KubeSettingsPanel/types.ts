export interface FormValues {
  helmRepositoryURL: string;
  kubeconfigExpiry: string;
  globalDeploymentOptions: {
    hideAddWithForm: boolean;
    perEnvOverride: boolean;
    hideWebEditor: boolean;
    hideFileUpload: boolean;
    requireNoteOnApplications: boolean;
    minApplicationNoteLength: number;
    hideStacksFunctionality: boolean;
  };
}
