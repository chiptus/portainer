import * as notifications from '@/portainer/services/notifications';
import { EnvironmentId } from '@/portainer/environments/types';
import { Job } from '@/nomad/types';
import { deleteJob } from '@/nomad/jobs.service';

export async function deleteJobs(environmentID: EnvironmentId, jobs: Job[]) {
  return Promise.all(
    jobs.map(async (job) => {
      try {
        await deleteJob(environmentID, job.ID, job.Namespace);
        notifications.success('Job successfully removed', job.ID);
      } catch (err) {
        notifications.error(
          'Failure',
          err as Error,
          `Failed to delete job ${job.ID}`
        );
      }
    })
  );
}
