import { TFile } from "@/types/file";

export type SortKey = "name" | "date" | "size";
export type SortDirection = "asc" | "desc";

export type Sort = {
  key: SortKey;
  direction: SortDirection;
};

export const DEFAULT_SORT: Sort = { key: "name", direction: "asc" };

export const SORT_LABELS: Record<SortKey, string> = {
  name: "Name",
  date: "Date",
  size: "Size",
};

// What each direction means depends on the column: A→Z for names, but newest
// and largest first are the useful defaults for dates and sizes, so the labels
// say what you get rather than "ascending".
export const SORT_DIRECTION_LABELS: Record<
  SortKey,
  Record<SortDirection, string>
> = {
  name: { asc: "A → Z", desc: "Z → A" },
  date: { asc: "Oldest first", desc: "Newest first" },
  size: { asc: "Smallest first", desc: "Largest first" },
};

/** Sorts a copy of the files; the input array is left alone. */
export const sortFiles = (files: TFile[], sort: Sort): TFile[] => {
  const compare = (a: TFile, b: TFile) => {
    switch (sort.key) {
      case "date":
        return (
          new Date(a.CreatedAt).getTime() - new Date(b.CreatedAt).getTime()
        );
      case "size":
        return (a.size ?? 0) - (b.size ?? 0);
      default:
        // Locale-aware and numeric, so icon2 sorts before icon10.
        return a.file_name.localeCompare(b.file_name, undefined, {
          numeric: true,
          sensitivity: "base",
        });
    }
  };

  return [...files].sort((a, b) =>
    sort.direction === "asc" ? compare(a, b) : -compare(a, b)
  );
};
