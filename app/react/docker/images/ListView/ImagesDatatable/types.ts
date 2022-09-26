import {
  PaginationTableSettings,
  RefreshableTableSettings,
  SortableTableSettings,
} from '@/react/components/datatables/types';

export interface TableSettings
  extends SortableTableSettings,
    PaginationTableSettings,
    RefreshableTableSettings {}
