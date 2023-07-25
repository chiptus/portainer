import { useQuery } from 'react-query';

import axios, { parseAxiosError } from '@/portainer/services/axios';
import { withError } from '@/react-tools/react-query';
import { withPaginationQueryParams } from '@/react/common/api/pagination.types';

import { EdgeConfiguration } from '../../types';
import { queryKeys } from '../query-keys';
import { buildUrl } from '../urls';

import { Query, QueryParams } from './types';

async function getList(params: QueryParams) {
  try {
    const { data } = await axios.get<EdgeConfiguration[]>(buildUrl(), {
      params,
    });

    return data;
  } catch (err) {
    throw parseAxiosError(
      err as Error,
      'Failed to get list of edge configurations'
    );
  }
}

export function useList({ page, pageLimit, ...query }: Query = {}) {
  return useQuery(
    queryKeys.list({ page, pageLimit, ...query }),
    async () =>
      getList({ ...query, ...withPaginationQueryParams({ page, pageLimit }) }),
    {
      ...withError('Failure retrieving configurations'),
    }
  );
}

//
// BACKEND PAGINATION
//

// async function getList(params: QueryParams) {
//   try {
//     const response = await axios.get<EdgeConfiguration[]>(buildUrl(), {
//       params,
//     });

//     return withTotalCount(response);
//   } catch (err) {
//     throw parseAxiosError(
//       err as Error,
//       'Failed to get list of edge configurations'
//     );
//   }
// }

// used for backend pagination
// export function useList({ page, pageLimit, ...query }: Query = {}) {
//   const { isLoading, data } = useQuery(
//     queryKeys.list({ page, pageLimit, ...query }),
//     async () =>
//       getList({ ...query, ...withPaginationQueryParams({ page, pageLimit }) }),
//     {
//       // keepPreviousData: true,
//       ...withError('Failure retrieving configurations'),
//     }
//   );
//   return {
//     data,
//     isLoading,
//     configurations: data ? data.value : [],
//     totalCount: data ? data.totalCount : 0,
//   };
// }
