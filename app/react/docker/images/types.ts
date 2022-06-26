type Status = 'outdated' | 'updated' | 'inprocess' | string;

export interface ImageStatus {
  Status: Status;
  Message: string;
}
