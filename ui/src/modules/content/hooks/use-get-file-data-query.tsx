import { constant } from "@/lib/constant";
import { cdnApiClient } from "@/services/authService";
import { FileMetadata } from "@/types/fileMetadata";
import { useQuery } from "@tanstack/react-query";

type FileDataParams = {
  filename: string;
  folder?: string;
  type: "documents" | "images";
};

const useGetFileDataQuery = ({
  filename,
  folder = "",
  type,
}: FileDataParams) => {
  return useQuery({
    queryKey: constant.queryKeys.image(folder ? `${folder}/${filename}` : filename),
    queryFn: async (): Promise<FileMetadata> => {
      const res = await cdnApiClient.get(
        `/${type === "documents" ? "doc" : "image"}/${encodeURIComponent(
          filename
        )}`,
        { params: { folder } }
      );
      return res.data;
    },
  });
};

export default useGetFileDataQuery;
