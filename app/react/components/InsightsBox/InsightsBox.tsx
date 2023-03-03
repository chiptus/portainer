import clsx from 'clsx';
import { Lightbulb, X } from 'lucide-react';
import { ReactNode, useMemo } from 'react';
import sanitize from 'sanitize-html';
import { useStore } from 'zustand';

import { Button } from '@@/buttons';

import { insightStore } from './insights-store';

export type Props = {
  header: string;
  content: ReactNode;
  setHtmlContent?: boolean;
  insightCloseId?: string; // set if you want to be able to close the box and not show it again
};

export function InsightsBox({
  header,
  content,
  setHtmlContent,
  insightCloseId,
}: Props) {
  // allow to close the box and not show it again in local storage with zustand
  const { addInsightIDClosed, isClosed } = useStore(insightStore);
  const isInsightClosed = isClosed(insightCloseId);

  // allow angular views to set html messages for the insights box
  const htmlContent = useMemo(() => {
    if (setHtmlContent && typeof content === 'string') {
      // eslint-disable-next-line react/no-danger
      return <div dangerouslySetInnerHTML={{ __html: sanitize(content) }} />;
    }
    return null;
  }, [setHtmlContent, content]);

  if (isInsightClosed) {
    return null;
  }

  return (
    <div className="relative flex w-full gap-1 rounded-lg bg-gray-modern-3 p-4 text-sm th-highcontrast:bg-legacy-grey-3 th-dark:bg-legacy-grey-3">
      <div className="shrink-0">
        <Lightbulb className="h-4 text-warning-7 th-highcontrast:text-warning-6 th-dark:text-warning-6" />
      </div>
      <div>
        <p className={clsx('mb-2 font-bold', insightCloseId && 'pr-4')}>
          {header}
        </p>
        <div>{htmlContent || content}</div>
      </div>
      {insightCloseId && (
        <Button
          icon={X}
          className="absolute top-2 right-2 flex !text-gray-7 hover:!text-gray-8 th-highcontrast:!text-gray-6 th-highcontrast:hover:!text-gray-5 th-dark:!text-gray-6 th-dark:hover:!text-gray-5"
          color="link"
          size="medium"
          onClick={() => addInsightIDClosed(insightCloseId)}
        />
      )}
    </div>
  );
}