import { createContext, PropsWithChildren, useContext } from 'react';

const SetSizeContext = createContext<
  null | ((index: number, size: number) => void)
>(null);

export function useSetSizeContext() {
  const context = useContext(SetSizeContext);
  if (!context) {
    throw new Error('useSizeContext must be used within a SizeContext');
  }

  return context;
}

export function SetSizeProvider({
  children,
  setSize,
}: PropsWithChildren<{ setSize: (index: number, size: number) => void }>) {
  return (
    <SetSizeContext.Provider value={setSize}>
      {children}
    </SetSizeContext.Provider>
  );
}
