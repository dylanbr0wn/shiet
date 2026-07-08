import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  periodProgressPercent,
  sortedCategories,
  sortedCategoryNames,
  varianceMinutes,
  type PeriodExportSummary,
} from "@/lib/export";
import { formatDateKey, formatDuration } from "@/lib/schedule";
import { formatVariance } from "@/lib/export/formatters";
import { cn } from "@/lib/utils";

interface PeriodStatsPanelProps {
  summary: PeriodExportSummary | null;
  className?: string;
}

function varianceToneClass(variance: number) {
  if (variance > 0) {
    return "text-emerald-700 dark:text-emerald-400";
  }

  if (variance < 0) {
    return "text-amber-700 dark:text-amber-400";
  }

  return "text-muted-foreground";
}

function categoryBarWidth(minutes: number, maxMinutes: number) {
  if (maxMinutes <= 0) {
    return 0;
  }

  return Math.max(4, Math.round((minutes / maxMinutes) * 100));
}

export function PeriodStatsPanel({
  summary,
  className,
}: PeriodStatsPanelProps) {
  if (!summary) {
    return (
      <Card className={cn("app-no-drag", className)}>
        <CardHeader>
          <CardTitle className="text-sm">Period stats</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">No active period.</p>
        </CardContent>
      </Card>
    );
  }

  const categories = sortedCategories(summary);
  const maxCategoryMinutes = Math.max(
    ...categories.map((category) => summary.periodTotals[category]),
    1,
  );
  const progressPercent = periodProgressPercent(summary);
  const progressWidth = Math.min(progressPercent, 100);
  const variance = varianceMinutes(summary.actualMinutes, summary.targetMinutes);

  return (
    <Card className={cn("app-no-drag min-h-0 overflow-auto overscroll-none", className)}>
      <CardHeader>
        <CardTitle className="text-sm">Period stats</CardTitle>
        <CardDescription>
          {summary.startDate} to {summary.endDate}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <div className="flex items-end justify-between gap-3">
            <p className="text-lg font-semibold text-foreground">
              {formatDuration(summary.actualMinutes)}
              <span className="text-sm font-medium text-muted-foreground">
                {" "}
                / {formatDuration(summary.targetMinutes)}
              </span>
            </p>
            <span className="text-sm font-medium text-muted-foreground">
              {progressPercent}%
            </span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-muted">
            <div
              className="h-full rounded-full bg-primary transition-[width]"
              style={{ width: `${progressWidth}%` }}
            />
          </div>
          <p className={cn("text-xs font-medium", varianceToneClass(variance))}>
            {formatVariance(summary.actualMinutes, summary.targetMinutes)}{" "}
            vs target · {summary.targetHoursPerDay}h/day
          </p>
        </div>

        {categories.length > 0 ? (
          <div className="space-y-2">
            <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              By category
            </h3>
            {categories.map((category) => {
              const minutes = summary.periodTotals[category];

              return (
                <div key={category} className="space-y-1">
                  <div className="flex items-center justify-between gap-3 text-sm">
                    <span className="truncate text-muted-foreground">{category}</span>
                    <span className="font-semibold text-foreground">
                      {formatDuration(minutes)}
                    </span>
                  </div>
                  <div className="h-1.5 overflow-hidden rounded-full bg-muted">
                    <div
                      className="h-full rounded-full bg-primary/70"
                      style={{
                        width: `${categoryBarWidth(minutes, maxCategoryMinutes)}%`,
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">No tracked time yet.</p>
        )}

        <Separator />
        <div className="space-y-3">
          <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Daily breakdown
          </h3>
          {summary.dailyTotals.map((day) => {
            const dayCategories = sortedCategoryNames(day.categories);
            const dayVariance = varianceMinutes(
              day.actualMinutes,
              day.targetMinutes,
            );

            return (
              <div
                key={day.date}
                className="rounded-md border border-border bg-muted/30 p-3"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-sm font-medium text-foreground">
                      {formatDateKey(day.date)}
                    </p>
                    <p className="text-xs text-muted-foreground">{day.date}</p>
                  </div>
                  <div className="text-right text-sm">
                    <p className="font-semibold text-foreground">
                      {formatDuration(day.actualMinutes)}
                      <span className="font-medium text-muted-foreground">
                        {" "}
                        / {formatDuration(day.targetMinutes)}
                      </span>
                    </p>
                    <p className={cn("text-xs font-medium", varianceToneClass(dayVariance))}>
                      {formatVariance(day.actualMinutes, day.targetMinutes)}
                    </p>
                  </div>
                </div>
                {dayCategories.length > 0 ? (
                  <div className="mt-2 space-y-1 border-t border-border/70 pt-2">
                    {dayCategories.map((category) => (
                      <div
                        key={category}
                        className="flex items-center justify-between gap-3 text-xs"
                      >
                        <span className="truncate text-muted-foreground">
                          {category}
                        </span>
                        <span className="font-medium text-foreground">
                          {formatDuration(day.categories[category])}
                        </span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="mt-2 border-t border-border/70 pt-2 text-xs text-muted-foreground">
                    No tracked time
                  </p>
                )}
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}
