import clsx from 'clsx';

import { removeTrailingGitExtension } from '@/react/edge/edge-stacks/utils';

interface Props {
  url: string;
  configFilePath: string;
  additionalFiles?: string[];
  className: string;
  type: string;
  commitHash: string;
}

export function InfoPanel({
  url,
  configFilePath,
  additionalFiles = [],
  className,
  type,
  commitHash,
}: Props) {
  return (
    <div className={clsx('form-group', className)}>
      <div className="col-sm-12">
        <p>
          This {type} was deployed from the git repository <code>{url}</code>{' '}
          and the current version deployed is{' '}
          <a
            href={`${removeTrailingGitExtension(url)}/commit/${commitHash}`}
            target="_blank"
            rel="noreferrer"
          >
            {commitHash.slice(0, 7).toString()}
          </a>
        </p>
        <p>
          Update
          <code>{[configFilePath, ...additionalFiles].join(', ')}</code>
          in git and pull from here to update the {type}.
        </p>
      </div>
    </div>
  );
}
