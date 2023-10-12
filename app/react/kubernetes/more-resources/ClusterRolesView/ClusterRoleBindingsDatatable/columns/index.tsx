import { name } from './name';
import { kind } from './kind';
import { created } from './created';
import { subjects } from './subjects';
import { subjectKind } from './subjectKind';
import { subjectName } from './subjectName';
import { subjectNamespace } from './subjectNamespace';

export const columns = [
  name,
  kind,
  subjects,
  subjectKind,
  subjectName,
  subjectNamespace,
  created,
];
