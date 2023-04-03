export interface MTLSCertOptions {
  UseSeparateCert: boolean;
  CaCertFile?: File;
  CertFile?: File;
  KeyFile?: File;
  CaCert?: string;
  Cert?: string;
  Key?: string;
}

export interface FormValues {
  EnableEdgeComputeFeatures: boolean;
  EdgePortainerUrl: string;
  EnforceEdgeID: boolean;
  Edge: {
    TunnelServerAddress: string;
    MTLS: MTLSCertOptions;
  };
}
