import { useEffect, useMemo, useState } from "react";
import { ItemGroup } from "@/components/ui/item";
import {
  useConnectIntegration,
  useDisconnectIntegration,
  useIntegrationConnections,
} from "@/lib/api";
import { SettingBlock } from "../SettingBlock";
import { AuthModeDescription } from "./AuthModeDescription";
import { ConnectActions } from "./ConnectActions";
import { ConnectionCard } from "./ConnectionCard";
import type { IntegrationProviderId } from "./registry";

type IntegrationConnectShellProps = {
  providerId: IntegrationProviderId;
  displayName: string;
  disabled?: boolean;
  onBusyChange?: (busy: boolean) => void;
};

export function IntegrationConnectShell({
  providerId,
  displayName,
  disabled = false,
  onBusyChange,
}: IntegrationConnectShellProps) {
  const connectionsQuery = useIntegrationConnections();
  const connectIntegration = useConnectIntegration();
  const disconnectIntegration = useDisconnectIntegration();

  const [accountEmail, setAccountEmail] = useState("");
  const [connectError, setConnectError] = useState<string | null>(null);

  const providerConnections = useMemo(
    () =>
      (connectionsQuery.data ?? []).filter(
        (connection) => connection.provider === providerId,
      ),
    [connectionsQuery.data, providerId],
  );

  const connectBusy =
    connectIntegration.isPending || disconnectIntegration.isPending;

  useEffect(() => {
    onBusyChange?.(connectBusy);
  }, [connectBusy, onBusyChange]);

  const isDisabled = disabled || connectBusy;

  const handleConnect = async (accountID: string, accountLabel: string) => {
    setConnectError(null);
    try {
      await connectIntegration.mutateAsync({
        provider: providerId,
        accountId: accountID,
        accountLabel,
      });
      setAccountEmail("");
    } catch (error) {
      setConnectError(
        error instanceof Error
          ? error.message
          : "Unable to connect Google account",
      );
    }
  };

  const handleDisconnect = async (accountID: string) => {
    setConnectError(null);
    try {
      await disconnectIntegration.mutateAsync({
        provider: providerId,
        accountID,
      });
    } catch (error) {
      setConnectError(
        error instanceof Error
          ? error.message
          : "Unable to disconnect Google account",
      );
    }
  };

  const handleReconnect = async (accountID: string, accountLabel: string) => {
    setConnectError(null);
    try {
      await connectIntegration.mutateAsync({
        provider: providerId,
        accountId: accountID,
        accountLabel,
      });
    } catch (error) {
      setConnectError(
        error instanceof Error
          ? error.message
          : "Unable to reconnect Google account",
      );
    }
  };

  if (providerId !== "google") {
    return null;
  }

  return (
    <>
      <SettingBlock
        title={displayName}
        description={<AuthModeDescription provider={providerId} />}
      >
        <ConnectActions
          provider="google"
          accountEmail={accountEmail}
          onAccountEmailChange={setAccountEmail}
          onConnect={() => {
            const email = accountEmail.trim();
            if (!email) {
              return;
            }
            void handleConnect(email, email);
          }}
          isConnecting={connectIntegration.isPending}
          disabled={isDisabled}
          connectError={connectError}
        />
      </SettingBlock>
      <SettingBlock
        title="Connected Google Accounts"
        description="Manage your connected Google accounts."
      >
        {providerConnections.length > 0 ? (
          <ItemGroup className="gap-2">
            {providerConnections.map((connection) => (
              <ConnectionCard
                key={connection.id}
                connection={connection}
                disabled={isDisabled}
                onDisconnect={(accountID) => void handleDisconnect(accountID)}
                onReconnect={(accountID, accountLabel) =>
                  void handleReconnect(accountID, accountLabel)
                }
              />
            ))}
          </ItemGroup>
        ) : (
          <p className="text-sm text-muted-foreground">
            No Google account connected yet.
          </p>
        )}
      </SettingBlock>
    </>
  );
}
