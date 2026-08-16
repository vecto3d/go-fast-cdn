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

// One request per batch instead of one per file. The server caps a batch, so
// anything larger is split; the batches run in sequence because the point is to
// stop flooding the server, not to move the flood to a different endpoint.
const BATCH_SIZE = 250;

export type BulkDeleteResult = {
  deleted: number;
  failed: { file: string; error: string }[];
};

export const useBulkDeleteMutation = (type: TFileType) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      files,
      folder = "",
    }: {
      files: string[];
      folder?: string;
    }): Promise<BulkDeleteResult> => {
      const result: BulkDeleteResult = { deleted: 0, failed: [] };

      for (let start = 0; start < files.length; start += BATCH_SIZE) {
        const batch = files.slice(start, start + BATCH_SIZE);
        const res = await cdnApiClient.post(
          `/delete/${apiSegment(type)}/bulk`,
          { folder, files: batch }
        );
        result.deleted += res.data.deleted ?? 0;
        result.failed.push(...(res.data.failed ?? []));
      }

      return result;
    },
    onSuccess: (result) => {
      // A single summary, not one toast per file.
      toast.success(
        result.deleted === 1
          ? "Deleted 1 file."
          : `Deleted ${result.deleted} files.`
      );
      if (result.failed.length > 0) {
        toast.error(
          result.failed.length === 1
            ? `${result.failed[0].file}: ${result.failed[0].error}`
            : `${result.failed.length} files could not be deleted.`
        );
      }

      queryClient.invalidateQueries({ queryKey: constant.queryKeys.size() });
      queryClient.invalidateQueries({ queryKey: constant.queryKeys.images(type) });
    },
    onError: (error) => {
      const err = error as AxiosError<IErrorResponse>;
      toast.error(err.response?.data?.error || err.message || "Delete failed");
    },
  });
};
