export type StorageQuotaFormValues = {
  className: string;
  enabled: boolean;
  size?: string;
  sizeUnit: 'M' | 'G' | 'T';
};

/**
 * @type limit - The storage limit for a particular class in the Kubernetes namespace.
 * This is represented as a string with the size unit abbreviated (M, G, or T for megabytes, gigabytes, or terabytes respectively).
 * For example, '1G' represents 1 gigabyte, '500M' represents 500 megabytes, and '1T' represents 1 terabyte.
 * This field is optional.
 */
type StorageQuotaPayloadValues = {
  enabled: boolean;
  limit?: string;
};

export type StorageQuotaPayload = Record<string, StorageQuotaPayloadValues>;
