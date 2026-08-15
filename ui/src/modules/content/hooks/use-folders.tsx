import { constant } from "@/lib/constant";
import { apiSegment, TFileType } from "@/lib/file-types";
import { cdnApiClient } from "@/services/authService";
import { IErrorResponse } from "@/types/response";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AxiosError } from "axios";
import toast from "react-hot-toast";

const errorMessage = (error: Error, fallback: string) =>
  (error as AxiosError<IErrorResponse>).response?.data?.error || fallback;

/** Every folder that exists on disk, including ones with no files yet. */
export const useFoldersQuery = (type: TFileType) =>
  useQuery({
    queryKey: constant.queryKeys.folders(type),
    queryFn: async () => {
      const res = await cdnApiClient.get<string[]>(`/folder/${apiSegment(type)}`);
      return res.data;
    },
  });

export const useCreateFolderMutation = (type: TFileType) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (folder: string) => {
      const res = await cdnApiClient.post(`/folder/${apiSegment(type)}`, {
        folder,
      });
      return res.data;
    },
    onSuccess: () => {
      toast.success("Folder created!");
      queryClient.invalidateQueries({
        queryKey: constant.queryKeys.folders(type),
      });
    },
    onError: (error) => {
      toast.error(errorMessage(error, "Failed to create folder"));
    },
  });
};

export const useDeleteFolderMutation = (type: TFileType) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (folder: string) => {
      const res = await cdnApiClient.delete(`/folder/${apiSegment(type)}`, {
        params: { folder },
      });
      return res.data as { files_deleted: number };
    },
    onSuccess: (data) => {
      toast.success(
        data.files_deleted === 1
          ? "Folder deleted, along with 1 file."
          : `Folder deleted, along with ${data.files_deleted} files.`
      );
      queryClient.invalidateQueries({
        queryKey: constant.queryKeys.folders(type),
      });
      queryClient.invalidateQueries({
        queryKey: constant.queryKeys.images(type),
      });
      queryClient.invalidateQueries({ queryKey: constant.queryKeys.size() });
    },
    onError: (error) => {
      toast.error(errorMessage(error, "Failed to delete folder"));
    },
  });
};
