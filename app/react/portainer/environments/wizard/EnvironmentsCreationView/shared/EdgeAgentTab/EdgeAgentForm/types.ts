import { EdgeAsyncIntervalsValues } from '@/edge/components/EdgeAsyncIntervalsForm';
import { EnvironmentMetadata } from '@/portainer/environments/environment.service/create';

export interface FormValues {
  name: string;

  portainerUrl: string;
  pollFrequency: number;
  meta: EnvironmentMetadata;

  edge: EdgeAsyncIntervalsValues;
}
