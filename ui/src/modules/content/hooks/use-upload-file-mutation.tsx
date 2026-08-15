import { apiSegment, TFileType } from "@/lib/file-types";
import { cdnApiClient } from "@/services/authService";
import { useMutation } from "@tanstack/react-query";

// The multipart field the API expects for each type.
const FORM_FIELD = {
  images: "image",
  docs: "doc",
  audio: "audio",
} as const;

const useUploadFileMutation = () => {
  return useMutation({
    mutationFn: async (payload: {
      file: File;
      type: TFileType;
      folder?: string;
    }) => {
      const segment = apiSegment(payload.type);
      const form = new FormData();
      form.append(FORM_FIELD[segment], payload.file);
      form.append("folder", payload.folder ?? "");
      const res = await cdnApiClient.post(`/upload/${segment}`, form, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });
      return res.data;
    },
  });
};

export default useUploadFileMutation;
