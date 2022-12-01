import clsx from 'clsx';
import { ChevronDown } from 'lucide-react';

import { Button } from '@@/buttons';
import { Icon } from '@@/Icon';

import './BottomButton.css';

interface Props {
  visible: boolean;
  onClick: () => void;
}

export function BottomButton({ visible, onClick }: Props) {
  return (
    <div className="bottom-button">
      <Button
        onClick={onClick}
        className={clsx('vertical-center', 'button', { visible })}
      >
        <Icon icon={ChevronDown} />
        <span>Jump to Bottom</span>
      </Button>
    </div>
  );
}
