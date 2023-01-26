import { ChevronDown } from 'lucide-react';

import { Button } from '@@/buttons';

interface Props {
  visible: boolean;
  onClick: () => void;
}

export function BottomButton({ visible, onClick }: Props) {
  if (!visible) {
    return null;
  }

  return (
    <div className="absolute bottom-5 w-full text-center">
      <Button
        onClick={onClick}
        className="rounded-3xl px-5 py-1"
        icon={ChevronDown}
      >
        <span>Jump to Bottom</span>
      </Button>
    </div>
  );
}
