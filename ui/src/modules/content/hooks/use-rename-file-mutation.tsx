import { apiSegment, TFileType } from "@/lib/file-types";
import { cdnApiClient } from "@/services/authService";
import { useMutation, UseMutationOptions } from "@tanstack/react-query";

const useRenameFileMutation = (
  type: TFileType,
  options?: UseMutationOptions<string, Error, FormData>
) => {
  return useMutation({
    mutationFn: async (formData: FormData) => {
      const res = await cdnApiClient.put(
        `/rename/${apiSegment(type)}`,
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        }
      );
      return res.data;
    },
    ...options,
  });
};

export default useRenameFileMutation;
