/**
 * Generated by orval v6.15.0 🍺
 * Do not edit manually.
 * PortainerEE API
 * Portainer API is an HTTP API served by Portainer. It is used by the Portainer UI and everything you can do with the UI can be done using the HTTP API.
Examples are available at https://documentation.portainer.io/api/api-examples/
You can find out more about Portainer at [http://portainer.io](http://portainer.io) and get some support on [Slack](http://portainer.io/slack/).

# Authentication

Most of the API environments(endpoints) require to be authenticated as well as some level of authorization to be used.
Portainer API uses JSON Web Token to manage authentication and thus requires you to provide a token in the **Authorization** header of each request
with the **Bearer** authentication mechanism.

Example:

```
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MSwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOjEsImV4cCI6MTQ5OTM3NjE1NH0.NJ6vE8FY1WG6jsRQzfMqeatJ4vh2TWAeeYfDhP71YEE
```

# Security

Each API environment(endpoint) has an associated access policy, it is documented in the description of each environment(endpoint).

Different access policies are available:

- Public access
- Authenticated access
- Restricted access
- Administrator access

### Public access

No authentication is required to access the environments(endpoints) with this access policy.

### Authenticated access

Authentication is required to access the environments(endpoints) with this access policy.

### Restricted access

Authentication is required to access the environments(endpoints) with this access policy.
Extra-checks might be added to ensure access to the resource is granted. Returned data might also be filtered.

### Administrator access

Authentication as well as an administrator role are required to access the environments(endpoints) with this access policy.

# Execute Docker requests

Portainer **DO NOT** expose specific environments(endpoints) to manage your Docker resources (create a container, remove a volume, etc...).

Instead, it acts as a reverse-proxy to the Docker HTTP API. This means that you can execute Docker requests **via** the Portainer HTTP API.

To do so, you can use the `/endpoints/{id}/docker` Portainer API environment(endpoint) (which is not documented below due to Swagger limitations). This environment(endpoint) has a restricted access policy so you still need to be authenticated to be able to query this environment(endpoint). Any query on this environment(endpoint) will be proxied to the Docker API of the associated environment(endpoint) (requests and responses objects are the same as documented in the Docker API).

# Private Registry

Using private registry, you will need to pass a based64 encoded JSON string ‘{"registryId":\<registryID value\>}’ inside the Request Header. The parameter name is "X-Registry-Auth".
\<registryID value\> - The registry ID where the repository was created.

Example:

```
eyJyZWdpc3RyeUlkIjoxfQ==
```

**NOTE**: You can find more information on how to query the Docker API in the [Docker official documentation](https://docs.docker.com/engine/api/v1.30/) as well as in [this Portainer example](https://documentation.portainer.io/api/api-examples/).

 * OpenAPI spec version: 2.19.0
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
  PortainereeEndpointGroup,
  EndpointgroupsEndpointGroupCreatePayload,
  EndpointgroupsEndpointGroupUpdatePayload,
} from '../portainerEEAPI.schemas';

/**
 * List all environment(endpoint) groups based on the current user authorizations. Will
return all environment(endpoint) groups if using an administrator account otherwise it will
only return authorized environment(endpoint) groups.
**Access policy**: restricted
 * @summary List Environment(Endpoint) groups
 */
export const endpointGroupList = (
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeEndpointGroup[]>> =>
  axios.get(`/endpoint_groups`, options);

export const getEndpointGroupListQueryKey = () => [`/endpoint_groups`] as const;

export const getEndpointGroupListQueryOptions = <
  TData = Awaited<ReturnType<typeof endpointGroupList>>,
  TError = AxiosError<void>
>(options?: {
  query?: UseQueryOptions<
    Awaited<ReturnType<typeof endpointGroupList>>,
    TError,
    TData
  >;
  axios?: AxiosRequestConfig;
}): UseQueryOptions<
  Awaited<ReturnType<typeof endpointGroupList>>,
  TError,
  TData
> & { queryKey: QueryKey } => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey = queryOptions?.queryKey ?? getEndpointGroupListQueryKey();

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof endpointGroupList>>
  > = ({ signal }) => endpointGroupList({ signal, ...axiosOptions });

  return { queryKey, queryFn, ...queryOptions };
};

export type EndpointGroupListQueryResult = NonNullable<
  Awaited<ReturnType<typeof endpointGroupList>>
>;
export type EndpointGroupListQueryError = AxiosError<void>;

export const useEndpointGroupList = <
  TData = Awaited<ReturnType<typeof endpointGroupList>>,
  TError = AxiosError<void>
>(options?: {
  query?: UseQueryOptions<
    Awaited<ReturnType<typeof endpointGroupList>>,
    TError,
    TData
  >;
  axios?: AxiosRequestConfig;
}): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getEndpointGroupListQueryOptions(options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * Create a new environment(endpoint) group.
 **Access policy**: administrator
 * @summary Create an Environment(Endpoint) Group
 */
export const postEndpointGroups = (
  endpointgroupsEndpointGroupCreatePayload: EndpointgroupsEndpointGroupCreatePayload,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeEndpointGroup>> =>
  axios.post(
    `/endpoint_groups`,
    endpointgroupsEndpointGroupCreatePayload,
    options
  );

export const getPostEndpointGroupsMutationOptions = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof postEndpointGroups>>,
    TError,
    { data: EndpointgroupsEndpointGroupCreatePayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof postEndpointGroups>>,
  TError,
  { data: EndpointgroupsEndpointGroupCreatePayload },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof postEndpointGroups>>,
    { data: EndpointgroupsEndpointGroupCreatePayload }
  > = (props) => {
    const { data } = props ?? {};

    return postEndpointGroups(data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type PostEndpointGroupsMutationResult = NonNullable<
  Awaited<ReturnType<typeof postEndpointGroups>>
>;
export type PostEndpointGroupsMutationBody =
  EndpointgroupsEndpointGroupCreatePayload;
export type PostEndpointGroupsMutationError = AxiosError<void>;

export const usePostEndpointGroups = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof postEndpointGroups>>,
    TError,
    { data: EndpointgroupsEndpointGroupCreatePayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getPostEndpointGroupsMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * Remove an environment(endpoint) group.
 **Access policy**: administrator
 * @summary Remove an environment(endpoint) group
 */
export const endpointGroupDelete = (
  id: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.delete(`/endpoint_groups/${id}`, options);

export const getEndpointGroupDeleteMutationOptions = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof endpointGroupDelete>>,
    TError,
    { id: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof endpointGroupDelete>>,
  TError,
  { id: number },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof endpointGroupDelete>>,
    { id: number }
  > = (props) => {
    const { id } = props ?? {};

    return endpointGroupDelete(id, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type EndpointGroupDeleteMutationResult = NonNullable<
  Awaited<ReturnType<typeof endpointGroupDelete>>
>;

export type EndpointGroupDeleteMutationError = AxiosError<unknown>;

export const useEndpointGroupDelete = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof endpointGroupDelete>>,
    TError,
    { id: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getEndpointGroupDeleteMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * Retrieve details abont an environment(endpoint) group.
 **Access policy**: administrator
 * @summary Inspect an Environment(Endpoint) group
 */
export const getEndpointGroupsId = (
  id: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeEndpointGroup>> =>
  axios.get(`/endpoint_groups/${id}`, options);

export const getGetEndpointGroupsIdQueryKey = (id: number) =>
  [`/endpoint_groups/${id}`] as const;

export const getGetEndpointGroupsIdQueryOptions = <
  TData = Awaited<ReturnType<typeof getEndpointGroupsId>>,
  TError = AxiosError<void>
>(
  id: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof getEndpointGroupsId>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<
  Awaited<ReturnType<typeof getEndpointGroupsId>>,
  TError,
  TData
> & { queryKey: QueryKey } => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey = queryOptions?.queryKey ?? getGetEndpointGroupsIdQueryKey(id);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof getEndpointGroupsId>>
  > = ({ signal }) => getEndpointGroupsId(id, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!id, ...queryOptions };
};

export type GetEndpointGroupsIdQueryResult = NonNullable<
  Awaited<ReturnType<typeof getEndpointGroupsId>>
>;
export type GetEndpointGroupsIdQueryError = AxiosError<void>;

export const useGetEndpointGroupsId = <
  TData = Awaited<ReturnType<typeof getEndpointGroupsId>>,
  TError = AxiosError<void>
>(
  id: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof getEndpointGroupsId>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getGetEndpointGroupsIdQueryOptions(id, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * Update an environment(endpoint) group.
 **Access policy**: administrator
 * @summary Update an environment(endpoint) group
 */
export const endpointGroupUpdate = (
  id: number,
  endpointgroupsEndpointGroupUpdatePayload: EndpointgroupsEndpointGroupUpdatePayload,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeEndpointGroup>> =>
  axios.put(
    `/endpoint_groups/${id}`,
    endpointgroupsEndpointGroupUpdatePayload,
    options
  );

export const getEndpointGroupUpdateMutationOptions = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof endpointGroupUpdate>>,
    TError,
    { id: number; data: EndpointgroupsEndpointGroupUpdatePayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof endpointGroupUpdate>>,
  TError,
  { id: number; data: EndpointgroupsEndpointGroupUpdatePayload },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof endpointGroupUpdate>>,
    { id: number; data: EndpointgroupsEndpointGroupUpdatePayload }
  > = (props) => {
    const { id, data } = props ?? {};

    return endpointGroupUpdate(id, data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type EndpointGroupUpdateMutationResult = NonNullable<
  Awaited<ReturnType<typeof endpointGroupUpdate>>
>;
export type EndpointGroupUpdateMutationBody =
  EndpointgroupsEndpointGroupUpdatePayload;
export type EndpointGroupUpdateMutationError = AxiosError<void>;

export const useEndpointGroupUpdate = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof endpointGroupUpdate>>,
    TError,
    { id: number; data: EndpointgroupsEndpointGroupUpdatePayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getEndpointGroupUpdateMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * **Access policy**: administrator
 * @summary Removes environment(endpoint) from an environment(endpoint) group
 */
export const endpointGroupDeleteEndpoint = (
  id: number,
  endpointId: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.delete(`/endpoint_groups/${id}/endpoints/${endpointId}`, options);

export const getEndpointGroupDeleteEndpointMutationOptions = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof endpointGroupDeleteEndpoint>>,
    TError,
    { id: number; endpointId: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof endpointGroupDeleteEndpoint>>,
  TError,
  { id: number; endpointId: number },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof endpointGroupDeleteEndpoint>>,
    { id: number; endpointId: number }
  > = (props) => {
    const { id, endpointId } = props ?? {};

    return endpointGroupDeleteEndpoint(id, endpointId, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type EndpointGroupDeleteEndpointMutationResult = NonNullable<
  Awaited<ReturnType<typeof endpointGroupDeleteEndpoint>>
>;

export type EndpointGroupDeleteEndpointMutationError = AxiosError<unknown>;

export const useEndpointGroupDeleteEndpoint = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof endpointGroupDeleteEndpoint>>,
    TError,
    { id: number; endpointId: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions =
    getEndpointGroupDeleteEndpointMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * Add an environment(endpoint) to an environment(endpoint) group
 **Access policy**: administrator
 * @summary Add an environment(endpoint) to an environment(endpoint) group
 */
export const endpointGroupAddEndpoint = (
  id: number,
  endpointId: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.put(
    `/endpoint_groups/${id}/endpoints/${endpointId}`,
    undefined,
    options
  );

export const getEndpointGroupAddEndpointMutationOptions = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof endpointGroupAddEndpoint>>,
    TError,
    { id: number; endpointId: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof endpointGroupAddEndpoint>>,
  TError,
  { id: number; endpointId: number },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof endpointGroupAddEndpoint>>,
    { id: number; endpointId: number }
  > = (props) => {
    const { id, endpointId } = props ?? {};

    return endpointGroupAddEndpoint(id, endpointId, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type EndpointGroupAddEndpointMutationResult = NonNullable<
  Awaited<ReturnType<typeof endpointGroupAddEndpoint>>
>;

export type EndpointGroupAddEndpointMutationError = AxiosError<unknown>;

export const useEndpointGroupAddEndpoint = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof endpointGroupAddEndpoint>>,
    TError,
    { id: number; endpointId: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getEndpointGroupAddEndpointMutationOptions(options);

  return useMutation(mutationOptions);
};
