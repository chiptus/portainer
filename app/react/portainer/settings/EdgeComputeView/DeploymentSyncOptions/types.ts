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
  Edge: {
    PingInterval: number;
    SnapshotInterval: number;
    CommandInterval: number;
    MTLS: MTLSCertOptions;
  };
  EdgeAgentCheckinInterval: number;
}
