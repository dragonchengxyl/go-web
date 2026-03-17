'use client';

import { createContext, useCallback, useContext, useMemo, useState } from 'react';

export interface AssistantPageContextData {
  path?: string;
  kind: string;
  title: string;
  summary?: string;
  prompt_hints?: string[];
  fields?: Record<string, string>;
}

interface AssistantPageContextValue {
  pageContext: AssistantPageContextData | null;
  setPageContext: (value: AssistantPageContextData | null) => void;
  clearPageContext: () => void;
}

const AssistantPageContext = createContext<AssistantPageContextValue>({
  pageContext: null,
  setPageContext: () => {},
  clearPageContext: () => {},
});

export function AssistantPageContextProvider({ children }: { children: React.ReactNode }) {
  const [pageContext, setPageContextState] = useState<AssistantPageContextData | null>(null);

  const setPageContext = useCallback((value: AssistantPageContextData | null) => {
    setPageContextState(value);
  }, []);

  const clearPageContext = useCallback(() => {
    setPageContextState(null);
  }, []);

  const value = useMemo(
    () => ({
      pageContext,
      setPageContext,
      clearPageContext,
    }),
    [pageContext, setPageContext, clearPageContext],
  );

  return (
    <AssistantPageContext.Provider value={value}>
      {children}
    </AssistantPageContext.Provider>
  );
}

export function useAssistantPageContext() {
  return useContext(AssistantPageContext);
}
