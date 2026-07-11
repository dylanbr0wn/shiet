import { createFileRoute } from "@tanstack/react-router";
import { IntegrationsSettings } from "@/components/settings/integrations";

export const Route = createFileRoute("/settings/integrations")({
  component: IntegrationsSettingsPage,
});

function IntegrationsSettingsPage() {
  return (
    <div className="flex h-full min-h-0 flex-col overflow-hidden">
      <IntegrationsSettings />
    </div>
  );
}
