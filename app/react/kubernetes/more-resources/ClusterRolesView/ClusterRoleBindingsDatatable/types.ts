export type ClusterRoleRef = {
  name: string;
  kind: string;
  apiGroup?: string;
};

export type ClusterRoleSubject = {
  name: string;
  kind: string;
  apiGroup?: string;
  namespace?: string;
};

export type ClusterRoleBinding = {
  name: string;
  uid: string;
  namespace: string;
  resourceVersion: string;
  creationDate: string;
  annotations?: Record<string, string>;

  roleRef: ClusterRoleRef;
  subjects: ClusterRoleSubject[];

  isSystem: boolean;
};
