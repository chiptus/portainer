import { ChevronUp, ChevronDown } from 'lucide-react';

import { AutomationTestingProps } from '@/types';

import { Button } from '@@/buttons';

export interface SearchStatusInterface {
  totalKeywords: number;
  focusedKeywordIndex: number;
  nextKeyword: () => void;
  previousKeyword: () => void;
}

interface Props extends AutomationTestingProps {
  searchIndicator: SearchStatusInterface;
}

export function SearchStatus({
  searchIndicator: {
    totalKeywords,
    focusedKeywordIndex,
    nextKeyword,
    previousKeyword,
  },
}: Props) {
  return (
    <div className="ml-2 flex shrink-0">
      <div className="text-xs">
        <span>
          {totalKeywords && focusedKeywordIndex !== undefined
            ? `${focusedKeywordIndex + 1} / `
            : ''}
        </span>
        <span>{totalKeywords}</span>
      </div>
      <Button
        onClick={previousKeyword}
        icon={ChevronUp}
        color="none"
        disabled={!totalKeywords}
      />
      <Button
        onClick={nextKeyword}
        icon={ChevronDown}
        color="none"
        disabled={!totalKeywords}
      />
    </div>
  );
}
