import { constant } from "@/lib/constant";
import { apiSegment, TFileType } from "@/lib/file-types";
import { cdnApiClient } from "@/services/authService";
import { IErrorResponse } from "@/types/response";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { AxiosError } from "axios";
import toast from "react-hot-toast";

const useDeleteFileMutation = (type: TFileType) => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      filename,
      folder = "",
    }: {
      filename: string;
      folder?: string;
    }) => {
      const res = await cdnApiClient.delete(
        `/delete/${apiSegment(type)}/${encodeURIComponent(filename)}`,
        { params: { folder } }
      );
      return res.data;
    },
    onSuccess: () => {
      toast.dismiss();
      toast.success("Successfully deleted file!");
      queryClient.invalidateQueries({
        queryKey: constant.queryKeys.size(),
      });
      queryClient.invalidateQueries({
        queryKey: constant.queryKeys.images(type),
      });
    },
    onError: (error) => {
      const err = error as AxiosError<IErrorResponse>;
      toast.dismiss();
      const message =
        err.response?.data?.error || err.message || "Delete failed";
      toast.error(message);
    },
  });
};

export default useDeleteFileMutation;
