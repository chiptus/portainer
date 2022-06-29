import { useQuery } from 'react-query';

import { EnvironmentId } from '@/portainer/environments/types';
import { listJobs } from '@/nomad/jobs.service';
import { Job } from '@/nomad/types';

export function useJobs(environmentId: EnvironmentId) {
  return useQuery<Job[]>(
    ['environments', environmentId, 'nomad', 'jobs'],
    () => listJobs(environmentId),
    {
      meta: {
        error: {
          title: 'Failure',
          message: 'Unable to list jobs',
        },
      },
    }
  );
}
