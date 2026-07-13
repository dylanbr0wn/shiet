import type { TimeEntry } from "@/lib/api";

export function eventItemId(eventId: number): string {
  return `event-${eventId}`;
}

export function timeEntryItemId(timeEntryId: number): string {
  return `time-entry-${timeEntryId}`;
}

export function allDayChipId(eventId: number, day: string): string {
  return `${eventItemId(eventId)}@${day}`;
}

export function parseEventItemId(itemId: string): number | null {
  const match = /^event-(\d+)$/.exec(itemId);
  if (!match) {
    return null;
  }

  const eventId = Number(match[1]);
  return Number.isFinite(eventId) ? eventId : null;
}

export function parseTimeEntryItemId(itemId: string): number | null {
  const match = /^time-entry-(\d+)$/.exec(itemId);
  if (!match) {
    return null;
  }

  const timeEntryId = Number(match[1]);
  return Number.isFinite(timeEntryId) ? timeEntryId : null;
}

export function isEventItemId(itemId: string): boolean {
  return parseEventItemId(itemId) !== null;
}

export function isTimeEntryItemId(itemId: string): boolean {
  return parseTimeEntryItemId(itemId) !== null;
}

export function buildTimeEntriesByItemId(
  timeEntries: TimeEntry[],
): Map<string, TimeEntry> {
  return new Map(
    timeEntries.map((timeEntry) => [timeEntryItemId(timeEntry.id), timeEntry]),
  );
}
