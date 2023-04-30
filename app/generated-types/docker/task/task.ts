/**
 * Generated by orval v6.15.0 🍺
 * Do not edit manually.
 * Docker Engine API
 * The Engine API is an HTTP API served by Docker Engine. It is the API the
Docker client uses to communicate with the Engine, so everything the Docker
client can do can be done with the API.

Most of the client's commands map directly to API endpoints (e.g. `docker ps`
is `GET /containers/json`). The notable exception is running containers,
which consists of several API calls.

# Errors

The API uses standard HTTP status codes to indicate the success or failure
of the API call. The body of the response will be JSON in the following
format:

```
{
  "message": "page not found"
}
```

# Versioning

The API is usually changed in each release, so API calls are versioned to
ensure that clients don't break. To lock to a specific version of the API,
you prefix the URL with its version, for example, call `/v1.30/info` to use
the v1.30 version of the `/info` endpoint. If the API version specified in
the URL is not supported by the daemon, a HTTP `400 Bad Request` error message
is returned.

If you omit the version-prefix, the current version of the API (v1.41) is used.
For example, calling `/info` is the same as calling `/v1.41/info`. Using the
API without a version-prefix is deprecated and will be removed in a future release.

Engine releases in the near future should support this version of the API,
so your client will continue to work even if it is talking to a newer Engine.

The API uses an open schema model, which means server may add extra properties
to responses. Likewise, the server will ignore any extra query parameters and
request body properties. When you write clients, you need to ignore additional
properties in responses to ensure they do not break when talking to newer
daemons.


# Authentication

Authentication for registries is handled client side. The client has to send
authentication details to various endpoints that need to communicate with
registries, such as `POST /images/(name)/push`. These are sent as
`X-Registry-Auth` header as a [base64url encoded](https://tools.ietf.org/html/rfc4648#section-5)
(JSON) string with the following structure:

```
{
  "username": "string",
  "password": "string",
  "email": "string",
  "serveraddress": "string"
}
```

The `serveraddress` is a domain/IP without a protocol. Throughout this
structure, double quotes are required.

If you have already got an identity token from the [`/auth` endpoint](#operation/SystemAuth),
you can just pass this instead of credentials:

```
{
  "identitytoken": "9cbaf023786cd7..."
}
```

 * OpenAPI spec version: 1.41
 */
import axios from 'axios';
import type { AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';
import { useQuery } from 'react-query';
import type {
  UseQueryOptions,
  QueryFunction,
  UseQueryResult,
  QueryKey,
} from 'react-query';

import type {
  Task,
  ErrorResponse,
  TaskListParams,
  TaskLogsParams,
} from '../dockerEngineAPI.schemas';

/**
 * @summary List tasks
 */
export const taskList = (
  endpointId: number,
  params?: TaskListParams,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<Task[]>> =>
  axios.get(`/endpoints/${endpointId}/docker/tasks`, {
    ...options,
    params: { ...params, ...options?.params },
  });

export const getTaskListQueryKey = (
  endpointId: number,
  params?: TaskListParams
) =>
  [
    `/endpoints/${endpointId}/docker/tasks`,
    ...(params ? [params] : []),
  ] as const;

export const getTaskListQueryOptions = <
  TData = Awaited<ReturnType<typeof taskList>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  params?: TaskListParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof taskList>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<Awaited<ReturnType<typeof taskList>>, TError, TData> & {
  queryKey: QueryKey;
} => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getTaskListQueryKey(endpointId, params);

  const queryFn: QueryFunction<Awaited<ReturnType<typeof taskList>>> = ({
    signal,
  }) => taskList(endpointId, params, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!endpointId, ...queryOptions };
};

export type TaskListQueryResult = NonNullable<
  Awaited<ReturnType<typeof taskList>>
>;
export type TaskListQueryError = AxiosError<ErrorResponse>;

export const useTaskList = <
  TData = Awaited<ReturnType<typeof taskList>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  params?: TaskListParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof taskList>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getTaskListQueryOptions(endpointId, params, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * @summary Inspect a task
 */
export const taskInspect = (
  endpointId: number,
  id: string,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<Task>> =>
  axios.get(`/endpoints/${endpointId}/docker/tasks/${id}`, options);

export const getTaskInspectQueryKey = (endpointId: number, id: string) =>
  [`/endpoints/${endpointId}/docker/tasks/${id}`] as const;

export const getTaskInspectQueryOptions = <
  TData = Awaited<ReturnType<typeof taskInspect>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  id: string,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof taskInspect>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<Awaited<ReturnType<typeof taskInspect>>, TError, TData> & {
  queryKey: QueryKey;
} => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getTaskInspectQueryKey(endpointId, id);

  const queryFn: QueryFunction<Awaited<ReturnType<typeof taskInspect>>> = ({
    signal,
  }) => taskInspect(endpointId, id, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!(endpointId && id), ...queryOptions };
};

export type TaskInspectQueryResult = NonNullable<
  Awaited<ReturnType<typeof taskInspect>>
>;
export type TaskInspectQueryError = AxiosError<ErrorResponse>;

export const useTaskInspect = <
  TData = Awaited<ReturnType<typeof taskInspect>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  id: string,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof taskInspect>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getTaskInspectQueryOptions(endpointId, id, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * Get `stdout` and `stderr` logs from a task.
See also [`/containers/{id}/logs`](#operation/ContainerLogs).

**Note**: This endpoint works only for services with the `local`,
`json-file` or `journald` logging drivers.

 * @summary Get task logs
 */
export const taskLogs = (
  endpointId: number,
  id: string,
  params?: TaskLogsParams,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<Blob>> =>
  axios.get(`/endpoints/${endpointId}/docker/tasks/${id}/logs`, {
    responseType: 'blob',
    ...options,
    params: { ...params, ...options?.params },
  });

export const getTaskLogsQueryKey = (
  endpointId: number,
  id: string,
  params?: TaskLogsParams
) =>
  [
    `/endpoints/${endpointId}/docker/tasks/${id}/logs`,
    ...(params ? [params] : []),
  ] as const;

export const getTaskLogsQueryOptions = <
  TData = Awaited<ReturnType<typeof taskLogs>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  id: string,
  params?: TaskLogsParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof taskLogs>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<Awaited<ReturnType<typeof taskLogs>>, TError, TData> & {
  queryKey: QueryKey;
} => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getTaskLogsQueryKey(endpointId, id, params);

  const queryFn: QueryFunction<Awaited<ReturnType<typeof taskLogs>>> = ({
    signal,
  }) => taskLogs(endpointId, id, params, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!(endpointId && id), ...queryOptions };
};

export type TaskLogsQueryResult = NonNullable<
  Awaited<ReturnType<typeof taskLogs>>
>;
export type TaskLogsQueryError = AxiosError<ErrorResponse>;

export const useTaskLogs = <
  TData = Awaited<ReturnType<typeof taskLogs>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  id: string,
  params?: TaskLogsParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof taskLogs>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getTaskLogsQueryOptions(endpointId, id, params, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};