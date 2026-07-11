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

const sectionLinkClass = cn(
  "relative inline-flex w-full items-center justify-start gap-1.5 rounded-md border border-transparent px-1.5 py-1.5 text-sm font-medium text-foreground/60 transition-all hover:text-foreground",
  "[&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
);

function SettingsLayout() {
  return (
    <section className="app-no-drag grid min-h-0 flex-1 grid-cols-[180px_minmax(0,1fr)] overflow-hidden border-t border-border">
      <nav className="flex h-full flex-col gap-0.5 border-r border-border bg-muted p-1">
        {settingsNavItems.map((section) => {
          const Icon = section.icon;

          if (!section.ready) {
            return (
              <button
                key={section.id}
                type="button"
                disabled
                className={cn(sectionLinkClass, "opacity-50")}
              >
                <Icon className="size-4" />
                <span className="truncate">{section.label}</span>
              </button>
            );
          }

          return (
            <Link
              key={section.id}
              to={sectionPaths[section.id]}
              className={sectionLinkClass}
              activeProps={{
                className:
                  "bg-background text-foreground shadow-sm dark:border-input dark:bg-input/30",
              }}
            >
              <Icon className="size-4" />
              <span className="truncate">{section.label}</span>
            </Link>
          );
        })}
      </nav>

      <div className="grid min-h-0 grid-rows-[minmax(0,1fr)_auto]">
        <div className="min-h-0 overflow-hidden">
          <Outlet />
        </div>
        <footer className="flex items-center justify-end border-t border-border px-5 py-3">
          <Button asChild variant="secondary" size="sm">
            <Link to="/">Done</Link>
          </Button>
        </footer>
      </div>
    </section>
  );
}
