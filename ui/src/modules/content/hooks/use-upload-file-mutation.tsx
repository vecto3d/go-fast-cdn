import { apiSegment, TFileType } from "@/lib/file-types";
import { cdnApiClient } from "@/services/authService";
import { useMutation } from "@tanstack/react-query";

// The multipart field the API expects for each type.
const FORM_FIELD = {
  images: "image",
  docs: "doc",
  audio: "audio",
  video: "video",
} as const;

// How many uploads are in flight at once. Firing every file at the browser
// simultaneously is what made large batches fail: browsers cap connections per
// host and queue the rest, memory for all the request bodies is held at once,
// and one failure took the whole Promise.all down with it.
const CONCURRENCY = 4;

export type UploadOutcome = {
  file: File;
  ok: boolean;
  error?: string;
};

const errorText = (error: unknown) => {
  const response = (error as { response?: { data?: unknown } }).response?.data;
  if (typeof response === "string") return response;
  if (response && typeof response === "object" && "error" in response) {
    return String((response as { error: unknown }).error);
  }
  return (error as Error).message || "Upload failed";
};

export const uploadFile = async (
  file: File,
  type: TFileType,
  folder: string
) => {
  const segment = apiSegment(type);
  const form = new FormData();
  form.append(FORM_FIELD[segment], file);
  form.append("folder", folder);

  const res = await cdnApiClient.post(`/upload/${segment}`, form, {
    headers: { "Content-Type": "multipart/form-data" },
  });

  return res.data;
};

/**
 * Uploads every file, at most CONCURRENCY at a time, reporting progress as it
 * goes. Never rejects: each file's outcome is returned so one bad file cannot
 * discard the rest of the batch.
 */
export const uploadAll = async (
  files: File[],
  type: TFileType,
  folder: string,
  onProgress?: (done: number, total: number) => void
): Promise<UploadOutcome[]> => {
  const outcomes: UploadOutcome[] = new Array(files.length);
  let next = 0;
  let done = 0;

  const worker = async () => {
    while (next < files.length) {
      const index = next++;
      const file = files[index];
      try {
        await uploadFile(file, type, folder);
        outcomes[index] = { file, ok: true };
      } catch (error) {
        outcomes[index] = { file, ok: false, error: errorText(error) };
      }
      onProgress?.(++done, files.length);
    }
  };

  await Promise.all(
    Array.from({ length: Math.min(CONCURRENCY, files.length) }, worker)
  );

  return outcomes;
};

const useUploadFileMutation = () => {
  return useMutation({
    mutationFn: async (payload: {
      file: File;
      type: TFileType;
      folder?: string;
    }) => uploadFile(payload.file, payload.type, payload.folder ?? ""),
  });
};

export default useUploadFileMutation;
