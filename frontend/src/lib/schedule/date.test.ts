import { describe, expect, it } from "vitest";
import { buildDays, inclusiveDayCount, periodDayCount } from "./date";
import type { Period } from "@/lib/api";

describe("schedule dates", () => {
  it("counts inclusive UTC period days", () => {
    expect(inclusiveDayCount("2026-06-08", "2026-06-08")).toBe(1);
    expect(inclusiveDayCount("2026-06-08", "2026-06-14")).toBe(7);
  });

  it("falls back to one day for invalid date keys", () => {
    expect(inclusiveDayCount("not-a-date", "2026-06-14")).toBe(1);
  });

  it("uses the fallback day count when no period is active", () => {
    expect(periodDayCount(null)).toBe(7);
  });

  it("builds weekend metadata for scheduler days", () => {
    const days = buildDays("2026-06-12", 3);

    expect(days.map((day) => day.date)).toEqual([
      "2026-06-12",
      "2026-06-13",
      "2026-06-14",
    ]);
    expect(days.map((day) => day.metadata?.isWeekend)).toEqual([
      false,
      true,
      true,
    ]);
  });

  it("counts days from period boundaries", () => {
    const period: Period = {
      id: 1,
      startDate: "2026-06-08",
      endDate: "2026-06-21",
      cadence: "bi_weekly",
      anchorDate: "2026-06-08",
      targetHoursPerDay: 6,
    };

    expect(periodDayCount(period)).toBe(14);
  });
});
