export default class GitFormController {
  /* @ngInject */
  constructor() {
    this.onChangeField = this.onChangeField.bind(this);
    this.onChangeURL = this.onChangeField('RepositoryURL');
    this.onChangeRefName = this.onChangeField('RepositoryReferenceName');
    this.onChangeComposePath = this.onChangeField('ComposeFilePathInRepository');
    this.onChangeRepositoryUsername = this.onChangeField('RepositoryUsername');
    this.onChangeRepositoryPassword = this.onChangeField('RepositoryPassword');
    this.onChangeSaveCredential = this.onChangeField('SaveCredential');
    this.onChangeNewCredentialName = this.onChangeField('NewCredentialName');
    this.onChangeRepositoryAuthentication = this.onChangeField('RepositoryAuthentication');
  }

  onChangeField(field) {
    return (value) => {
      this.onChange({
        ...this.model,
        [field]: value,
      });
    };
  }

  $onInit() {
    this.deployMethod = this.deployMethod || 'compose';
  }
}
