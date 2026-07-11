import {
  Download,
  Plug,
  Settings,
  Shield,
  Sparkles,
  Tags,
  type LucideIcon,
} from "lucide-react";

export type SettingsSectionId =
  | "general"
  | "integrations"
  | "categories"
  | "ai"
  | "export";

export const settingsNavItems: Array<
  | {
      id: SettingsSectionId;
      label: string;
      icon: LucideIcon;
      ready: true;
    }
  | {
      id: "privacy";
      label: string;
      icon: LucideIcon;
      ready: false;
    }
> = [
  { id: "general", label: "General", icon: Settings, ready: true },
  { id: "integrations", label: "Integrations", icon: Plug, ready: true },
  { id: "categories", label: "Categories", icon: Tags, ready: true },
  { id: "ai", label: "AI Model", icon: Sparkles, ready: true },
  { id: "privacy", label: "Privacy", icon: Shield, ready: false },
  { id: "export", label: "Export", icon: Download, ready: true },
];
