import { CATEGORY_PALETTE } from "@/lib/category/colors";
import { cn } from "@/lib/utils";

export function ColorPaletteSwatches({
  value,
  label,
  disabled,
  pending,
  onSelect,
}: {
  value: string;
  label: string;
  disabled?: boolean;
  pending?: boolean;
  onSelect: (color: string) => void;
}) {
  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {CATEGORY_PALETTE.map((color) => {
        const selected = value.toUpperCase() === color.toUpperCase();

        return (
          <button
            key={color}
            type="button"
            aria-label={`Set ${label} to ${color}`}
            disabled={disabled || pending}
            onClick={() => {
              if (!selected) {
                onSelect(color);
              }
            }}
            className={cn(
              "size-6 rounded-full border-4 transition-transform hover:scale-105 disabled:opacity-50",
              selected
                ? "border-foreground ring-2 ring-foreground/20"
                : "border-transparent",
            )}
            style={{ backgroundColor: color }}
          />
        );
      })}
    </div>
  );
}
