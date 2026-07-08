import { SparklesIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import type { ScheduleGapOverlay as ScheduleGapOverlayModel } from "@/lib/schedule";
import { minuteToTimelinePercent } from "../ScheduleTimeline.helpers";
import type { TimelineScheduler } from "./types";

interface ScheduleGapOverlayProps {
  gap: ScheduleGapOverlayModel;
  isHovered: boolean;
  aiConfigured: boolean;
  scheduler: TimelineScheduler;
  onSelectGap: (gap: ScheduleGapOverlayModel) => void;
}

export function ScheduleGapOverlay({
  gap,
  isHovered,
  aiConfigured,
  scheduler,
  onSelectGap,
}: ScheduleGapOverlayProps) {
  const top = minuteToTimelinePercent(gap.startMinutes, scheduler.visibleRange);
  const bottom = minuteToTimelinePercent(gap.endMinutes, scheduler.visibleRange);
  const height = Math.max(bottom - top, 1);

  return (
    <div
      className={cn([
        "pointer-events-none absolute inset-x-1 z-[5] rounded-md border border-dashed transition-[border-color,background-color,opacity] duration-150",
        isHovered
          ? "border-zinc-300/70 bg-muted/10 opacity-100"
          : "border-transparent bg-transparent opacity-0",
      ])}
      style={{
        top: `${top}%`,
        height: `${height}%`,
      }}
    >
      <button
        type="button"
        data-scheduler-ignore-create=""
        className={cn([
          "absolute right-1 top-1 inline-flex size-6 items-center justify-center rounded-full border text-muted-foreground transition-[opacity,border-color,background-color,color,box-shadow] duration-150",
          isHovered
            ? "pointer-events-auto border-border/50 bg-background/60 opacity-35 shadow-none"
            : "pointer-events-none border-transparent bg-transparent opacity-0",
          "hover:opacity-100 hover:border-emerald-400/70 hover:bg-emerald-50/90 hover:text-emerald-900 hover:shadow-sm",
        ])}
        onClick={(event) => {
          event.preventDefault();
          event.stopPropagation();
          onSelectGap(gap);
        }}
        title={aiConfigured ? "Suggest gap fill" : "Configure AI for suggestions"}
      >
        <SparklesIcon className="size-3" />
      </button>
    </div>
  );
}
