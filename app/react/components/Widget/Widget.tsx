import clsx from 'clsx';
import { createContext, PropsWithChildren, Ref, useContext } from 'react';

const Context = createContext<null | boolean>(null);
Context.displayName = 'WidgetContext';

export function useWidgetContext() {
  const context = useContext(Context);

  if (context == null) {
    throw new Error('Should be inside a Widget component');
  }
}

export function Widget({
  children,
  className,
  mRef,
}: PropsWithChildren<{ className?: string; mRef?: Ref<HTMLDivElement> }>) {
  return (
    <Context.Provider value>
      <div className={clsx('widget', className)} ref={mRef}>
        {children}
      </div>
    </Context.Provider>
  );
}
