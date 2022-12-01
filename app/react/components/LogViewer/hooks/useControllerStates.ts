import { useState } from 'react';

import { useUser } from '@/react/hooks/useUser';
import { useLocalStorage } from '@/react/hooks/useLocalStorage';

import { ControllerStatesInterface, TailType } from '@@/LogViewer/types';

function useStorage<T>(key: string, defaultValue: T) {
  const user = useUser();
  const fullKey = `log-viewer.user-${user.user.Id}.${key}`;
  return useLocalStorage<T>(fullKey, defaultValue);
}

export function useControllerStates() {
  const [keyword, setKeyword] = useState<string>('');
  const [filter, setFilter] = useState<boolean>(false);
  const [autoRefresh, setAutoRefresh] = useStorage<boolean>('refresh', true);
  const [since, setSince] = useStorage<number>('since', 0);
  const [tail, setTail] = useStorage<TailType>('tail', 1000);
  const [showTimestamp, setShowTimestamp] = useStorage<boolean>(
    'timestamp',
    false
  );
  const [showLineNumbers, setShowLineNumbers] = useStorage<boolean>(
    'lineNumber',
    true
  );
  const [wrapLine, setWrapLine] = useStorage<boolean>('wrapLine', true);

  const controllerStates: ControllerStatesInterface = {
    keyword,
    setKeyword,
    filter,
    setFilter,
    autoRefresh,
    setAutoRefresh,
    since,
    setSince,
    tail,
    setTail,
    showTimestamp,
    setShowTimestamp,
    wrapLine,
    setWrapLine,
    showLineNumbers,
    setShowLineNumbers,
  };

  return { controllerStates };
}
