import { useEffect, useMemo } from "react";
import {
  Field,
  FieldLabel,
} from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from "@/components/ui/item";
import { Toggle } from "@/components/ui/toggle";
import {
  useCalendars,
  useCategories,
  useIntegrationConnections,
  useSetCalendarDefaultCategory,
  useSetCalendarSelected,
} from "@/lib/api";
import { SettingBlock } from "../SettingBlock";
import type { IntegrationConfigSlotProps } from "./types";

const NONE_CATEGORY = "__none__";
const PROVIDER_ID = "google";

export function CalendarSourceConfig({
  disabled = false,
  onBusyChange,
}: IntegrationConfigSlotProps) {
  const connectionsQuery = useIntegrationConnections();
  const calendarsQuery = useCalendars();
  const categoriesQuery = useCategories();
  const setCalendarSelected = useSetCalendarSelected();
  const setCalendarDefaultCategory = useSetCalendarDefaultCategory();

  const googleConnections = useMemo(
    () =>
      (connectionsQuery.data ?? []).filter(
        (connection) => connection.provider === PROVIDER_ID,
      ),
    [connectionsQuery.data],
  );

  const googleCalendars = useMemo(
    () =>
      (calendarsQuery.data ?? []).filter(
        (calendar) => calendar.provider === PROVIDER_ID,
      ),
    [calendarsQuery.data],
  );

  const categories = categoriesQuery.data ?? [];

  const slotBusy =
    setCalendarSelected.isPending || setCalendarDefaultCategory.isPending;

  useEffect(() => {
    onBusyChange?.(slotBusy);
  }, [onBusyChange, slotBusy]);

  const isDisabled = disabled || slotBusy;

  return (
    <SettingBlock
      title="Calendars"
      description="Choose which calendars to import. Primary is selected by default on first connect."
    >
      {calendarsQuery.isLoading ? (
        <p className="text-sm text-muted-foreground">Loading calendars…</p>
      ) : googleCalendars.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          {googleConnections.length > 0
            ? "Connect syncs your calendar list. Run Sync on the schedule view if calendars are missing."
            : "Connect a Google account to see calendars here."}
        </p>
      ) : (
        <ItemGroup className="gap-2">
          {googleCalendars.map((calendar) => (
            <Item
              key={calendar.id}
              variant="outline"
              className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto_minmax(0,180px)] sm:items-center"
            >
              <ItemContent className="min-w-0">
                <ItemTitle className="flex flex-wrap items-center gap-2">
                  <span className="truncate">{calendar.name}</span>
                  {calendar.isPrimary ? (
                    <span className="rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary">
                      Primary
                    </span>
                  ) : null}
                </ItemTitle>
              </ItemContent>

              <ItemActions>
                <Toggle
                  pressed={calendar.selected}
                  variant="outline"
                  size="sm"
                  disabled={isDisabled}
                  aria-label={`Import ${calendar.name}`}
                  onPressedChange={(pressed) => {
                    void setCalendarSelected.mutateAsync({
                      calendarID: calendar.id,
                      selected: pressed,
                    });
                  }}
                >
                  {calendar.selected ? "Importing" : "Import"}
                </Toggle>
              </ItemActions>

              <ItemContent className="min-w-0">
                <Field>
                  <FieldLabel
                    htmlFor={`calendar-category-${calendar.id}`}
                    className="sr-only"
                  >
                    Default category for {calendar.name}
                  </FieldLabel>
                  <Select
                    value={
                      calendar.defaultCategoryId
                        ? String(calendar.defaultCategoryId)
                        : NONE_CATEGORY
                    }
                    onValueChange={(value) => {
                      void setCalendarDefaultCategory.mutateAsync({
                        calendarID: calendar.id,
                        categoryID:
                          value === NONE_CATEGORY ? null : Number(value),
                      });
                    }}
                    disabled={isDisabled}
                  >
                    <SelectTrigger
                      id={`calendar-category-${calendar.id}`}
                      className="w-full bg-background"
                    >
                      <SelectValue placeholder="Default category" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value={NONE_CATEGORY}>No default</SelectItem>
                      {categories.map((category) => (
                        <SelectItem
                          key={category.id}
                          value={String(category.id)}
                        >
                          {category.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </Field>
              </ItemContent>
            </Item>
          ))}
        </ItemGroup>
      )}
    </SettingBlock>
  );
}
