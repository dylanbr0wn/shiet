import {
  createFileRoute,
  Link,
  Outlet,
} from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import {
  settingsNavItems,
  type SettingsSectionId,
} from "@/components/settings/settingsSections";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/settings")({
  component: SettingsLayout,
});

const sectionPaths = {
  general: "/settings/general",
  integrations: "/settings/integrations",
  categories: "/settings/categories",
  ai: "/settings/ai",
  export: "/settings/export",
} as const satisfies Record<SettingsSectionId, string>;

function SettingsLayout() {
  return (
    <section className="app-no-drag grid min-h-0 flex-1 grid-cols-[180px_minmax(0,1fr)] overflow-hidden border-t border-border">
      <nav className="flex h-full flex-col gap-0.5 border-r border-border bg-sidebar p-1">
        {settingsNavItems.map((section) => {
          const Icon = section.icon;

          if (!section.ready) {
            return (
              <Button
                key={section.id}
                type="button"
                disabled
                variant="ghost"
                className="opacity-50 justify-start"
              >
                <Icon className="size-4 " />
                <span className="truncate">{section.label}</span>
              </Button>
            );
          }

          return (
            <Button asChild variant="ghost" className="justify-start">
              <Link
                key={section.id}
                activeProps={{
                  className: cn("text-green-300 hover:text-green-300 bg-muted")
                }}
                to={sectionPaths[section.id]}
              >
                <Icon className="size-4" />
                <span className="truncate">{section.label}</span>
              </Link>
            </Button>

          );
        })}
      </nav>

      <div className="grid min-h-0 grid-rows-[minmax(0,1fr)_auto]">
        <div className="min-h-0 overflow-hidden">
          <Outlet />
        </div>
        <footer className="flex items-center justify-end border-t border-border px-5 py-3 bg-background-lighter">
          <Button asChild variant="secondary" size="sm">
            <Link to="/">Done</Link>
          </Button>
        </footer>
      </div>
    </section>
  );
}
