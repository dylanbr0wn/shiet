import type { HTMLAttributes, ReactNode } from "react";
import { useScheduler, type SchedulerApi } from "./useScheduler";
import { formatMinutes } from "./time";
import type {
  SchedulerDay,
  SchedulerLayoutItem,
  SchedulerOptions,
} from "./types";

export interface SchedulerProps<TMetadata = unknown>
  extends SchedulerOptions<TMetadata> {
  children: (scheduler: SchedulerApi<TMetadata>) => ReactNode;
}

export function Scheduler<TMetadata = unknown>({
  children,
  ...options
}: SchedulerProps<TMetadata>) {
  const scheduler = useScheduler(options);
  return <>{children(scheduler)}</>;
}

export interface SchedulerRootProps<TMetadata = unknown>
  extends HTMLAttributes<HTMLDivElement> {
  scheduler: SchedulerApi<TMetadata>;
}

export function SchedulerRoot<TMetadata = unknown>({
  scheduler,
  ...props
}: SchedulerRootProps<TMetadata>) {
  return <div {...scheduler.getRootProps(props)} />;
}

export interface SchedulerDayColumnProps<TMetadata = unknown>
  extends HTMLAttributes<HTMLDivElement> {
  scheduler: SchedulerApi<TMetadata>;
  day: SchedulerDay;
}

export function SchedulerDayColumn<TMetadata = unknown>({
  scheduler,
  day,
  ...props
}: SchedulerDayColumnProps<TMetadata>) {
  return <div {...scheduler.getDayColumnProps(day, props)} />;
}

export interface SchedulerItemLayerProps<TMetadata = unknown>
  extends Omit<HTMLAttributes<HTMLDivElement>, "children"> {
  scheduler: SchedulerApi<TMetadata>;
  day: SchedulerDay;
  children: (layoutItem: SchedulerLayoutItem<TMetadata>) => ReactNode;
}

export function SchedulerItemLayer<TMetadata = unknown>({
  scheduler,
  day,
  children,
  style,
  ...props
}: SchedulerItemLayerProps<TMetadata>) {
  const items = scheduler.layoutsByDay[day.date] ?? [];

  return (
    <div
      {...props}
      data-scheduler-layer={day.date}
      style={{
        position: "absolute",
        inset: 0,
        ...style,
      }}
    >
      {items.map((layoutItem) => children(layoutItem))}
    </div>
  );
}

export interface SchedulerTimeAxisProps<TMetadata = unknown>
  extends Omit<HTMLAttributes<HTMLDivElement>, "children"> {
  scheduler: SchedulerApi<TMetadata>;
  stepMinutes?: number;
  children?: (minutes: number, label: string) => ReactNode;
}

export function SchedulerTimeAxis<TMetadata = unknown>({
  scheduler,
  stepMinutes = 60,
  children,
  ...props
}: SchedulerTimeAxisProps<TMetadata>) {
  const marks: number[] = [];
  for (
    let minute = scheduler.visibleRange.startMinutes;
    minute <= scheduler.visibleRange.endMinutes;
    minute += stepMinutes
  ) {
    marks.push(minute);
  }

  return (
    <div {...props}>
      {marks.map((minute) => {
        const label = formatMinutes(minute);
        return (
          <div key={minute} data-scheduler-time={minute}>
            {children ? children(minute, label) : label}
          </div>
        );
      })}
    </div>
  );
}
