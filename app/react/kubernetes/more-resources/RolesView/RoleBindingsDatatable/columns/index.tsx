import { name } from './name';
import { roleKind } from './roleKind';
import { roleName } from './roleName';
import { subjectKind } from './subjectKind';
import { subjectName } from './subjectName';
import { subjectNamespace } from './subjectNamespace';
import { created } from './created';
import { subjects } from './subjects';

export const columns = [
  name,
  roleKind,
  roleName,
  subjects,
  subjectKind,
  subjectName,
  subjectNamespace,
  created,
];
