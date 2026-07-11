import { useEffect } from "react";
import { useJsonSetting } from "./useJsonSetting";

export type ThemeSetting = "system" | "light" | "dark";

export function useConfiguredTheme() {
  const theme = useJsonSetting<ThemeSetting>("app.theme", "system");

  useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)");

    const applyTheme = () => {
      const resolvedTheme =
        theme.value === "system"
          ? media.matches
            ? "dark"
            : "light"
          : theme.value;
      document.documentElement.classList.toggle(
        "dark",
        resolvedTheme === "dark",
      );
    };

    applyTheme();
    media.addEventListener("change", applyTheme);

    return () => {
      media.removeEventListener("change", applyTheme);
    };
  }, [theme.value]);
}
