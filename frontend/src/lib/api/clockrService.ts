import * as generatedService from "../../../wailsjs/go/service/Service";
import type {
  Calendar,
  Category,
  DayTimeline,
  Event,
  GapFill,
  Period,
  ReviewItem,
  TzSegment,
} from "./types";

interface ClockrBackend {
  ComputeGaps(ctx: unknown, periodId: number): Promise<DayTimeline[]>;
  GetSetting(ctx: unknown, key: string): Promise<string>;
  ListCalendars(ctx: unknown): Promise<Calendar[]>;
  ListCategories(ctx: unknown): Promise<Category[]>;
  ListEvents(ctx: unknown, periodId: number): Promise<Event[]>;
  ListGapFills(ctx: unknown, periodId: number): Promise<GapFill[]>;
  ListOpenReviewItems(ctx: unknown, periodId: number): Promise<ReviewItem[]>;
  ListPeriods(ctx: unknown): Promise<Period[]>;
  ListSelectedCalendars(ctx: unknown): Promise<Calendar[]>;
  ListTzSegments(ctx: unknown, periodId: number): Promise<TzSegment[]>;
}

declare global {
  interface Window {
    go?: {
      service?: {
        Service?: unknown;
      };
    };
  }
}

const backend = generatedService as unknown as ClockrBackend;
const wailsContext = undefined;

export function isClockrBackendAvailable() {
  return Boolean(
    typeof window !== "undefined" &&
      window.go?.service?.Service,
  );
}

async function readFromBackend<T>(fallback: T, read: () => Promise<T>) {
  if (!isClockrBackendAvailable()) {
    return fallback;
  }

  return read();
}

export function listPeriods() {
  return readFromBackend<Period[]>([], () => backend.ListPeriods(wailsContext));
}

export function listCategories() {
  return readFromBackend<Category[]>([], () =>
    backend.ListCategories(wailsContext),
  );
}

export function listCalendars() {
  return readFromBackend<Calendar[]>([], () =>
    backend.ListCalendars(wailsContext),
  );
}

export function listSelectedCalendars() {
  return readFromBackend<Calendar[]>([], () =>
    backend.ListSelectedCalendars(wailsContext),
  );
}

export function listEvents(periodId: number) {
  return readFromBackend<Event[]>([], () =>
    backend.ListEvents(wailsContext, periodId),
  );
}

export function listGapFills(periodId: number) {
  return readFromBackend<GapFill[]>([], () =>
    backend.ListGapFills(wailsContext, periodId),
  );
}

export function listOpenReviewItems(periodId: number) {
  return readFromBackend<ReviewItem[]>([], () =>
    backend.ListOpenReviewItems(wailsContext, periodId),
  );
}

export function listTzSegments(periodId: number) {
  return readFromBackend<TzSegment[]>([], () =>
    backend.ListTzSegments(wailsContext, periodId),
  );
}

export function computeGaps(periodId: number) {
  return readFromBackend<DayTimeline[]>([], () =>
    backend.ComputeGaps(wailsContext, periodId),
  );
}

export function getSetting(key: string) {
  return readFromBackend<string | null>(null, () =>
    backend.GetSetting(wailsContext, key),
  );
}
