import {
  useEffect,
  useMemo,
  useState,
  type Dispatch,
  type SetStateAction,
} from "react";
import {
  useCategories,
  useCreateManualEvent,
  useCurrentPeriod,
  useEvents,
  useGapFills,
  useOpenReviewItems,
  usePeriods,
  useTzSegments,
  useUpdateManualEvent,
  type Category,
  type Period,
} from "@/lib/api";
import type { SchedulerCreateRequest } from "@/lib/scheduler";
import {
  START_DATE,
  buildDays,
  defaultTimeZone,
  eventToSchedulerItem,
  gapFillToSchedulerItem,
  localDateKey,
  periodContainsDate,
  periodDayCount,
  type ScheduleChange,
  type ScheduleDay,
  type ScheduleItem,
  type ScheduleMetadata,
  type SchedulePlacement,
} from "@/lib/schedule";

export interface SchedulePageViewModel {
  selectedPeriodId: number | null;
  setSelectedPeriodId: Dispatch<SetStateAction<number | null>>;
  periods: Period[];
  activePeriod: Period | null;
  activePeriodId: number | undefined;
  categories: Category[];
  days: ScheduleDay[];
  items: ScheduleItem[];
  totals: Record<string, number>;
  visibleDayCount: number;
  preview: ScheduleChange | null;
  setPreview: Dispatch<SetStateAction<ScheduleChange | null>>;
  isBackendLoading: boolean;
  backendError: unknown;
  counts: {
    events: number;
    gapFills: number;
    categories: number;
    reviewItems: number;
  };
  createPending: boolean;
  handleCreate: (request: SchedulerCreateRequest) => void;
  handleCommit: (change: ScheduleChange) => void;
}

export function useSchedulePage(): SchedulePageViewModel {
  const [selectedPeriodId, setSelectedPeriodId] = useState<number | null>(null);
  const [draftPlacements, setDraftPlacements] = useState<
    Record<string, SchedulePlacement>
  >({});
  const [preview, setPreview] = useState<ScheduleChange | null>(null);
  const today = useMemo(() => localDateKey(), []);
  const currentTimeZone = useMemo(() => defaultTimeZone(), []);
  const periodsQuery = usePeriods();
  const currentPeriodQuery = useCurrentPeriod(today, currentTimeZone);
  const categoriesQuery = useCategories();
  const createManualEventMutation = useCreateManualEvent();
  const updateManualEventMutation = useUpdateManualEvent();
  const persistedPeriods = useMemo(
    () => periodsQuery.data ?? [],
    [periodsQuery.data],
  );
  const currentPeriod = currentPeriodQuery.data ?? null;
  const periods = useMemo(() => {
    if (
      currentPeriod &&
      !persistedPeriods.some((period) => period.id === currentPeriod.id)
    ) {
      return [currentPeriod, ...persistedPeriods];
    }

    return persistedPeriods;
  }, [currentPeriod, persistedPeriods]);
  const categories = useMemo(
    () => categoriesQuery.data ?? [],
    [categoriesQuery.data],
  );
  const activePeriod = useMemo(
    () =>
      periods.find((period) => period.id === selectedPeriodId) ??
      currentPeriod ??
      periods.find((period) => periodContainsDate(period, today)) ??
      periods[0] ??
      null,
    [currentPeriod, periods, selectedPeriodId, today],
  );
  const activePeriodId = activePeriod?.id;
  const eventsQuery = useEvents(activePeriodId);
  const gapFillsQuery = useGapFills(activePeriodId);
  const reviewItemsQuery = useOpenReviewItems(activePeriodId);
  const tzSegmentsQuery = useTzSegments(activePeriodId);
  const categoriesById = useMemo(() => {
    return new Map(categories.map((category) => [category.id, category]));
  }, [categories]);
  const visibleDayCount = periodDayCount(activePeriod);
  const days = useMemo(
    () => buildDays(activePeriod?.startDate ?? START_DATE, visibleDayCount),
    [activePeriod?.startDate, visibleDayCount],
  );
  const gapFillsByItemId = useMemo(() => {
    return new Map(
      (gapFillsQuery.data ?? []).map((gapFill) => [
        `gap-fill-${gapFill.id}`,
        gapFill,
      ]),
    );
  }, [gapFillsQuery.data]);
  const backendItems = useMemo(() => {
    const events = eventsQuery.data ?? [];
    const gapFills = gapFillsQuery.data ?? [];
    const tzSegments = tzSegmentsQuery.data ?? [];

    return [
      ...events
        .map((event) =>
          eventToSchedulerItem(
            event,
            tzSegments,
            draftPlacements[`event-${event.id}`],
          ),
        )
        .filter((item): item is ScheduleItem => item !== null),
      ...gapFills
        .map((gapFill) =>
          gapFillToSchedulerItem(
            gapFill,
            categoriesById,
            tzSegments,
            draftPlacements[`gap-fill-${gapFill.id}`],
          ),
        )
        .filter((item): item is ScheduleItem => item !== null),
    ];
  }, [
    categoriesById,
    draftPlacements,
    eventsQuery.data,
    gapFillsQuery.data,
    tzSegmentsQuery.data,
  ]);
  const items = backendItems;
  const totals = useMemo(() => {
    return items.reduce<Record<string, number>>((next, item) => {
      const key = item.metadata?.category ?? "Unassigned";
      next[key] = (next[key] ?? 0) + item.endMinutes - item.startMinutes;
      return next;
    }, {});
  }, [items]);
  const isBackendLoading =
    periodsQuery.isLoading ||
    currentPeriodQuery.isLoading ||
    categoriesQuery.isLoading ||
    eventsQuery.isLoading ||
    gapFillsQuery.isLoading ||
    reviewItemsQuery.isLoading ||
    tzSegmentsQuery.isLoading ||
    createManualEventMutation.isPending ||
    updateManualEventMutation.isPending;
  const backendError =
    periodsQuery.error ??
    currentPeriodQuery.error ??
    categoriesQuery.error ??
    eventsQuery.error ??
    gapFillsQuery.error ??
    reviewItemsQuery.error ??
    tzSegmentsQuery.error ??
    createManualEventMutation.error ??
    updateManualEventMutation.error;

  useEffect(() => {
    setSelectedPeriodId((current) => {
      if (
        currentPeriod &&
        (!current || !periods.some((period) => period.id === current))
      ) {
        return currentPeriod.id;
      }

      if (current && periods.some((period) => period.id === current)) {
        return current;
      }

      return currentPeriod?.id ?? periods[0]?.id ?? null;
    });
  }, [currentPeriod, periods]);

  useEffect(() => {
    setDraftPlacements({});
    setPreview(null);
  }, [activePeriodId]);

  const handleCreate = (request: SchedulerCreateRequest) => {
    if (activePeriodId) {
      createManualEventMutation.mutate({
        periodId: activePeriodId,
        day: request.day,
        startMinutes: request.startMinutes,
        endMinutes: request.endMinutes,
        note: "New block",
      });
      return;
    }
  };

  const handleCommit = (change: ScheduleChange) => {
    if (change.itemId.startsWith("gap-fill-")) {
      const gapFill = gapFillsByItemId.get(change.itemId);

      if (gapFill) {
        setDraftPlacements((current) => ({
          ...current,
          [change.itemId]: {
            day: change.day,
            startMinutes: change.startMinutes,
            endMinutes: change.endMinutes,
          },
        }));
        updateManualEventMutation.mutate(
          {
            id: gapFill.id,
            periodId: gapFill.periodId,
            day: change.day,
            startMinutes: change.startMinutes,
            endMinutes: change.endMinutes,
            categoryId: gapFill.categoryId,
            note: gapFill.note ?? "",
          },
          {
            onSettled: () => {
              setDraftPlacements((current) => {
                const next = { ...current };
                delete next[change.itemId];
                return next;
              });
            },
          },
        );
      }

      setPreview(null);
      return;
    }

    if (change.itemId.startsWith("event-")) {
      setDraftPlacements((current) => ({
        ...current,
        [change.itemId]: {
          day: change.day,
          startMinutes: change.startMinutes,
          endMinutes: change.endMinutes,
        },
      }));
    }

    setPreview(null);
  };

  return {
    selectedPeriodId,
    setSelectedPeriodId,
    periods,
    activePeriod,
    activePeriodId,
    categories,
    days,
    items,
    totals,
    visibleDayCount,
    preview,
    setPreview: setPreview as Dispatch<
      SetStateAction<ScheduleChange | null>
    >,
    isBackendLoading,
    backendError,
    counts: {
      events: eventsQuery.data?.length ?? 0,
      gapFills: gapFillsQuery.data?.length ?? 0,
      categories: categories.length,
      reviewItems: reviewItemsQuery.data?.length ?? 0,
    },
    createPending: createManualEventMutation.isPending,
    handleCreate,
    handleCommit,
  };
}
