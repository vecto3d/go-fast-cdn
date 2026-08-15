// One place describing every file type the CDN stores. The UI names ("documents")
// and the API segments ("docs") differ, so the mapping lives here rather than in
// a conditional at each call site.
export type TFileType = "images" | "documents" | "audio";

type FileTypeConfig = {
  label: string;
  /** URL segment the API uses: /api/cdn/upload/<api> */
  api: "images" | "docs" | "audio";
  /** accept attribute for the file input */
  accept: string;
  dropMessage: string;
};

export const FILE_TYPES: Record<TFileType, FileTypeConfig> = {
  images: {
    label: "Images",
    api: "images",
    accept: "image/jpeg,image/png,image/jpg,image/webp,image/gif,image/bmp",
    dropMessage: "Drop your images here, or click to select files.",
  },
  documents: {
    label: "Documents",
    api: "docs",
    accept:
      "text/plain,application/zip,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.openxmlformats-officedocument.presentationml.presentation,application/pdf,application/rtf,application/x-freearc",
    dropMessage: "Drop your documents here, or click to select files.",
  },
  audio: {
    label: "Audio",
    api: "audio",
    accept: "audio/mpeg,audio/wav,audio/ogg,audio/flac,audio/mp4,audio/aac",
    dropMessage: "Drop your audio files here, or click to select files.",
  },
};

export const FILE_TYPE_NAMES = Object.keys(FILE_TYPES) as TFileType[];

/** The MIME types a dropped file must match for this type, as a list. */
export const acceptedTypes = (type: TFileType) => FILE_TYPES[type].accept.split(",");

export const apiSegment = (type: TFileType) => FILE_TYPES[type].api;
