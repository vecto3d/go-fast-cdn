import { cdnApiClient } from "@/services/authService";
import { useMutation } from "@tanstack/react-query";

const useUploadFileMutation = () => {
  return useMutation({
    mutationFn: async (payload: {
      file: File;
      type: "doc" | "image";
      folder?: string;
    }) => {
      const form = new FormData();
      form.append(payload.type, payload.file);
      form.append("folder", payload.folder ?? "");
      const res = await cdnApiClient.post(`/upload/${payload.type}`, form, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });
      return res.data;
    },
  });
};

export default useUploadFileMutation;
