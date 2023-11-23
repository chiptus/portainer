import { Option } from '@/react/portainer/access-control/AccessManagement/PorAccessManagementUsersSelector';

export interface FormValues {
  selectedUsersAndTeams: Option[];
}

export interface TableEntry {
  id: number;
  name: string;
  type: 'user' | 'team';
}
