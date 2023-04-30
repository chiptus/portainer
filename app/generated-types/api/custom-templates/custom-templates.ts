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
  PortainereeCustomTemplate,
  CustomTemplateListParams,
  CustomtemplatesCustomTemplateUpdatePayload,
  CustomtemplatesFileResponse,
  CustomTemplateCreateFileBody,
  CustomtemplatesCustomTemplateFromGitRepositoryPayload,
  CustomtemplatesCustomTemplateFromFileContentPayload,
} from '../portainerEEAPI.schemas';

/**
 * List available custom templates.
 **Access policy**: authenticated
 * @summary List available custom templates
 */
export const customTemplateList = (
  params: CustomTemplateListParams,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeCustomTemplate[]>> =>
  axios.get(`/custom_templates`, {
    ...options,
    params: { ...params, ...options?.params },
  });

export const getCustomTemplateListQueryKey = (
  params: CustomTemplateListParams
) => [`/custom_templates`, ...(params ? [params] : [])] as const;

export const getCustomTemplateListQueryOptions = <
  TData = Awaited<ReturnType<typeof customTemplateList>>,
  TError = AxiosError<void>
>(
  params: CustomTemplateListParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof customTemplateList>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<
  Awaited<ReturnType<typeof customTemplateList>>,
  TError,
  TData
> & { queryKey: QueryKey } => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getCustomTemplateListQueryKey(params);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof customTemplateList>>
  > = ({ signal }) => customTemplateList(params, { signal, ...axiosOptions });

  return { queryKey, queryFn, ...queryOptions };
};

export type CustomTemplateListQueryResult = NonNullable<
  Awaited<ReturnType<typeof customTemplateList>>
>;
export type CustomTemplateListQueryError = AxiosError<void>;

export const useCustomTemplateList = <
  TData = Awaited<ReturnType<typeof customTemplateList>>,
  TError = AxiosError<void>
>(
  params: CustomTemplateListParams,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof customTemplateList>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getCustomTemplateListQueryOptions(params, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * Remove a template.
 **Access policy**: authenticated
 * @summary Remove a template
 */
export const customTemplateDelete = (
  id: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<void>> =>
  axios.delete(`/custom_templates/${id}`, options);

export const getCustomTemplateDeleteMutationOptions = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateDelete>>,
    TError,
    { id: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof customTemplateDelete>>,
  TError,
  { id: number },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof customTemplateDelete>>,
    { id: number }
  > = (props) => {
    const { id } = props ?? {};

    return customTemplateDelete(id, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type CustomTemplateDeleteMutationResult = NonNullable<
  Awaited<ReturnType<typeof customTemplateDelete>>
>;

export type CustomTemplateDeleteMutationError = AxiosError<unknown>;

export const useCustomTemplateDelete = <
  TError = AxiosError<unknown>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateDelete>>,
    TError,
    { id: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getCustomTemplateDeleteMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * Retrieve details about a template.
 **Access policy**: authenticated
 * @summary Inspect a custom template
 */
export const customTemplateInspect = (
  id: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeCustomTemplate>> =>
  axios.get(`/custom_templates/${id}`, options);

export const getCustomTemplateInspectQueryKey = (id: number) =>
  [`/custom_templates/${id}`] as const;

export const getCustomTemplateInspectQueryOptions = <
  TData = Awaited<ReturnType<typeof customTemplateInspect>>,
  TError = AxiosError<void>
>(
  id: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof customTemplateInspect>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<
  Awaited<ReturnType<typeof customTemplateInspect>>,
  TError,
  TData
> & { queryKey: QueryKey } => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getCustomTemplateInspectQueryKey(id);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof customTemplateInspect>>
  > = ({ signal }) => customTemplateInspect(id, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!id, ...queryOptions };
};

export type CustomTemplateInspectQueryResult = NonNullable<
  Awaited<ReturnType<typeof customTemplateInspect>>
>;
export type CustomTemplateInspectQueryError = AxiosError<void>;

export const useCustomTemplateInspect = <
  TData = Awaited<ReturnType<typeof customTemplateInspect>>,
  TError = AxiosError<void>
>(
  id: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof customTemplateInspect>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getCustomTemplateInspectQueryOptions(id, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * Update a template.
 **Access policy**: authenticated
 * @summary Update a template
 */
export const customTemplateUpdate = (
  id: number,
  customtemplatesCustomTemplateUpdatePayload: CustomtemplatesCustomTemplateUpdatePayload,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeCustomTemplate>> =>
  axios.put(
    `/custom_templates/${id}`,
    customtemplatesCustomTemplateUpdatePayload,
    options
  );

export const getCustomTemplateUpdateMutationOptions = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateUpdate>>,
    TError,
    { id: number; data: CustomtemplatesCustomTemplateUpdatePayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof customTemplateUpdate>>,
  TError,
  { id: number; data: CustomtemplatesCustomTemplateUpdatePayload },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof customTemplateUpdate>>,
    { id: number; data: CustomtemplatesCustomTemplateUpdatePayload }
  > = (props) => {
    const { id, data } = props ?? {};

    return customTemplateUpdate(id, data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type CustomTemplateUpdateMutationResult = NonNullable<
  Awaited<ReturnType<typeof customTemplateUpdate>>
>;
export type CustomTemplateUpdateMutationBody =
  CustomtemplatesCustomTemplateUpdatePayload;
export type CustomTemplateUpdateMutationError = AxiosError<void>;

export const useCustomTemplateUpdate = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateUpdate>>,
    TError,
    { id: number; data: CustomtemplatesCustomTemplateUpdatePayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getCustomTemplateUpdateMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * Retrieve the content of the Stack file for the specified custom template
 **Access policy**: authenticated
 * @summary Get Template stack file content.
 */
export const customTemplateFile = (
  id: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<CustomtemplatesFileResponse>> =>
  axios.get(`/custom_templates/${id}/file`, options);

export const getCustomTemplateFileQueryKey = (id: number) =>
  [`/custom_templates/${id}/file`] as const;

export const getCustomTemplateFileQueryOptions = <
  TData = Awaited<ReturnType<typeof customTemplateFile>>,
  TError = AxiosError<void>
>(
  id: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof customTemplateFile>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryOptions<
  Awaited<ReturnType<typeof customTemplateFile>>,
  TError,
  TData
> & { queryKey: QueryKey } => {
  const { query: queryOptions, axios: axiosOptions } = options ?? {};

  const queryKey = queryOptions?.queryKey ?? getCustomTemplateFileQueryKey(id);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof customTemplateFile>>
  > = ({ signal }) => customTemplateFile(id, { signal, ...axiosOptions });

  return { queryKey, queryFn, enabled: !!id, ...queryOptions };
};

export type CustomTemplateFileQueryResult = NonNullable<
  Awaited<ReturnType<typeof customTemplateFile>>
>;
export type CustomTemplateFileQueryError = AxiosError<void>;

export const useCustomTemplateFile = <
  TData = Awaited<ReturnType<typeof customTemplateFile>>,
  TError = AxiosError<void>
>(
  id: number,
  options?: {
    query?: UseQueryOptions<
      Awaited<ReturnType<typeof customTemplateFile>>,
      TError,
      TData
    >;
    axios?: AxiosRequestConfig;
  }
): UseQueryResult<TData, TError> & { queryKey: QueryKey } => {
  const queryOptions = getCustomTemplateFileQueryOptions(id, options);

  const query = useQuery(queryOptions) as UseQueryResult<TData, TError> & {
    queryKey: QueryKey;
  };

  query.queryKey = queryOptions.queryKey;

  return query;
};

/**
 * Retrieve details about a template created from git repository method.
 **Access policy**: authenticated
 * @summary Fetch the latest config file content based on custom template's git repository configuration
 */
export const customTemplateGitFetch = (
  id: number,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<CustomtemplatesFileResponse>> =>
  axios.put(`/custom_templates/${id}/git_fetch`, undefined, options);

export const getCustomTemplateGitFetchMutationOptions = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateGitFetch>>,
    TError,
    { id: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof customTemplateGitFetch>>,
  TError,
  { id: number },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof customTemplateGitFetch>>,
    { id: number }
  > = (props) => {
    const { id } = props ?? {};

    return customTemplateGitFetch(id, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type CustomTemplateGitFetchMutationResult = NonNullable<
  Awaited<ReturnType<typeof customTemplateGitFetch>>
>;

export type CustomTemplateGitFetchMutationError = AxiosError<void>;

export const useCustomTemplateGitFetch = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateGitFetch>>,
    TError,
    { id: number },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getCustomTemplateGitFetchMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * Create a custom template.
 **Access policy**: authenticated
 * @summary Create a custom template
 */
export const customTemplateCreateFile = (
  customTemplateCreateFileBody: CustomTemplateCreateFileBody,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeCustomTemplate>> => {
  const formData = new FormData();
  if (customTemplateCreateFileBody.Title !== undefined) {
    formData.append('Title', customTemplateCreateFileBody.Title);
  }
  if (customTemplateCreateFileBody.Description !== undefined) {
    formData.append('Description', customTemplateCreateFileBody.Description);
  }
  if (customTemplateCreateFileBody.Note !== undefined) {
    formData.append('Note', customTemplateCreateFileBody.Note);
  }
  if (customTemplateCreateFileBody.Platform !== undefined) {
    formData.append(
      'Platform',
      customTemplateCreateFileBody.Platform.toString()
    );
  }
  if (customTemplateCreateFileBody.Type !== undefined) {
    formData.append('Type', customTemplateCreateFileBody.Type.toString());
  }
  if (customTemplateCreateFileBody.file !== undefined) {
    formData.append('file', customTemplateCreateFileBody.file);
  }

  return axios.post(`/custom_templates/file`, formData, options);
};

export const getCustomTemplateCreateFileMutationOptions = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateCreateFile>>,
    TError,
    { data: CustomTemplateCreateFileBody },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof customTemplateCreateFile>>,
  TError,
  { data: CustomTemplateCreateFileBody },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof customTemplateCreateFile>>,
    { data: CustomTemplateCreateFileBody }
  > = (props) => {
    const { data } = props ?? {};

    return customTemplateCreateFile(data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type CustomTemplateCreateFileMutationResult = NonNullable<
  Awaited<ReturnType<typeof customTemplateCreateFile>>
>;
export type CustomTemplateCreateFileMutationBody = CustomTemplateCreateFileBody;
export type CustomTemplateCreateFileMutationError = AxiosError<void>;

export const useCustomTemplateCreateFile = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateCreateFile>>,
    TError,
    { data: CustomTemplateCreateFileBody },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getCustomTemplateCreateFileMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * Create a custom template.
 **Access policy**: authenticated
 * @summary Create a custom template
 */
export const customTemplateCreateRepository = (
  customtemplatesCustomTemplateFromGitRepositoryPayload: CustomtemplatesCustomTemplateFromGitRepositoryPayload,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeCustomTemplate>> =>
  axios.post(
    `/custom_templates/repository`,
    customtemplatesCustomTemplateFromGitRepositoryPayload,
    options
  );

export const getCustomTemplateCreateRepositoryMutationOptions = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateCreateRepository>>,
    TError,
    { data: CustomtemplatesCustomTemplateFromGitRepositoryPayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof customTemplateCreateRepository>>,
  TError,
  { data: CustomtemplatesCustomTemplateFromGitRepositoryPayload },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof customTemplateCreateRepository>>,
    { data: CustomtemplatesCustomTemplateFromGitRepositoryPayload }
  > = (props) => {
    const { data } = props ?? {};

    return customTemplateCreateRepository(data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type CustomTemplateCreateRepositoryMutationResult = NonNullable<
  Awaited<ReturnType<typeof customTemplateCreateRepository>>
>;
export type CustomTemplateCreateRepositoryMutationBody =
  CustomtemplatesCustomTemplateFromGitRepositoryPayload;
export type CustomTemplateCreateRepositoryMutationError = AxiosError<void>;

export const useCustomTemplateCreateRepository = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateCreateRepository>>,
    TError,
    { data: CustomtemplatesCustomTemplateFromGitRepositoryPayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions =
    getCustomTemplateCreateRepositoryMutationOptions(options);

  return useMutation(mutationOptions);
};
/**
 * Create a custom template.
 **Access policy**: authenticated
 * @summary Create a custom template
 */
export const customTemplateCreateString = (
  customtemplatesCustomTemplateFromFileContentPayload: CustomtemplatesCustomTemplateFromFileContentPayload,
  options?: AxiosRequestConfig
): Promise<AxiosResponse<PortainereeCustomTemplate>> =>
  axios.post(
    `/custom_templates/string`,
    customtemplatesCustomTemplateFromFileContentPayload,
    options
  );

export const getCustomTemplateCreateStringMutationOptions = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateCreateString>>,
    TError,
    { data: CustomtemplatesCustomTemplateFromFileContentPayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}): UseMutationOptions<
  Awaited<ReturnType<typeof customTemplateCreateString>>,
  TError,
  { data: CustomtemplatesCustomTemplateFromFileContentPayload },
  TContext
> => {
  const { mutation: mutationOptions, axios: axiosOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<ReturnType<typeof customTemplateCreateString>>,
    { data: CustomtemplatesCustomTemplateFromFileContentPayload }
  > = (props) => {
    const { data } = props ?? {};

    return customTemplateCreateString(data, axiosOptions);
  };

  return { mutationFn, ...mutationOptions };
};

export type CustomTemplateCreateStringMutationResult = NonNullable<
  Awaited<ReturnType<typeof customTemplateCreateString>>
>;
export type CustomTemplateCreateStringMutationBody =
  CustomtemplatesCustomTemplateFromFileContentPayload;
export type CustomTemplateCreateStringMutationError = AxiosError<void>;

export const useCustomTemplateCreateString = <
  TError = AxiosError<void>,
  TContext = unknown
>(options?: {
  mutation?: UseMutationOptions<
    Awaited<ReturnType<typeof customTemplateCreateString>>,
    TError,
    { data: CustomtemplatesCustomTemplateFromFileContentPayload },
    TContext
  >;
  axios?: AxiosRequestConfig;
}) => {
  const mutationOptions = getCustomTemplateCreateStringMutationOptions(options);

  return useMutation(mutationOptions);
};
