import { createFileRoute } from "@tanstack/react-router";
import { AIModelSettings } from "@/components/settings/AIModelSettings";

export const Route = createFileRoute("/settings/ai")({
  component: AISettingsPage,
});

function AISettingsPage() {
  return (
    <div className="h-full min-h-0 overflow-auto p-5">
      <AIModelSettings />
    </div>
  );
}
