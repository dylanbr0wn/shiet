import { createFileRoute } from "@tanstack/react-router";
import { PrivacySettings } from "@/components/settings/PrivacySettings";

export const Route = createFileRoute("/settings/privacy")({
  component: PrivacySettingsPage,
});

function PrivacySettingsPage() {
  return (
    <div className="h-full min-h-0 overflow-auto p-5">
      <PrivacySettings />
    </div>
  );
}
