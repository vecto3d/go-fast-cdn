// Encodes a folder + filename into a URL path. Each segment is encoded on its
// own so that the "/" separators survive, while spaces and other reserved
// characters in the names themselves don't break the link.
export const encodeFilePath = (folder: string, fileName: string) =>
  [...folder.split("/"), fileName]
    .filter(Boolean)
    .map(encodeURIComponent)
    .join("/");

// Mirrors the backend's SanitizeFolder: drops empty and traversal segments so
// what we display, send and compare is always the canonical form.
export const sanitizeFolder = (folder: string) =>
  folder
    .replace(/\\/g, "/")
    .split("/")
    .map((segment) => segment.trim())
    .filter((segment) => segment !== "" && segment !== "." && segment !== "..")
    .join("/");

export const sanitizeFileName = (file: File) => {
  const ext = file.name.split(".").pop();
  const base = file.name.replace(/\.[^/.]+$/, "").replace(/\./g, "-");
  const newName = `${base}.${ext}`;
  return new File([file], newName, { type: file.type });
};
