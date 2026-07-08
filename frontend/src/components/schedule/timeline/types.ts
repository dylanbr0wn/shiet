import type { MutableRefObject } from "react";
import type {
  SchedulerApi,
  SchedulerCreateRequest,
  SchedulerDay,
  SchedulerLayoutItem,
} from "@/lib/scheduler";
import type {
  ScheduleDayMetadata,
  ScheduleMetadata,
} from "@/lib/schedule";

export type TimelineScheduler = SchedulerApi<
  ScheduleMetadata,
  ScheduleDayMetadata
>;
export type TimelineDay = SchedulerDay<ScheduleDayMetadata>;
export type TimelineLayoutItem = SchedulerLayoutItem<ScheduleMetadata>;
export type BackgroundRequestRef = MutableRefObject<SchedulerCreateRequest | null>;
