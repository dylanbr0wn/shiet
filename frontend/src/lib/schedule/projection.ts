/**
 * Schedule projection concentrates shiet schedule policy for the timeline.
 *
 * Intentional remaining leakage (not projected here):
 * - Sidebar / NeedsAttentionCard consume raw ReviewDecision[] for the review queue UI.
 * - Hover-gap highlight state lives in ScheduleTimeline (pointer UI, not schedule policy).
 * - ScheduleKind remains on metadata/chips for presentation styling via scheduleItemPresentation.
 */
import type {
  Category,
  DayTimeline,
  Event,
  EventCategoryOverlay,
  TimeEntry,
  ReviewDecision,
  TzSegment,
} from "@/lib/api";
import { buildTimeEntriesByItemId, eventItemId, timeEntryItemId } from "./ids";
import {
  buildAllDayChipsByDay,
  buildEventCategoryOverlayMap,
  eventToSchedulerItem,
  timeEntryToSchedulerItem,
  gapTimelineToOverlays,
  resolveEventCategoryId,
  type EventReviewState,
} from "./mappers";
import type {
  AllDayChip,
  ScheduleGapOverlay,
  ScheduleItem,
  SchedulePlacement,
} from "./types";

export interface ProjectSchedulePeriodArgs {
  events: Event[];
  eventCategoryOverlays: EventCategoryOverlay[];
  timeEntries: TimeEntry[];
  gapTimeline: DayTimeline[];
  reviewDecisions: ReviewDecision[];
  tzSegments: TzSegment[];
  categories: Category[];
  visibleDays: ReadonlySet<string>;
  draftPlacements: Record<string, SchedulePlacement>;
}

export interface ProjectedSchedulePeriod {
  categoriesById: Map<number, Category>;
  items: ScheduleItem[];
  allDayChipsByDay: Map<string, AllDayChip[]>;
  visibleGaps: ScheduleGapOverlay[];
  resettableDays: ReadonlySet<string>;
  timeEntriesByItemId: Map<string, TimeEntry>;
  reviewDecisionsByEventId: Map<number, EventReviewState>;
}

export function buildReviewStateByEventId(
  reviewDecisions: ReviewDecision[],
): Map<number, EventReviewState> {
  return new Map(
    reviewDecisions
      .filter((decision) => typeof decision.eventId === "number")
      .map((decision) => [
        decision.eventId as number,
        { reviewItemId: decision.id, kind: decision.kind },
      ]),
  );
}

export function buildResettableDays(timeEntries: TimeEntry[]): ReadonlySet<string> {
  return new Set(
    timeEntries
      .filter((timeEntry) => !timeEntry.method)
      .map((timeEntry) => timeEntry.localWorkDate),
  );
}

export function projectSchedulePeriod({
  events,
  eventCategoryOverlays,
  timeEntries,
  gapTimeline,
  reviewDecisions,
  tzSegments,
  categories,
  visibleDays,
  draftPlacements,
}: ProjectSchedulePeriodArgs): ProjectedSchedulePeriod {
  const categoriesById = new Map(categories.map((category) => [category.id, category]));
  const overlaysByKey = buildEventCategoryOverlayMap(eventCategoryOverlays);
  const timeEntriesByItemId = buildTimeEntriesByItemId(timeEntries);
  const reviewDecisionsByEventId = buildReviewStateByEventId(reviewDecisions);

  const allDayChipsByDay = buildAllDayChipsByDay(
    events,
    visibleDays,
    categoriesById,
    overlaysByKey,
    reviewDecisionsByEventId,
  );

  const items = [
    ...events
      .map((event) =>
        eventToSchedulerItem(
          event,
          tzSegments,
          categoriesById,
          resolveEventCategoryId(event, overlaysByKey),
          draftPlacements[eventItemId(event.id)],
          reviewDecisionsByEventId.get(event.id),
        ),
      )
      .filter((item): item is ScheduleItem => item !== null),
    ...timeEntries
      .map((timeEntry) =>
        timeEntryToSchedulerItem(
          timeEntry,
          categoriesById,
          tzSegments,
          draftPlacements[timeEntryItemId(timeEntry.id)],
        ),
      )
      .filter((item): item is ScheduleItem => item !== null),
  ];

  const visibleGaps = gapTimelineToOverlays(gapTimeline, visibleDays, tzSegments);
  const resettableDays = buildResettableDays(timeEntries);

  return {
    categoriesById,
    items,
    allDayChipsByDay,
    visibleGaps,
    resettableDays,
    timeEntriesByItemId,
    reviewDecisionsByEventId,
  };
}
