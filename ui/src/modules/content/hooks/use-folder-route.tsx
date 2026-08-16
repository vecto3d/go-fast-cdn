import { FILE_TYPES, TFileType } from "@/lib/file-types";
import { sanitizeFolder } from "@/utils";
import { useCallback } from "react";
import { useLocation } from "wouter";

/**
 * Keeps the current folder in the URL (/images/logos/dark) rather than in
 * component state, so folders are linkable and the browser's back button walks
 * back up the tree instead of leaving the page.
 */
export const useFolderRoute = (type: TFileType) => {
  const [location, navigate] = useLocation();
  const base = FILE_TYPES[type].route;

  const path = location.startsWith(base) ? location.slice(base.length) : "";
  const folder = sanitizeFolder(
    path
      .split("/")
      .map((segment) => decodeURIComponent(segment))
      .join("/")
  );

  const openFolder = useCallback(
    (next: string) => {
      const target = sanitizeFolder(next);
      navigate(
        target
          ? `${base}/${target.split("/").map(encodeURIComponent).join("/")}`
          : base
      );
    },
    [base, navigate]
  );

  return { folder, openFolder };
};
