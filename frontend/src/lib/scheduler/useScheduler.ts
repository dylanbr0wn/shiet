import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type CSSProperties,
  type HTMLAttributes,
  type MutableRefObject,
  type PointerEvent as ReactPointerEvent,
  type Ref,
} from "react";
import { packOverlaps } from "./layout";
import {
  MINUTES_PER_DAY,
  calculateVisibleRange,
  clamp,
  normalizeConfig,
  snapMinutes,
} from "./time";
import type {
  SchedulerChange,
  SchedulerConfig,
  SchedulerCreateRequest,
  SchedulerDay,
  SchedulerInteraction,
  SchedulerItem,
  SchedulerLayoutItem,
  SchedulerOptions,
  SchedulerVisibleRange,
} from "./types";

interface ActiveInteraction<TMetadata> {
  interaction: SchedulerInteraction;
  item: SchedulerItem<TMetadata>;
  pointerStartMinute: number;
  lastChange: SchedulerChange<TMetadata>;
}

type ElementProps<TElement extends HTMLElement> = HTMLAttributes<TElement> & {
  ref?: Ref<TElement>;
  [key: `data-${string}`]: string | boolean | undefined;
};

function assignRef<TElement>(ref: Ref<TElement> | undefined, value: TElement | null) {
  if (!ref) {
    return;
  }

  if (typeof ref === "function") {
    ref(value);
    return;
  }

  (ref as MutableRefObject<TElement | null>).current = value;
}

function composeHandlers<TEvent>(
  userHandler: ((event: TEvent) => void) | undefined,
  schedulerHandler: (event: TEvent) => void,
) {
  return (event: TEvent) => {
    userHandler?.(event);
    schedulerHandler(event);
  };
}

function toItemChange<TMetadata>(
  item: SchedulerItem<TMetadata>,
  interaction: SchedulerInteraction,
  day: string,
  startMinutes: number,
  endMinutes: number,
): SchedulerChange<TMetadata> {
  return {
    itemId: item.id,
    day,
    startMinutes,
    endMinutes,
    interaction,
    item,
  };
}

export interface SchedulerApi<TMetadata = unknown> {
  days: SchedulerDay[];
  items: SchedulerItem<TMetadata>[];
  visibleRange: SchedulerVisibleRange;
  config: SchedulerConfig;
  previewChange: SchedulerChange<TMetadata> | null;
  layoutsByDay: Record<string, SchedulerLayoutItem<TMetadata>[]>;
  getRootProps: <TElement extends HTMLElement>(
    props?: ElementProps<TElement>,
  ) => ElementProps<TElement>;
  getDayColumnProps: <TElement extends HTMLElement>(
    day: SchedulerDay,
    props?: ElementProps<TElement>,
  ) => ElementProps<TElement>;
  getItemProps: <TElement extends HTMLElement>(
    layoutItem: SchedulerLayoutItem<TMetadata>,
    props?: ElementProps<TElement>,
  ) => ElementProps<TElement>;
  getResizeHandleProps: <TElement extends HTMLElement>(
    layoutItem: SchedulerLayoutItem<TMetadata>,
    edge: "start" | "end",
    props?: ElementProps<TElement>,
  ) => ElementProps<TElement>;
}

export function useScheduler<TMetadata = unknown>({
  days,
  items,
  config: configOverrides,
  onCreate,
  onPreviewChange,
  onCommitChange,
}: SchedulerOptions<TMetadata>): SchedulerApi<TMetadata> {
  const config = useMemo(
    () => normalizeConfig(configOverrides),
    [configOverrides],
  );
  const activeDays = useMemo(
    () => days.slice(0, config.maxDays),
    [config.maxDays, days],
  );
  const dayDates = useMemo(
    () => activeDays.map((day) => day.date),
    [activeDays],
  );
  const dayRefs = useRef(new Map<string, HTMLElement>());
  const activeRef = useRef<ActiveInteraction<TMetadata> | null>(null);
  const suppressColumnClickUntilRef = useRef(0);
  const [previewChange, setPreviewChange] =
    useState<SchedulerChange<TMetadata> | null>(null);

  const visibleRange = useMemo(
    () => calculateVisibleRange(items, config),
    [config, items],
  );

  const previewItems = useMemo(() => {
    if (!previewChange) {
      return items;
    }

    return items.map((item) =>
      item.id === previewChange.itemId
        ? {
            ...item,
            day: previewChange.day,
            startMinutes: previewChange.startMinutes,
            endMinutes: previewChange.endMinutes,
          }
        : item,
    );
  }, [items, previewChange]);

  const layoutsByDay = useMemo(() => {
    return activeDays.reduce<Record<string, SchedulerLayoutItem<TMetadata>[]>>(
      (layouts, day) => {
        layouts[day.date] = packOverlaps(
          previewItems.filter((item) => item.day === day.date),
          visibleRange,
          previewChange?.itemId,
        );
        return layouts;
      },
      {},
    );
  }, [activeDays, previewChange?.itemId, previewItems, visibleRange]);

  const pointToMinutes = useCallback(
    (clientY: number, day: string) => {
      const column = dayRefs.current.get(day);
      if (!column) {
        return visibleRange.startMinutes;
      }

      const rect = column.getBoundingClientRect();
      const percent = rect.height <= 0 ? 0 : clamp((clientY - rect.top) / rect.height, 0, 1);
      const rawMinutes =
        visibleRange.startMinutes +
        percent * (visibleRange.endMinutes - visibleRange.startMinutes);
      return snapMinutes(rawMinutes, config.slotMinutes);
    },
    [config.slotMinutes, visibleRange],
  );

  const dayAtX = useCallback(
    (clientX: number) => {
      for (const day of dayDates) {
        const column = dayRefs.current.get(day);
        if (!column) {
          continue;
        }
        const rect = column.getBoundingClientRect();
        if (clientX >= rect.left && clientX <= rect.right) {
          return day;
        }
      }
      return null;
    },
    [dayDates],
  );

  const emitPreview = useCallback(
    (change: SchedulerChange<TMetadata>) => {
      activeRef.current = activeRef.current
        ? { ...activeRef.current, lastChange: change }
        : null;
      setPreviewChange(change);
      onPreviewChange?.(change);
    },
    [onPreviewChange],
  );

  const calculateChange = useCallback(
    (event: PointerEvent, active: ActiveInteraction<TMetadata>) => {
      const item = active.item;
      const duration = item.endMinutes - item.startMinutes;
      const targetDay =
        active.interaction === "move"
          ? dayAtX(event.clientX) ?? item.day
          : item.day;

      if (active.interaction === "move") {
        const pointerMinute = pointToMinutes(event.clientY, targetDay);
        const delta = pointerMinute - active.pointerStartMinute;
        const startMinutes = clamp(
          snapMinutes(item.startMinutes + delta, config.slotMinutes),
          0,
          MINUTES_PER_DAY - duration,
        );

        return toItemChange(
          item,
          active.interaction,
          targetDay,
          startMinutes,
          startMinutes + duration,
        );
      }

      const pointerMinute = pointToMinutes(event.clientY, item.day);

      if (active.interaction === "resize-start") {
        const startMinutes = clamp(
          pointerMinute,
          0,
          item.endMinutes - config.minDurationMinutes,
        );
        return toItemChange(
          item,
          active.interaction,
          item.day,
          startMinutes,
          item.endMinutes,
        );
      }

      const endMinutes = clamp(
        pointerMinute,
        item.startMinutes + config.minDurationMinutes,
        MINUTES_PER_DAY,
      );
      return toItemChange(
        item,
        active.interaction,
        item.day,
        item.startMinutes,
        endMinutes,
      );
    },
    [config.minDurationMinutes, config.slotMinutes, dayAtX, pointToMinutes],
  );

  const suppressColumnClick = useCallback(() => {
    suppressColumnClickUntilRef.current = window.performance.now() + 350;
  }, []);

  useEffect(() => {
    const handlePointerMove = (event: PointerEvent) => {
      const active = activeRef.current;
      if (!active) {
        return;
      }

      emitPreview(calculateChange(event, active));
    };

    const handlePointerUp = () => {
      const active = activeRef.current;
      if (!active) {
        return;
      }

      onCommitChange?.(active.lastChange);
      suppressColumnClick();
      activeRef.current = null;
      setPreviewChange(null);
    };

    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", handlePointerUp);
    window.addEventListener("pointercancel", handlePointerUp);

    return () => {
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", handlePointerUp);
      window.removeEventListener("pointercancel", handlePointerUp);
    };
  }, [calculateChange, emitPreview, onCommitChange, suppressColumnClick]);

  const startInteraction = useCallback(
    (
      interaction: SchedulerInteraction,
      item: SchedulerItem<TMetadata>,
      event: ReactPointerEvent<HTMLElement>,
    ) => {
      if (item.disabled) {
        return;
      }

      event.preventDefault();
      event.stopPropagation();

      const initialChange = toItemChange(
        item,
        interaction,
        item.day,
        item.startMinutes,
        item.endMinutes,
      );
      activeRef.current = {
        interaction,
        item,
        pointerStartMinute: pointToMinutes(event.clientY, item.day),
        lastChange: initialChange,
      };
      suppressColumnClick();
    },
    [pointToMinutes, suppressColumnClick],
  );

  const getRootProps = useCallback(
    <TElement extends HTMLElement>(
      props: ElementProps<TElement> = {},
    ): ElementProps<TElement> => ({
      ...props,
      role: props.role ?? "application",
      "aria-label": props["aria-label"] ?? "Scheduler",
      "data-scheduler-root": "",
    }),
    [],
  );

  const getDayColumnProps = useCallback(
    <TElement extends HTMLElement>(
      day: SchedulerDay,
      props: ElementProps<TElement> = {},
    ): ElementProps<TElement> => {
      const handleClick = (event: React.MouseEvent<TElement>) => {
        if (event.defaultPrevented || day.disabled) {
          return;
        }

        if (window.performance.now() < suppressColumnClickUntilRef.current) {
          event.preventDefault();
          event.stopPropagation();
          return;
        }

        const startMinutes = clamp(
          pointToMinutes(event.clientY, day.date),
          0,
          MINUTES_PER_DAY - config.createDurationMinutes,
        );
        const request: SchedulerCreateRequest = {
          day: day.date,
          startMinutes,
          endMinutes: startMinutes + config.createDurationMinutes,
        };
        onCreate?.(request);
      };

      const style: CSSProperties = {
        position: "relative",
        touchAction: "none",
        ...props.style,
      };

      return {
        ...props,
        ref: (node: TElement | null) => {
          if (node) {
            dayRefs.current.set(day.date, node);
          } else {
            dayRefs.current.delete(day.date);
          }
          assignRef(props.ref, node);
        },
        role: props.role ?? "gridcell",
        "aria-disabled": day.disabled || undefined,
        "data-scheduler-day": day.date,
        style,
        onClick: composeHandlers(props.onClick, handleClick),
      };
    },
    [config.createDurationMinutes, onCreate, pointToMinutes],
  );

  const getItemProps = useCallback(
    <TElement extends HTMLElement>(
      layoutItem: SchedulerLayoutItem<TMetadata>,
      props: ElementProps<TElement> = {},
    ): ElementProps<TElement> => {
      const item = layoutItem.item;
      const style: CSSProperties = {
        position: "absolute",
        top: `${layoutItem.topPercent}%`,
        height: `${layoutItem.heightPercent}%`,
        left: `${layoutItem.leftPercent}%`,
        width: `${layoutItem.widthPercent}%`,
        touchAction: "none",
        ...props.style,
      };

      return {
        ...props,
        role: props.role ?? "button",
        tabIndex: props.tabIndex ?? (item.disabled ? -1 : 0),
        "aria-disabled": item.disabled || undefined,
        "data-scheduler-item": item.id,
        "data-scheduler-preview": layoutItem.isPreview || undefined,
        style,
        onPointerDown: composeHandlers(props.onPointerDown, (event) =>
          startInteraction("move", item, event),
        ),
      };
    },
    [startInteraction],
  );

  const getResizeHandleProps = useCallback(
    <TElement extends HTMLElement>(
      layoutItem: SchedulerLayoutItem<TMetadata>,
      edge: "start" | "end",
      props: ElementProps<TElement> = {},
    ): ElementProps<TElement> => {
      const item = layoutItem.item;

      return {
        ...props,
        role: props.role ?? "separator",
        "aria-label":
          props["aria-label"] ??
          (edge === "start" ? "Resize start time" : "Resize end time"),
        "data-scheduler-resize": edge,
        onPointerDown: composeHandlers(props.onPointerDown, (event) =>
          startInteraction(edge === "start" ? "resize-start" : "resize-end", item, event),
        ),
      };
    },
    [startInteraction],
  );

  return {
    days: activeDays,
    items: previewItems,
    visibleRange,
    config,
    previewChange,
    layoutsByDay,
    getRootProps,
    getDayColumnProps,
    getItemProps,
    getResizeHandleProps,
  };
}
