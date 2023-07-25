import {
  SortOptions,
  SortQuery,
  SortQueryParams,
  makeIsSortTypeFunc,
} from '@/react/common/api/sort.types';
import {
  SearchQuery,
  SearchQueryParams,
} from '@/react/common/api/search.types';
import {
  PaginationQuery,
  PaginationQueryParams,
} from '@/react/common/api/pagination.types';

const sortOptions: SortOptions = ['name'] as const;
export const isSortType = makeIsSortTypeFunc(sortOptions);

export type Query = SearchQuery &
  PaginationQuery &
  SortQuery<typeof sortOptions>;

export type QueryParams = SearchQueryParams &
  PaginationQueryParams &
  SortQueryParams<typeof sortOptions>;
