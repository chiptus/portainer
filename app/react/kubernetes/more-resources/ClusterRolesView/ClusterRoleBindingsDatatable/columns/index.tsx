import { name } from './name';
import { kind } from './kind';
import { created } from './created';
import { subjectKind } from './subjectKind';
import { subjectName } from './subjectName';
import { subjectNamespace } from './subjectNamespace';

export const columns = [
  name,
  kind,
  subjectKind,
  subjectName,
  subjectNamespace,
  created,
];
