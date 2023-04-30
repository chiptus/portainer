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
import { useQuery, useMutation } from 'react-query';
import type {
  UseQueryOptions,
  UseMutationOptions,
  QueryFunction,
  MutationFunction,
  UseQueryResult,
  QueryKey,
} from 'react-query';

import type {
  Swarm,
  ErrorResponse,
  SwarmInitBodyOne,
  SwarmInitBodyTwo,
  SwarmJoinBodyOne,
  SwarmJoinBodyTwo,
  SwarmLeaveParams,
  SwarmSpec,
  SwarmUpdateParams,
  SwarmUnlockkey200One,
  SwarmUnlockkey200Two,
  SwarmUnlockBody,
} from '../dockerEngineAPI.schemas';

/**
 * @summary Inspect swarm
 */
export const swarmInspect = (
  endpointId: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<Swarm>> =>
  axios.get(`/endpoints/${endpointId}/docker/swarm`, options);

export const getSwarmInspectQueryKey = (endpointId: number) =>
  [`/endpoints/${endpointId}/docker/swarm`] as const;

export const getSwarmInspectQueryOptions = <
  TData = Awaited<ReturnType<typeof swarmInspect>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof swarmInspect>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<Awaited<ReturnType<typeof swarmInspect>>, TError, TData> & {
  queryKey: QueryKey;
} => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getSwarmInspectQueryKey(endpointId);

  const queryFn: QueryFunction<Awaited<ReturnType<typeof swarmInspect>>> = ({
    signal,
  }) => swarmInspect(endpointId, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!endpointId, ...queryOptions };
};

export type SwarmInspectQueryResult = NonNullable<
  Awaited<ReturnType<typeof swarmInspect>>
>;
export type SwarmInspectQueryError = AxiosError<ErrorResponse>;

export const useSwarmInspect = <
  TData = Awaited<ReturnType<typeof swarmInspect>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof swarmInspect>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getSwarmInspectQueryOptions(endpointId, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * @summary Initialize a new swarm
 */
export const swarmInit = (
  endpointId: number,
  swarmInitBody: SwarmInitBodyOne | SwarmInitBodyTwo,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<string>> =>
  axios.post(
    `/endpoints/${endpointId}/docker/swarm/init`,
    swarmInitBody,
    options
  );

export const getSwarmInitMutationOptions = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmInit>>,
    TError,
    { endpointId: number; data: SwarmInitBodyOne | SwarmInitBodyTwo },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof swarmInit>>,
  TError,
  { endpointId: number; data: SwarmInitBodyOne | SwarmInitBodyTwo },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof swarmInit>>,
    { endpointId: number; data: SwarmInitBodyOne | SwarmInitBodyTwo }
  > = (props) => {
    const { endpointId, data } = props ?? {};

    return swarmInit(endpointId, data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type SwarmInitMutationResult = NonNullable<
  Awaited<ReturnType<typeof swarmInit>>
>;
export type SwarmInitMutationBody = SwarmInitBodyOne | SwarmInitBodyTwo;
export type SwarmInitMutationError = AxiosError<ErrorResponse>;

export const useSwarmInit = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmInit>>,
    TError,
    { endpointId: number; data: SwarmInitBodyOne | SwarmInitBodyTwo },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getSwarmInitMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * @summary Join an existing swarm
 */
export const swarmJoin = (
  endpointId: number,
  swarmJoinBody: SwarmJoinBodyOne | SwarmJoinBodyTwo,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.post(
    `/endpoints/${endpointId}/docker/swarm/join`,
    swarmJoinBody,
    options
  );

export const getSwarmJoinMutationOptions = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmJoin>>,
    TError,
    { endpointId: number; data: SwarmJoinBodyOne | SwarmJoinBodyTwo },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof swarmJoin>>,
  TError,
  { endpointId: number; data: SwarmJoinBodyOne | SwarmJoinBodyTwo },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof swarmJoin>>,
    { endpointId: number; data: SwarmJoinBodyOne | SwarmJoinBodyTwo }
  > = (props) => {
    const { endpointId, data } = props ?? {};

    return swarmJoin(endpointId, data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type SwarmJoinMutationResult = NonNullable<
  Awaited<ReturnType<typeof swarmJoin>>
>;
export type SwarmJoinMutationBody = SwarmJoinBodyOne | SwarmJoinBodyTwo;
export type SwarmJoinMutationError = AxiosError<ErrorResponse>;

export const useSwarmJoin = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmJoin>>,
    TError,
    { endpointId: number; data: SwarmJoinBodyOne | SwarmJoinBodyTwo },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getSwarmJoinMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * @summary Leave a swarm
 */
export const swarmLeave = (
  endpointId: number,
  params?: SwarmLeaveParams,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.post(`/endpoints/${endpointId}/docker/swarm/leave`, undefined, {
    ...options,
    params: { ...params, ...options?.params },
  });

export const getSwarmLeaveMutationOptions = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmLeave>>,
    TError,
    { endpointId: number; params?: SwarmLeaveParams },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof swarmLeave>>,
  TError,
  { endpointId: number; params?: SwarmLeaveParams },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof swarmLeave>>,
    { endpointId: number; params?: SwarmLeaveParams }
  > = (props) => {
    const { endpointId, params } = props ?? {};

    return swarmLeave(endpointId, params, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type SwarmLeaveMutationResult = NonNullable<
  Awaited<ReturnType<typeof swarmLeave>>
>;

export type SwarmLeaveMutationError = AxiosError<ErrorResponse>;

export const useSwarmLeave = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmLeave>>,
    TError,
    { endpointId: number; params?: SwarmLeaveParams },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getSwarmLeaveMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * @summary Update a swarm
 */
export const swarmUpdate = (
  endpointId: number,
  swarmSpec: SwarmSpec,
  params: SwarmUpdateParams,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.post(`/endpoints/${endpointId}/docker/swarm/update`, swarmSpec, {
    ...options,
    params: { ...params, ...options?.params },
  });

export const getSwarmUpdateMutationOptions = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmUpdate>>,
    TError,
    { endpointId: number; data: SwarmSpec; params: SwarmUpdateParams },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof swarmUpdate>>,
  TError,
  { endpointId: number; data: SwarmSpec; params: SwarmUpdateParams },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof swarmUpdate>>,
    { endpointId: number; data: SwarmSpec; params: SwarmUpdateParams }
  > = (props) => {
    const { endpointId, data, params } = props ?? {};

    return swarmUpdate(endpointId, data, params, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type SwarmUpdateMutationResult = NonNullable<
  Awaited<ReturnType<typeof swarmUpdate>>
>;
export type SwarmUpdateMutationBody = SwarmSpec;
export type SwarmUpdateMutationError = AxiosError<ErrorResponse>;

export const useSwarmUpdate = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmUpdate>>,
    TError,
    { endpointId: number; data: SwarmSpec; params: SwarmUpdateParams },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getSwarmUpdateMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * @summary Get the unlock key
 */
export const swarmUnlockkey = (
  endpointId: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<SwarmUnlockkey200One | SwarmUnlockkey200Two>> =>
  axios.get(`/endpoints/${endpointId}/docker/swarm/unlockkey`, options);

export const getSwarmUnlockkeyQueryKey = (endpointId: number) =>
  [`/endpoints/${endpointId}/docker/swarm/unlockkey`] as const;

export const getSwarmUnlockkeyQueryOptions = <
  TData = Awaited<ReturnType<typeof swarmUnlockkey>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof swarmUnlockkey>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<
  Awaited<ReturnType<typeof swarmUnlockkey>>,
  TError,
  TData
> & { queryKey: QueryKey } => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getSwarmUnlockkeyQueryKey(endpointId);

  const queryFn: QueryFunction<Awaited<ReturnType<typeof swarmUnlockkey>>> = ({
    signal,
  }) => swarmUnlockkey(endpointId, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!endpointId, ...queryOptions };
};

export type SwarmUnlockkeyQueryResult = NonNullable<
  Awaited<ReturnType<typeof swarmUnlockkey>>
>;
export type SwarmUnlockkeyQueryError = AxiosError<ErrorResponse>;

export const useSwarmUnlockkey = <
  TData = Awaited<ReturnType<typeof swarmUnlockkey>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof swarmUnlockkey>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getSwarmUnlockkeyQueryOptions(endpointId, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * @summary Unlock a locked manager
 */
export const swarmUnlock = (
  endpointId: number,
  swarmUnlockBody: SwarmUnlockBody,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.post(
    `/endpoints/${endpointId}/docker/swarm/unlock`,
    swarmUnlockBody,
    options
  );

export const getSwarmUnlockMutationOptions = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmUnlock>>,
    TError,
    { endpointId: number; data: SwarmUnlockBody },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof swarmUnlock>>,
  TError,
  { endpointId: number; data: SwarmUnlockBody },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof swarmUnlock>>,
    { endpointId: number; data: SwarmUnlockBody }
  > = (props) => {
    const { endpointId, data } = props ?? {};

    return swarmUnlock(endpointId, data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type SwarmUnlockMutationResult = NonNullable<
  Awaited<ReturnType<typeof swarmUnlock>>
>;
export type SwarmUnlockMutationBody = SwarmUnlockBody;
export type SwarmUnlockMutationError = AxiosError<ErrorResponse>;

export const useSwarmUnlock = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof swarmUnlock>>,
    TError,
    { endpointId: number; data: SwarmUnlockBody },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getSwarmUnlockMutationOptions(options);

  return useMutation(mutationOptions);
};
