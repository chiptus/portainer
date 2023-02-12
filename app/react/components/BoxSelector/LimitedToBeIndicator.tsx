import { HelpCircle } from 'lucide-react';
import clsx from 'clsx';

import { TooltipWithChildren } from '@@/Tip/TooltipWithChildren';

interface Props {
  tooltipId: string;
}

export function LimitedToBeIndicator({ tooltipId }: Props) {
  return (
    <div className="absolute left-0 top-0 w-full">
      <div className="mx-auto flex max-w-fit items-center gap-1 rounded-b-lg bg-warning-4 py-1 px-3 text-sm">
        <span className="text-warning-9">BE Feature</span>
        <TooltipWithChildren
          position="bottom"
          className={clsx(tooltipId, 'portainer-tooltip')}
          heading="Business Edition feature."
          message="This feature is currently limited to Business Edition users only."
        >
          <HelpCircle className="ml-1" aria-hidden="true" />
        </TooltipWithChildren>
      </div>
    </div>
  );
}
