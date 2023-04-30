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
  LicensesDeleteResponse,
  LicensesDeletePayload,
  LiblicensePortainerLicense,
  LicensesAttachResponse,
  LicensesAttachPayload,
  LicensesLicenseInfo,
} from '../portainerEEAPI.schemas';

/**
 * **Access policy**: administrator
 * @summary delete license from portainer instance
 */
export const licensesDelete = (
  licensesDeletePayload: LicensesDeletePayload,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<LicensesDeleteResponse>> =>
  axios.delete(`/licenses`, { data: licensesDeletePayload, ...options });

export const getLicensesDeleteMutationOptions = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof licensesDelete>>,
    TError,
    { data: LicensesDeletePayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof licensesDelete>>,
  TError,
  { data: LicensesDeletePayload },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof licensesDelete>>,
    { data: LicensesDeletePayload }
  > = (props) => {
    const { data } = props ?? {};

    return licensesDelete(data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type LicensesDeleteMutationResult = NonNullable<
  Awaited<ReturnType<typeof licensesDelete>>
>;
export type LicensesDeleteMutationBody = LicensesDeletePayload;
export type LicensesDeleteMutationError = AxiosError<unknown>;

export const useLicensesDelete = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof licensesDelete>>,
    TError,
    { data: LicensesDeletePayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getLicensesDeleteMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * **Access policy**: administrator
 * @summary fetches the list of licenses on Portainer
 */
export const licensesList = (
  options?: AxiosRequestConfig
): Promise<AxiosResponse<LiblicensePortainerLicense[]>> =>
  axios.get(`/licenses`, options);

export const getLicensesListQueryKey = () => [`/licenses`] as const;

export const getLicensesListQueryOptions = <
  TData = Awaited<ReturnType<typeof licensesList>>,
  TError = AxiosError<unknown>
>(options?: {
  query?: UseQueryOptions<
    Awaited<ReturnType<typeof licensesList>>,
    TError,
    TData
  >;
  axios?: AxiosRequestConfig;
}): UseQueryOptions<Awaited<ReturnType<typeof licensesList>>, TError, TData> & {
  queryKey: QueryKey;
} => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey = queryOptions?.queryKey ?? getLicensesListQueryKey();

  const queryFn: QueryFunction<Awaited<ReturnType<typeof licensesList>>> = ({
    signal,
  }) => licensesList({ signal, ...axiosOptions });

  return { queryKey, queryFn, ...queryOptions };
};

export type LicensesListQueryResult = NonNullable<
  Awaited<ReturnType<typeof licensesList>>
>;
export type LicensesListQueryError = AxiosError<unknown>;

export const useLicensesList = <
  TData = Awaited<ReturnType<typeof licensesList>>,
  TError = AxiosError<unknown>
>(options?: {
  query?: UseQueryOptions<
    Awaited<ReturnType<typeof licensesList>>,
    TError,
    TData
  >;
  axios?: AxiosRequestConfig;
}): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getLicensesListQueryOptions(options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * **Access policy**: administrator
 * @summary attaches a list of licenses to Portainer
 */
export const licensesAttach = (
  licensesAttachPayload: LicensesAttachPayload,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<LicensesAttachResponse>> =>
  axios.post(`/licenses`, licensesAttachPayload, options);

export const getLicensesAttachMutationOptions = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof licensesAttach>>,
    TError,
    { data: LicensesAttachPayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof licensesAttach>>,
  TError,
  { data: LicensesAttachPayload },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof licensesAttach>>,
    { data: LicensesAttachPayload }
  > = (props) => {
    const { data } = props ?? {};

    return licensesAttach(data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type LicensesAttachMutationResult = NonNullable<
  Awaited<ReturnType<typeof licensesAttach>>
>;
export type LicensesAttachMutationBody = LicensesAttachPayload;
export type LicensesAttachMutationError = AxiosError<unknown>;

export const useLicensesAttach = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof licensesAttach>>,
    TError,
    { data: LicensesAttachPayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getLicensesAttachMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * **Access policy**: administrator
 * @summary summarizes licenses on Portainer
 */
export const licensesInfo = (
  options?: AxiosRequestConfig
): Promise<AxiosResponse<LicensesLicenseInfo>> =>
  axios.get(`/licenses/info`, options);

export const getLicensesInfoQueryKey = () => [`/licenses/info`] as const;

export const getLicensesInfoQueryOptions = <
  TData = Awaited<ReturnType<typeof licensesInfo>>,
  TError = AxiosError<unknown>
>(options?: {
  query?: UseQueryOptions<
    Awaited<ReturnType<typeof licensesInfo>>,
    TError,
    TData
  >;
  axios?: AxiosRequestConfig;
}): UseQueryOptions<Awaited<ReturnType<typeof licensesInfo>>, TError, TData> & {
  queryKey: QueryKey;
} => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey = queryOptions?.queryKey ?? getLicensesInfoQueryKey();

  const queryFn: QueryFunction<Awaited<ReturnType<typeof licensesInfo>>> = ({
    signal,
  }) => licensesInfo({ signal, ...axiosOptions });

  return { queryKey, queryFn, ...queryOptions };
};

export type LicensesInfoQueryResult = NonNullable<
  Awaited<ReturnType<typeof licensesInfo>>
>;
export type LicensesInfoQueryError = AxiosError<unknown>;

export const useLicensesInfo = <
  TData = Awaited<ReturnType<typeof licensesInfo>>,
  TError = AxiosError<unknown>
>(options?: {
  query?: UseQueryOptions<
    Awaited<ReturnType<typeof licensesInfo>>,
    TError,
    TData
  >;
  axios?: AxiosRequestConfig;
}): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getLicensesInfoQueryOptions(options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};
