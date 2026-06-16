import { Clock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { formatPeriodLabel } from "@/lib/schedule";
import type { SchedulePageViewModel } from "./useSchedulePage";

interface ScheduleHeaderProps {
  titlebarPaddingClass: string;
  schedule: SchedulePageViewModel;
}

export function ScheduleHeader({
  titlebarPaddingClass,
  schedule,
}: ScheduleHeaderProps) {
  return (
    <header
      className={`shrink-0 flex items-center gap-3 py-2 pr-3 ${titlebarPaddingClass}`}
    >
      <div className="bg-primary rounded-md text-accent p-1.5">
        <Clock className="size-4" />
      </div>
      <div>
        <h1 className="text-base font-medium">Clockr</h1>
      </div>
      <Separator orientation="vertical" className="my-2" />
      <div className="grow" />
      <div className="app-no-drag flex flex-wrap items-center gap-2">
        {schedule.periods.length > 0 && (
          <select
            value={schedule.activePeriod?.id ?? ""}
            onChange={(event) =>
              schedule.setSelectedPeriodId(Number(event.target.value))
            }
            aria-label="Period"
            className="h-8 min-w-48 rounded-lg border border-border bg-white px-2.5 text-sm font-medium text-foreground outline-none transition-colors focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
          >
            {schedule.periods.map((period) => (
              <option key={period.id} value={period.id}>
                {formatPeriodLabel(period)}
              </option>
            ))}
          </select>
        )}
        <Button
          type="button"
          variant="outline"
          className="bg-white"
          disabled={
            !schedule.activePeriodId ||
            schedule.createPending ||
            schedule.days.length === 0
          }
          onClick={() =>
            schedule.handleCreate({
              day: schedule.days[0].date,
              startMinutes: 11 * 60,
              endMinutes: 12 * 60,
            })
          }
        >
          {schedule.createPending ? "Saving" : "Block"}
        </Button>
      </div>
    </header>
  );
}
