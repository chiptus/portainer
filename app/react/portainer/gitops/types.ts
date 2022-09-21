import { UserGitCredential } from '@/portainer/models/user';

export interface GitFormModel {
  RepositoryURL: string;
  RepositoryURLValid: boolean;
  ComposeFilePathInRepository: string;
  RepositoryAuthentication: boolean;
  RepositoryReferenceName?: string;
  RepositoryUsername?: string;
  RepositoryPassword?: string;
  SelectedGitCredential?: UserGitCredential;
}
