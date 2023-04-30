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
  Node,
  ErrorResponse,
  NodeListParams,
  NodeDeleteParams,
  NodeSpec,
  NodeUpdateParams,
} from '../dockerEngineAPI.schemas';

/**
 * @summary List nodes
 */
export const nodeList = (
  endpointId: number,
  params?: NodeListParams,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<Node[]>> =>
  axios.get(`/endpoints/${endpointId}/docker/nodes`, {
    ...options,
    params: { ...params, ...options?.params },
  });

export const getNodeListQueryKey = (
  endpointId: number,
  params?: NodeListParams
) =>
  [
    `/endpoints/${endpointId}/docker/nodes`,
    ...(params ? [params] : []),
  ] as const;

export const getNodeListQueryOptions = <
  TData = Awaited<ReturnType<typeof nodeList>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  params?: NodeListParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof nodeList>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<Awaited<ReturnType<typeof nodeList>>, TError, TData> & {
  queryKey: QueryKey;
} => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getNodeListQueryKey(endpointId, params);

  const queryFn: QueryFunction<Awaited<ReturnType<typeof nodeList>>> = ({
    signal,
  }) => nodeList(endpointId, params, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!endpointId, ...queryOptions };
};

export type NodeListQueryResult = NonNullable<
  Awaited<ReturnType<typeof nodeList>>
>;
export type NodeListQueryError = AxiosError<ErrorResponse>;

export const useNodeList = <
  TData = Awaited<ReturnType<typeof nodeList>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  params?: NodeListParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof nodeList>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getNodeListQueryOptions(endpointId, params, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * @summary Inspect a node
 */
export const nodeInspect = (
  endpointId: number,
  id: string,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<Node>> =>
  axios.get(`/endpoints/${endpointId}/docker/nodes/${id}`, options);

export const getNodeInspectQueryKey = (endpointId: number, id: string) =>
  [`/endpoints/${endpointId}/docker/nodes/${id}`] as const;

export const getNodeInspectQueryOptions = <
  TData = Awaited<ReturnType<typeof nodeInspect>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  id: string,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof nodeInspect>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<Awaited<ReturnType<typeof nodeInspect>>, TError, TData> & {
  queryKey: QueryKey;
} => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getNodeInspectQueryKey(endpointId, id);

  const queryFn: QueryFunction<Awaited<ReturnType<typeof nodeInspect>>> = ({
    signal,
  }) => nodeInspect(endpointId, id, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!(endpointId && id), ...queryOptions };
};

export type NodeInspectQueryResult = NonNullable<
  Awaited<ReturnType<typeof nodeInspect>>
>;
export type NodeInspectQueryError = AxiosError<ErrorResponse>;

export const useNodeInspect = <
  TData = Awaited<ReturnType<typeof nodeInspect>>,
  TError = AxiosError<ErrorResponse>
>(
  endpointId: number,
  id: string,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof nodeInspect>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getNodeInspectQueryOptions(endpointId, id, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * @summary Delete a node
 */
export const nodeDelete = (
  endpointId: number,
  id: string,
  params?: NodeDeleteParams,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.delete(`/endpoints/${endpointId}/docker/nodes/${id}`, {
    ...options,
    params: { ...params, ...options?.params },
  });

export const getNodeDeleteMutationOptions = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof nodeDelete>>,
    TError,
    { endpointId: number; id: string; params?: NodeDeleteParams },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof nodeDelete>>,
  TError,
  { endpointId: number; id: string; params?: NodeDeleteParams },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof nodeDelete>>,
    { endpointId: number; id: string; params?: NodeDeleteParams }
  > = (props) => {
    const { endpointId, id, params } = props ?? {};

    return nodeDelete(endpointId, id, params, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type NodeDeleteMutationResult = NonNullable<
  Awaited<ReturnType<typeof nodeDelete>>
>;

export type NodeDeleteMutationError = AxiosError<ErrorResponse>;

export const useNodeDelete = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof nodeDelete>>,
    TError,
    { endpointId: number; id: string; params?: NodeDeleteParams },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getNodeDeleteMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * @summary Update a node
 */
export const nodeUpdate = (
  endpointId: number,
  id: string,
  nodeSpec: NodeSpec,
  params: NodeUpdateParams,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.post(`/endpoints/${endpointId}/docker/nodes/${id}/update`, nodeSpec, {
    ...options,
    params: { ...params, ...options?.params },
  });

export const getNodeUpdateMutationOptions = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof nodeUpdate>>,
    TError,
    {
      endpointId: number;
      id: string;
      data: NodeSpec;
      params: NodeUpdateParams;
    },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof nodeUpdate>>,
  TError,
  { endpointId: number; id: string; data: NodeSpec; params: NodeUpdateParams },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof nodeUpdate>>,
    { endpointId: number; id: string; data: NodeSpec; params: NodeUpdateParams }
  > = (props) => {
    const { endpointId, id, data, params } = props ?? {};

    return nodeUpdate(endpointId, id, data, params, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type NodeUpdateMutationResult = NonNullable<
  Awaited<ReturnType<typeof nodeUpdate>>
>;
export type NodeUpdateMutationBody = NodeSpec;
export type NodeUpdateMutationError = AxiosError<ErrorResponse>;

export const useNodeUpdate = <
  TError = AxiosError<ErrorResponse>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof nodeUpdate>>,
    TError,
    {
      endpointId: number;
      id: string;
      data: NodeSpec;
      params: NodeUpdateParams;
    },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getNodeUpdateMutationOptions(options);

  return useMutation(mutationOptions);
};