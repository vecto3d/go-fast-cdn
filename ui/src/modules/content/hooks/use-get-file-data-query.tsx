import { constant } from "@/lib/constant";
import { apiSegment, TFileType } from "@/lib/file-types";
import { cdnApiClient } from "@/services/authService";
import { FileMetadata } from "@/types/fileMetadata";
import { useQuery } from "@tanstack/react-query";

type FileDataParams = {
  filename: string;
  folder?: string;
  type: TFileType;
  /** Off by default: metadata is only worth fetching once something shows it. */
  enabled?: boolean;
};

const useGetFileDataQuery = ({
  filename,
  folder = "",
  type,
  enabled = true,
}: FileDataParams) => {
  return useQuery({
    enabled,
    queryKey: constant.queryKeys.image(
      folder ? `${folder}/${filename}` : filename
    ),
    queryFn: async (): Promise<FileMetadata> => {
      const res = await cdnApiClient.get(
        `/${apiSegment(type)}/${encodeURIComponent(filename)}`,
        { params: { folder } }
      );
      return res.data;
    },
  });
};

export default useGetFileDataQuery;
