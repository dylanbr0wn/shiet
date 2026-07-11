import { createFileRoute } from "@tanstack/react-router";
import { ExportSettings } from "@/components/settings/ExportSettings";

export const Route = createFileRoute("/settings/export")({
  component: ExportSettingsPage,
});

function ExportSettingsPage() {
  return (
    <div className="h-full min-h-0 overflow-auto p-5">
      <ExportSettings />
    </div>
  );
}
