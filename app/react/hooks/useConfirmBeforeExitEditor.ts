// import { useTransitionHook } from '@uirouter/react';
import { useCallback, useEffect } from 'react';
import {
  HookMatchCriteria,
  HookRegOptions,
  TransitionHookFn,
  TransitionService,
  useRouter,
} from '@uirouter/react';

import { confirmWebEditorDiscard } from '@@/modals/confirm';

export function useConfirmBeforeExit(
  originalContent: string,
  currentContent: string,
  skipCheck = false
) {
  const didValueChange =
    removeWhitespace(currentContent) !== removeWhitespace(originalContent);

  const beforeUnloadHandler = useCallback(async () => {
    if (!didValueChange || skipCheck) {
      return undefined;
    }

    return '';
  }, [didValueChange, skipCheck]);

  useEffect(() => {
    window.addEventListener('beforeunload', beforeUnloadHandler);

    return () => {
      window.removeEventListener('beforeunload', beforeUnloadHandler);
    };
  }, [beforeUnloadHandler]);

  const transitionHook = useCallback(
    () => confirmExit(didValueChange, skipCheck),
    [didValueChange, skipCheck]
  );

  useTransitionHook('onBefore', {}, transitionHook, {});
}

type TransitionServiceKeys = keyof TransitionService;

type HookName = Extract<
  TransitionServiceKeys,
  'onBefore' | 'onStart' | 'onSuccess' | 'onError' | 'onSuccess' | 'onFinish'
>;

/**
 * This implemented useTransitionHook of uirouter, but with de-registering the callback if it changes,
 * this allows for the callback to rely on react changing state (using useCallback)
 *
 * @param hookName
 * @param criteria
 * @param callback should be a stable callback (using useCallback)
 * @param options
 */
function useTransitionHook(
  hookName: HookName,
  criteria: HookMatchCriteria,
  callback: TransitionHookFn,
  options?: HookRegOptions
) {
  const router = useRouter();
  const { transitionService } = router;

  useEffect(() => {
    if (!criteria) {
      return () => {};
    }

    const deRegister = transitionService[hookName](criteria, callback, options);
    return () => deRegister();
  }, [transitionService, criteria, callback, options, hookName]);
}

function confirmExit(didValueChange: boolean, skipCheck: boolean) {
  if (skipCheck || !didValueChange) {
    return true;
  }

  return confirmWebEditorDiscard();
}

function removeWhitespace(text: string) {
  return text.replace(/(\r\n|\n|\r)/gm, '');
}
