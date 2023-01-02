import { createContext, useContext } from 'react';

import { LogViewerContextInterface } from './types';

const LogViewerContext = createContext<null | LogViewerContextInterface>(null);

export function useLogViewerContext() {
  const context = useContext(LogViewerContext);
  if (context === null) {
    throw new Error(
      'useLogViewerContext must be used within a LogViewerProvider'
    );
  }
  return context;
}

export const LogViewerProvider = LogViewerContext.Provider;
