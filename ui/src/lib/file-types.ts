// One place describing every file type the CDN stores. The UI names ("documents")
// and the API segments ("docs") differ, so the mapping lives here rather than in
// a conditional at each call site.
export type TFileType = "images" | "documents" | "audio" | "video";

type FileTypeConfig = {
  label: string;
  /** URL segment the API uses: /api/cdn/upload/<api> */
  api: "images" | "docs" | "audio" | "video";
  /** Route this type's page lives at */
  route: string;
  /** accept attribute for the file input, mirroring the server's allowlist */
  accept: string;
  dropMessage: string;
};

export const FILE_TYPES: Record<TFileType, FileTypeConfig> = {
  images: {
    label: "Images",
    api: "images",
    route: "/images",
    accept:
      ".jpg,.jpeg,.png,.gif,.webp,.bmp,.svg,.avif,.heic,.heif,.tif,.tiff,.ico",
    dropMessage: "Drop your images here, or click to select files.",
  },
  documents: {
    label: "Documents",
    api: "docs",
    route: "/documents",
    accept: ".txt,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.pdf,.rtf,.arc,.zip",
    dropMessage: "Drop your documents here, or click to select files.",
  },
  audio: {
    label: "Audio",
    api: "audio",
    route: "/audio",
    accept: ".mp3,.wav,.ogg,.oga,.flac,.m4a,.aac,.aiff,.mid,.midi",
    dropMessage: "Drop your audio files here, or click to select files.",
  },
  video: {
    label: "Video",
    api: "video",
    route: "/video",
    accept: ".mp4,.webm,.mkv,.mov,.m4v,.avi",
    dropMessage: "Drop your video files here, or click to select files.",
  },
};

export const FILE_TYPE_NAMES = Object.keys(FILE_TYPES) as TFileType[];

export const apiSegment = (type: TFileType) => FILE_TYPES[type].api;

/** The extensions this type accepts, lowercased and including the dot. */
export const acceptedExtensions = (type: TFileType) =>
  FILE_TYPES[type].accept.split(",");

export const extensionOf = (fileName: string) => {
  const dot = fileName.lastIndexOf(".");
  return dot < 0 ? "" : fileName.slice(dot).toLowerCase();
};

export const isAccepted = (type: TFileType, fileName: string) =>
  acceptedExtensions(type).includes(extensionOf(fileName));

// The resize endpoint re-encodes the image, which it can only do for the
// formats it has an encoder for. Anything else has no resize control.
const RESIZABLE = [".png", ".jpg", ".jpeg", ".bmp", ".webp"];

export const isResizable = (fileName: string) =>
  RESIZABLE.includes(extensionOf(fileName));
