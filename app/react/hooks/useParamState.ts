import { useCurrentStateAndParams, useRouter } from '@uirouter/react';
import { useState } from 'react';

export function useParamState<T>(
  param: string,
  parseParam: (param: string | undefined) => T | undefined
) {
  const {
    params: { [param]: paramValue },
  } = useCurrentStateAndParams();

  const router = useRouter();
  const state = parseParam(paramValue);
  const [fastValue, setFastValue] = useState(state);

  return [
    fastValue,
    (value: T | undefined) => {
      setFastValue(value);
      router.stateService.go('.', { [param]: value });
    },
  ] as const;
}
