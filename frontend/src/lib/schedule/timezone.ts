import type { TzSegment } from "@/lib/api";

export function toDate(value: string | undefined) {
  if (!value) {
    return null;
  }

  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

export function defaultTimeZone() {
  return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
}

export function zonedDateTimeParts(date: Date, timeZone: string) {
  const parts = new Intl.DateTimeFormat("en-CA", {
    day: "2-digit",
    hour: "2-digit",
    hourCycle: "h23",
    minute: "2-digit",
    month: "2-digit",
    timeZone,
    year: "numeric",
  }).formatToParts(date);
  const values = Object.fromEntries(
    parts
      .filter((part) => part.type !== "literal")
      .map((part) => [part.type, part.value]),
  );

  return {
    day: `${values.year}-${values.month}-${values.day}`,
    minutes: Number(values.hour) * 60 + Number(values.minute),
  };
}

export function activeTimeZoneForDay(day: string, tzSegments: TzSegment[]) {
  if (tzSegments.length === 0) {
    return defaultTimeZone();
  }

  let activeSegment = tzSegments[0];
  for (const segment of tzSegments) {
    if (segment.effectiveFromDate <= day) {
      activeSegment = segment;
    } else {
      break;
    }
  }

  return activeSegment.ianaTz;
}

export function zonedPosition(
  value: string | undefined,
  tzSegments: TzSegment[],
) {
  const date = toDate(value);

  if (!date) {
    return null;
  }

  const initialTimeZone = tzSegments[0]?.ianaTz ?? defaultTimeZone();
  const initialParts = zonedDateTimeParts(date, initialTimeZone);
  const activeTimeZone = activeTimeZoneForDay(initialParts.day, tzSegments);

  if (activeTimeZone === initialTimeZone) {
    return initialParts;
  }

  return zonedDateTimeParts(date, activeTimeZone);
}
