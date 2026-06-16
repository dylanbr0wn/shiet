import type {
  Category,
  Event as ClockrEvent,
  GapFill,
  Period,
  TzSegment,
} from "@/lib/api";
import {
  SCHEDULE_END_MINUTES,
  SCHEDULE_START_MINUTES,
} from "./constants";
import {
  activeTimeZoneForDay,
  toDate,
  zonedDateTimeParts,
  zonedPosition,
} from "./timezone";
import type { ScheduleItem, SchedulePlacement } from "./types";

export function categoryName(
  categoryId: number | undefined,
  categoriesById: Map<number, Category>,
) {
  if (typeof categoryId !== "number") {
    return "Unassigned";
  }

  return categoriesById.get(categoryId)?.name ?? "Unassigned";
}

export function periodContainsDate(period: Period, day: string) {
  return period.startDate <= day && day <= period.endDate;
}

export function applyPlacement(
  item: ScheduleItem,
  placement?: SchedulePlacement,
) {
  if (!placement) {
    return item;
  }

  return {
    ...item,
    ...placement,
  };
}

export function eventToSchedulerItem(
  event: ClockrEvent,
  tzSegments: TzSegment[],
  placement?: SchedulePlacement,
): ScheduleItem | null {
  if (event.allDay && event.startDate) {
    return applyPlacement(
      {
        id: `event-${event.id}`,
        day: event.startDate,
        startMinutes: SCHEDULE_START_MINUTES,
        endMinutes: SCHEDULE_END_MINUTES,
        metadata: {
          title: event.title || "Untitled event",
          category: "Calendar",
          kind: "calendar",
        },
      },
      placement,
    );
  }

  const start = zonedPosition(event.start, tzSegments);
  const end = zonedPosition(event.end, tzSegments);

  if (!start || !end) {
    return null;
  }

  const endMinutes =
    end.day === start.day ? end.minutes : SCHEDULE_END_MINUTES;

  return applyPlacement(
    {
      id: `event-${event.id}`,
      day: start.day,
      startMinutes: start.minutes,
      endMinutes: Math.max(start.minutes + 15, endMinutes),
      metadata: {
        title: event.title || "Untitled event",
        category: "Calendar",
        kind: "calendar",
      },
    },
    placement,
  );
}

export function gapFillToSchedulerItem(
  gapFill: GapFill,
  categoriesById: Map<number, Category>,
  tzSegments: TzSegment[],
  placement?: SchedulePlacement,
): ScheduleItem | null {
  const timeZone = activeTimeZoneForDay(gapFill.day, tzSegments);
  const startDate = toDate(gapFill.start);
  const endDate = toDate(gapFill.end);

  if (!startDate || !endDate) {
    return null;
  }

  const start = zonedDateTimeParts(startDate, timeZone);
  const end = zonedDateTimeParts(endDate, timeZone);
  const startMinutes = start.minutes;
  const endMinutes =
    end.day === start.day ? end.minutes : SCHEDULE_END_MINUTES;
  const category = categoryName(gapFill.categoryId, categoriesById);

  return applyPlacement(
    {
      id: `gap-fill-${gapFill.id}`,
      day: gapFill.day || start.day,
      startMinutes,
      endMinutes: Math.max(startMinutes + 15, endMinutes),
      metadata: {
        title: gapFill.note || category,
        category,
        kind: "manual",
      },
    },
    placement,
  );
}
