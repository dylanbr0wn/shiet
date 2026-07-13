import { anyLoading, firstError } from "./useSchedulePage.helpers";

interface BuildSchedulePageStatusArgs {
  loadingFlags: boolean[];
  errors: unknown[];
  eventsCount: number;
  timeEntriesCount: number;
  categoriesCount: number;
  reviewDecisionsCount: number;
}

export function buildSchedulePageStatus({
  loadingFlags,
  errors,
  eventsCount,
  timeEntriesCount,
  categoriesCount,
  reviewDecisionsCount,
}: BuildSchedulePageStatusArgs) {
  return {
    isBackendLoading: anyLoading(loadingFlags),
    backendError: firstError(errors),
    counts: {
      events: eventsCount,
      timeEntries: timeEntriesCount,
      categories: categoriesCount,
      reviewDecisions: reviewDecisionsCount,
    },
  };
}
