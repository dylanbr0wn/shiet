import { createFileRoute } from "@tanstack/react-router";
import { GeneralSettings } from "@/components/settings/GeneralSettings";

export const Route = createFileRoute("/settings/general")({
  component: GeneralSettingsPage,
});

function GeneralSettingsPage() {
  return (
    <div className="h-full min-h-0 overflow-auto p-5">
      <GeneralSettings />
    </div>
  );
}
