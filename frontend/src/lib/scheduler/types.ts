export type SchedulerInteraction = "move" | "resize-start" | "resize-end";

export interface SchedulerItem<TMetadata = unknown> {
  id: string;
  day: string;
  startMinutes: number;
  endMinutes: number;
  disabled?: boolean;
  metadata?: TMetadata;
}

export interface SchedulerDay<TMetadata = unknown> {
  date: string;
  label?: string;
  disabled?: boolean;
  metadata?: TMetadata;
}

export interface SchedulerVisibleRange {
  startMinutes: number;
  endMinutes: number;
}

export interface SchedulerConfig {
  slotMinutes: number;
  minDurationMinutes: number;
  createDurationMinutes: number;
  maxDays: number;
  workingStartMinutes: number;
  workingEndMinutes: number;
}

export interface SchedulerCreateRequest {
  day: string;
  startMinutes: number;
  endMinutes: number;
}

export interface SchedulerChange<TMetadata = unknown> {
  itemId: string;
  day: string;
  startMinutes: number;
  endMinutes: number;
  interaction: SchedulerInteraction;
  item: SchedulerItem<TMetadata>;
}

export interface SchedulerLayoutItem<TMetadata = unknown> {
  item: SchedulerItem<TMetadata>;
  day: string;
  topPercent: number;
  heightPercent: number;
  leftPercent: number;
  widthPercent: number;
  lane: number;
  laneCount: number;
  overlaps: boolean;
  isPreview: boolean;
}

export interface SchedulerOptions<TMetadata = unknown> {
  days: SchedulerDay[];
  items: SchedulerItem<TMetadata>[];
  config?: Partial<SchedulerConfig>;
  onCreate?: (request: SchedulerCreateRequest) => void;
  onPreviewChange?: (change: SchedulerChange<TMetadata>) => void;
  onCommitChange?: (change: SchedulerChange<TMetadata>) => void;
}

export const DEFAULT_SCHEDULER_CONFIG: SchedulerConfig = {
  slotMinutes: 15,
  minDurationMinutes: 15,
  createDurationMinutes: 60,
  maxDays: 14,
  workingStartMinutes: 8 * 60,
  workingEndMinutes: 18 * 60,
};
