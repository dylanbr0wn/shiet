import { describe, expect, it } from "vitest";
import { parseScheduleViewDayCount } from "./schedulePage.types";

describe("parseScheduleViewDayCount", () => {
  it.each([1, 7, 14] as const)("accepts valid day count %i", (dayCount) => {
    expect(parseScheduleViewDayCount(dayCount)).toBe(dayCount);
  });

  it.each([null, undefined, 0, 3, 5, 99, "7", "14"])(
    "falls back to 7 for invalid value %s",
    (raw) => {
      expect(parseScheduleViewDayCount(raw)).toBe(7);
    },
  );
});
