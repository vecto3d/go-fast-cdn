import { createContext } from "react";

export type Theme = "light" | "dark" | "system";
export type ResolvedTheme = "light" | "dark";

/**
 * Keep this key in sync with the inline bootstrap script in `index.html`,
 * which applies the stored theme before React mounts to avoid a flash of
 * the wrong colour scheme.
 */
export const THEME_STORAGE_KEY = "go-fast-cdn-theme";

export const DARK_MEDIA_QUERY = "(prefers-color-scheme: dark)";

export const isTheme = (value: unknown): value is Theme =>
  value === "light" || value === "dark" || value === "system";

/** Reads the persisted preference, falling back to "system". */
export const getStoredTheme = (): Theme => {
  try {
    const stored = window.localStorage.getItem(THEME_STORAGE_KEY);
    return isTheme(stored) ? stored : "system";
  } catch {
    // localStorage can throw in private-browsing / sandboxed contexts.
    return "system";
  }
};

export const storeTheme = (theme: Theme): void => {
  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, theme);
  } catch {
    // Persisting is best-effort; the in-memory preference still applies.
  }
};

export const getSystemTheme = (): ResolvedTheme =>
  window.matchMedia(DARK_MEDIA_QUERY).matches ? "dark" : "light";

/** Toggles the `dark` class that `index.css` keys its token overrides off. */
export const applyTheme = (resolved: ResolvedTheme): void => {
  const root = document.documentElement;
  root.classList.toggle("dark", resolved === "dark");
  // Lets the browser theme scrollbars, form controls and the like.
  root.style.colorScheme = resolved;
};

export interface ThemeContextValue {
  /** The user's preference, which may be "system". */
  theme: Theme;
  /** The concrete theme currently applied to the document. */
  resolvedTheme: ResolvedTheme;
  setTheme: (theme: Theme) => void;
}

export const ThemeContext = createContext<ThemeContextValue | undefined>(
  undefined
);
