import { Clock, Settings } from "lucide-react";
import { Link, useRouterState } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface AppHeaderProps {
  titlebarPaddingClass: string;
}

export function AppHeader({ titlebarPaddingClass }: AppHeaderProps) {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const settingsActive = pathname.startsWith("/settings");

  return (
    <header
      className={`shrink-0 flex items-center gap-3 py-2 pr-3 ${titlebarPaddingClass}`}
    >
      <Link to="/" className="flex items-center gap-3 app-no-drag">
        <div className="bg-primary rounded-md text-accent p-1.5">
          <Clock className="size-4" />
        </div>
        <div>
          <h1 className="text-base font-medium">shiet</h1>
        </div>
      </Link>
      <div className="grow" />
      <div className="app-no-drag">
        <Button
          asChild
          variant="ghost"
          size="icon"
          aria-label="Open settings"
          aria-current={settingsActive ? "page" : undefined}
          className={cn(settingsActive && "bg-muted")}
        >
          <Link to="/settings/general">
            <Settings className="size-4" />
          </Link>
        </Button>
      </div>
    </header>
  );
}
