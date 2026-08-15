import { Monitor, Moon, Sun } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { useTheme } from "@/hooks/use-theme";
import { cn } from "@/lib/utils";
import type { ResolvedTheme, Theme } from "@/lib/theme";

const THEME_OPTIONS: { value: Theme; label: string; icon: typeof Sun }[] = [
  { value: "light", label: "Light", icon: Sun },
  { value: "dark", label: "Dark", icon: Moon },
  { value: "system", label: "System", icon: Monitor },
];

/** Mirrors what is actually on screen, so "system" shows the real state. */
const ThemeIcon = ({ resolved }: { resolved: ResolvedTheme }) =>
  resolved === "dark" ? <Moon /> : <Sun />;

interface ThemeOptionsProps {
  theme: Theme;
  onSelect: (theme: Theme) => void;
}

const ThemeOptions = ({ theme, onSelect }: ThemeOptionsProps) => (
  <DropdownMenuRadioGroup
    value={theme}
    onValueChange={(value) => onSelect(value as Theme)}
  >
    {THEME_OPTIONS.map(({ value, label, icon: Icon }) => (
      <DropdownMenuRadioItem key={value} value={value}>
        <Icon className="mr-2 size-4" />
        {label}
      </DropdownMenuRadioItem>
    ))}
  </DropdownMenuRadioGroup>
);

/**
 * Standalone theme switcher, for screens rendered outside the sidebar
 * (login, register, access-denied).
 */
export function ThemeToggle({ className }: { className?: string }) {
  const { theme, resolvedTheme, setTheme } = useTheme();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className={cn(className)}
          aria-label={`Change theme (currently ${theme})`}
        >
          <ThemeIcon resolved={resolvedTheme} />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <ThemeOptions theme={theme} onSelect={setTheme} />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

/** Theme switcher styled to sit in the sidebar footer. */
export function SidebarThemeToggle() {
  const { state } = useSidebar();
  const { theme, resolvedTheme, setTheme } = useTheme();

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          {/* asChild so the trigger *is* the menu button rather than
              wrapping one — nesting <button> inside <button> is invalid. */}
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton tooltip="Theme" aria-label="Theme">
              <ThemeIcon resolved={resolvedTheme} />
              <span>Theme</span>
              <span className="ml-auto text-xs capitalize text-muted-foreground">
                {theme}
              </span>
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            side={state === "collapsed" ? "right" : "top"}
            align="end"
          >
            <ThemeOptions theme={theme} onSelect={setTheme} />
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
