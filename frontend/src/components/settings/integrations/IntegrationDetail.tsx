import { useState } from "react";
import { Link } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { IntegrationConnectShell } from "./IntegrationConnectShell";
import { getIntegrationEntry } from "./registry";

export function IntegrationDetail({ providerId }: { providerId: string }) {
  const entry = getIntegrationEntry(providerId);
  const [connectBusy, setConnectBusy] = useState(false);
  const [slotBusy, setSlotBusy] = useState(false);

  if (!entry) {
    return (
      <div className="mx-auto max-w-2xl space-y-4 p-5">
        <Button type="button" variant="ghost" size="sm" asChild>
          <Link to="/settings/integrations">
            <ArrowLeft className="size-4" />
            Back to integrations
          </Link>
        </Button>
        <p className="text-sm text-muted-foreground">
          Unknown integration provider.
        </p>
      </div>
    );
  }

  const ConfigSlot = entry.ConfigSlot;

  return (
    <div className="flex h-full min-h-0 flex-col w-full">
      <div className="min-h-0 flex-1 overflow-auto p-5">
        <div className="mx-auto max-w-2xl space-y-6">
          <IntegrationConnectShell
            providerId={entry.id}
            displayName={entry.displayName}
            disabled={slotBusy}
            onBusyChange={setConnectBusy}
          />
          <ConfigSlot disabled={connectBusy} onBusyChange={setSlotBusy} />
        </div>
      </div>
    </div>
  );
}
