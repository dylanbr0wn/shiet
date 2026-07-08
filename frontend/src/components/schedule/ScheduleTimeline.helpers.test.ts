import { describe, expect, it } from "vitest";
import {
  ALL_DAY_CHIP_GAP,
  TIMELINE_MIN_HEIGHT,
  buildBackgroundCreateRequest,
  buildTimelineMarks,
  computeAllDayRowHeight,
  computeInitialTimelineScrollTop,
  computeTimelineHeight,
  getTimeLabelClass,
  minuteToTimelinePercent,
  pointerClientYToSnappedMinute,
} from "./ScheduleTimeline.helpers";
import {
  ALL_DAY_CHIP_HEIGHT,
  ALL_DAY_ROW_MIN_HEIGHT,
  ALL_DAY_ROW_PADDING,
  INITIAL_SCROLL_CONTEXT_MINUTES,
  TIMELINE_HOUR_HEIGHT,
  WORKING_START_MINUTES,
} from "@/lib/schedule";

const fullDayRange = { startMinutes: 0, endMinutes: 24 * 60 };
const workdayRange = { startMinutes: 8 * 60, endMinutes: 18 * 60 };

describe("ScheduleTimeline helpers", () => {
  describe("computeAllDayRowHeight", () => {
    it("returns zero when no day has chips", () => {
      expect(computeAllDayRowHeight(new Map<string, unknown[]>())).toBe(0);
    });

    it("uses padded chip height for one chip", () => {
      const height = computeAllDayRowHeight(new Map([["2026-06-08", [{}]]]));

      expect(height).toBe(ALL_DAY_ROW_PADDING * 2 + ALL_DAY_CHIP_HEIGHT);
      expect(height).toBeGreaterThan(ALL_DAY_ROW_MIN_HEIGHT);
    });

    it("sizes to the tallest chip stack", () => {
      const height = computeAllDayRowHeight(
        new Map([
          ["2026-06-08", [{}]],
          ["2026-06-09", [{}, {}, {}]],
        ]),
      );

      expect(height).toBe(
        ALL_DAY_ROW_PADDING * 2 +
          3 * ALL_DAY_CHIP_HEIGHT +
          2 * ALL_DAY_CHIP_GAP,
      );
    });
  });

  describe("timeline geometry", () => {
    it("computes timeline height with a minimum", () => {
      expect(computeTimelineHeight(workdayRange)).toBe(TIMELINE_MIN_HEIGHT);
      expect(computeTimelineHeight(fullDayRange)).toBe(
        24 * TIMELINE_HOUR_HEIGHT,
      );
    });

    it("maps minutes to clamped visible-range percentages", () => {
      expect(minuteToTimelinePercent(8 * 60, workdayRange)).toBe(0);
      expect(minuteToTimelinePercent(13 * 60, workdayRange)).toBe(50);
      expect(minuteToTimelinePercent(18 * 60, workdayRange)).toBe(100);
      expect(minuteToTimelinePercent(7 * 60, workdayRange)).toBe(0);
      expect(minuteToTimelinePercent(19 * 60, workdayRange)).toBe(100);
    });

    it("handles empty visible ranges", () => {
      expect(
        minuteToTimelinePercent(60, { startMinutes: 60, endMinutes: 60 }),
      ).toBe(0);
    });

    it("builds timeline marks through the end of the range", () => {
      expect(buildTimelineMarks({ startMinutes: 8 * 60, endMinutes: 10 * 60 })).toEqual([
        8 * 60,
        9 * 60,
        10 * 60,
      ]);
    });

    it("keeps boundary label placement distinct", () => {
      expect(getTimeLabelClass(workdayRange.startMinutes, workdayRange)).toContain(
        "translate-y-0",
      );
      expect(getTimeLabelClass(workdayRange.endMinutes, workdayRange)).toContain(
        "-translate-y-full",
      );
      expect(getTimeLabelClass(9 * 60, workdayRange)).toContain("-translate-y-2");
    });
  });

  describe("pointer coordinate helpers", () => {
    const rect = { top: 100, height: 400 };

    it("snaps pointer positions to minutes", () => {
      expect(pointerClientYToSnappedMinute(100, rect, workdayRange, 15)).toBe(
        8 * 60,
      );
      expect(pointerClientYToSnappedMinute(300, rect, workdayRange, 15)).toBe(
        13 * 60,
      );
      expect(pointerClientYToSnappedMinute(500, rect, workdayRange, 15)).toBe(
        18 * 60,
      );
    });

    it("falls back to the range start for zero-height columns", () => {
      expect(
        pointerClientYToSnappedMinute(
          300,
          { top: 100, height: 0 },
          workdayRange,
          15,
        ),
      ).toBe(8 * 60);
    });

    it("clamps background create requests near midnight", () => {
      const request = buildBackgroundCreateRequest({
        day: "2026-06-08",
        clientY: 500,
        rect,
        visibleRange: fullDayRange,
        slotMinutes: 15,
        createDurationMinutes: 60,
      });

      expect(request).toEqual({
        day: "2026-06-08",
        startMinutes: 23 * 60,
        endMinutes: 24 * 60,
      });
    });
  });

  describe("initial scroll", () => {
    it("scrolls to working start with context", () => {
      const timelineHeight = computeTimelineHeight(fullDayRange);

      expect(
        computeInitialTimelineScrollTop({
          visibleRange: fullDayRange,
          timelineHeight,
        }),
      ).toBe(
        ((WORKING_START_MINUTES - INITIAL_SCROLL_CONTEXT_MINUTES) /
          fullDayRange.endMinutes) *
          timelineHeight,
      );
    });
  });
});
