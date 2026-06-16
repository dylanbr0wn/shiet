import type {
  SchedulerDay,
  SchedulerItem,
  SchedulerChange,
} from "@/lib/scheduler";

export type ScheduleKind = "calendar" | "gap" | "manual" | "review";

export interface ScheduleMetadata {
  title: string;
  category: string;
  kind: ScheduleKind;
}

export interface ScheduleDayMetadata {
  isWeekend: boolean;
}

export type ScheduleItem = SchedulerItem<ScheduleMetadata>;
export type ScheduleDay = SchedulerDay<ScheduleDayMetadata>;
export type ScheduleChange = SchedulerChange<ScheduleMetadata>;
export type SchedulePlacement = Pick<
  ScheduleItem,
  "day" | "endMinutes" | "startMinutes"
>;
