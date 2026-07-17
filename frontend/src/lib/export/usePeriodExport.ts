import { useMemo } from "react";
import {
  useCategories,
  useEventCategoryOverlays,
  useEvents,
  useExpectedTimeForRange,
  useTimeEntries,
  useTzSegments,
  type Period,
} from "@/lib/api";
import {
  buildEventCategoryOverlayMap,
  eventToSchedulerItem,
  timeEntryToSchedulerItem,
  resolveEventCategoryId,
} from "@/lib/schedule";
import { buildPeriodExportSummary } from "./summary";

export function usePeriodExport(period: Period | null | undefined) {
  const periodId = period?.id;
  const eventsQuery = useEvents(periodId);
  const eventCategoryOverlaysQuery = useEventCategoryOverlays(periodId);
  const timeEntriesQuery = useTimeEntries(periodId);
  const tzSegmentsQuery = useTzSegments(periodId);
  const categoriesQuery = useCategories();
  const expectedQuery = useExpectedTimeForRange(period?.startDate, period?.endDate);

  const categoriesById = useMemo(() => {
    return new Map(
      (categoriesQuery.data ?? []).map((category) => [category.id, category]),
    );
  }, [categoriesQuery.data]);

  const items = useMemo(() => {
    const events = eventsQuery.data ?? [];
    const timeEntries = timeEntriesQuery.data ?? [];
    const tzSegments = tzSegmentsQuery.data ?? [];
    const overlaysByKey = buildEventCategoryOverlayMap(
      eventCategoryOverlaysQuery.data ?? [],
    );

    return [
      ...events
        .map((event) =>
          eventToSchedulerItem(
            event,
            tzSegments,
            categoriesById,
            resolveEventCategoryId(event, overlaysByKey),
          ),
        )
        .filter((item): item is NonNullable<typeof item> => item !== null),
      ...timeEntries
        .map((timeEntry) =>
          timeEntryToSchedulerItem(timeEntry, categoriesById, tzSegments),
        )
        .filter((item): item is NonNullable<typeof item> => item !== null),
    ];
  }, [
    categoriesById,
    eventCategoryOverlaysQuery.data,
    eventsQuery.data,
    timeEntriesQuery.data,
    tzSegmentsQuery.data,
  ]);

  const summary = useMemo(() => {
    if (!period || !expectedQuery.isSuccess) {
      return null;
    }

    return buildPeriodExportSummary(items, period, {
      categories: categoriesQuery.data ?? [],
      expectedDays: expectedQuery.data,
    });
  }, [categoriesQuery.data, expectedQuery.data, expectedQuery.isSuccess, items, period]);

  const totals = useMemo(() => {
    return items.reduce<Record<string, number>>((next, item) => {
      const key = item.metadata?.category ?? "Unassigned";
      next[key] = (next[key] ?? 0) + item.endMinutes - item.startMinutes;
      return next;
    }, {});
  }, [items]);

  const isLoading =
    eventsQuery.isLoading ||
    eventCategoryOverlaysQuery.isLoading ||
    timeEntriesQuery.isLoading ||
    tzSegmentsQuery.isLoading ||
    categoriesQuery.isLoading ||
    expectedQuery.isLoading;
  const error =
    eventsQuery.error ??
    eventCategoryOverlaysQuery.error ??
    timeEntriesQuery.error ??
    tzSegmentsQuery.error ??
    categoriesQuery.error ??
    expectedQuery.error;

  return {
    items,
    summary,
    totals,
    isLoading,
    error,
  };
}
