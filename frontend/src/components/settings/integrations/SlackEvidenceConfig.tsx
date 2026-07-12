import { useEffect, useMemo } from "react";
import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from "@/components/ui/item";
import { Toggle } from "@/components/ui/toggle";
import {
  useIntegrationConnections,
  useSetSlackChannelSelected,
  useSlackChannels,
} from "@/lib/api";
import { SettingBlock } from "../SettingBlock";
import { ScrollArea } from "../../ui/scroll-area";
import type { IntegrationConfigSlotProps } from "./types";

const PROVIDER_ID = "slack";

export function SlackEvidenceConfig({
  disabled = false,
  onBusyChange,
}: IntegrationConfigSlotProps) {
  const connectionsQuery = useIntegrationConnections();
  const channelsQuery = useSlackChannels();
  const setChannelSelected = useSetSlackChannelSelected();

  const slackConnections = useMemo(
    () =>
      (connectionsQuery.data ?? []).filter(
        (connection) => connection.provider === PROVIDER_ID,
      ),
    [connectionsQuery.data],
  );

  const accountLabels = useMemo(() => {
    const labels = new Map<string, string>();
    for (const connection of slackConnections) {
      labels.set(connection.accountId, connection.accountLabel);
    }
    return labels;
  }, [slackConnections]);

  const channels = channelsQuery.data ?? [];

  const slotBusy = setChannelSelected.isPending;

  useEffect(() => {
    onBusyChange?.(slotBusy);
  }, [onBusyChange, slotBusy]);

  const isDisabled = disabled || slotBusy;

  return (
    <SettingBlock
      title="Channels"
      description="Choose which channels to track as evidence for AI gap-fill. Tracked channels are used later when fetching message history."
    >
      {channelsQuery.isLoading ? (
        <p className="text-sm text-muted-foreground">Loading channels…</p>
      ) : channels.length === 0 ? (
        <p className="text-sm text-muted-foreground">
          {slackConnections.length > 0
            ? "No channels found. Try Refresh channels on the connection above."
            : "Connect a Slack workspace to see channels here."}
        </p>
      ) : (
        <ScrollArea className="max-h-64 h-64 overflow-auto rounded-md border px-2">
          <ItemGroup className="gap-2 my-2">
            {channels.map((channel) => (
              <Item key={channel.id} variant="outline">
                <ItemContent className="min-w-0">
                  <ItemTitle className="flex flex-wrap items-center gap-2">
                    <span className="truncate">#{channel.name}</span>
                    {accountLabels.get(channel.accountId) ? (
                      <span className="rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
                        {accountLabels.get(channel.accountId)}
                      </span>
                    ) : null}
                    {channel.private ? (
                      <span className="rounded-full bg-muted px-2 py-0.5 text-[10px] font-medium text-muted-foreground">
                        Private
                      </span>
                    ) : null}
                  </ItemTitle>
                </ItemContent>
                <ItemActions>
                  <Toggle
                    pressed={channel.selected}
                    variant="outline"
                    size="sm"
                    disabled={isDisabled}
                    aria-label={`Track #${channel.name} as evidence`}
                    onPressedChange={(pressed) => {
                      void setChannelSelected.mutateAsync({
                        channelID: channel.id,
                        selected: pressed,
                      });
                    }}
                  >
                    {channel.selected ? "Tracking" : "Track"}
                  </Toggle>
                </ItemActions>
              </Item>
            ))}
          </ItemGroup>
        </ScrollArea>
      )}
    </SettingBlock>
  );
}
