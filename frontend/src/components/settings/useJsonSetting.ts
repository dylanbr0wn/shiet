import { useMemo } from "react";
import { useSetSetting, useSetting } from "@/lib/api";

function parseJsonSetting<T>(raw: string | null | undefined, fallback: T): T {
  if (!raw) {
    return fallback;
  }

  try {
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

export function useJsonSetting<T>(key: string, fallback: T) {
  const query = useSetting(key);
  const mutation = useSetSetting();
  const value = useMemo(
    () => parseJsonSetting(query.data, fallback),
    [fallback, query.data],
  );

  return {
    error: query.error ?? mutation.error,
    isLoading: query.isLoading,
    isSaving: mutation.isPending,
    setValue: (nextValue: T) => {
      mutation.mutate({ key, value: JSON.stringify(nextValue) });
    },
    value,
  };
}
