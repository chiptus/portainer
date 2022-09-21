export type SupportedIngControllerNames = 'nginx' | 'traefik' | 'unknown';

export interface IngressControllerClassMap extends Record<string, unknown> {
  Name: string;
  ClassName: string;
  Type: SupportedIngControllerNames;
  Availability: boolean;
}
