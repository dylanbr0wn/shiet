import { createFileRoute } from "@tanstack/react-router";
import { CategorySettings } from "@/components/settings/CategorySettings";

export const Route = createFileRoute("/settings/categories")({
  component: CategoriesSettingsPage,
});

function CategoriesSettingsPage() {
  return (
    <div className="h-full min-h-0 overflow-auto p-5">
      <CategorySettings />
    </div>
  );
}
