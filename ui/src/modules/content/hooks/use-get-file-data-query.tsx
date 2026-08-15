import { constant } from "@/lib/constant";
import { apiSegment, TFileType } from "@/lib/file-types";
import { cdnApiClient } from "@/services/authService";
import { FileMetadata } from "@/types/fileMetadata";
import { useQuery } from "@tanstack/react-query";

type FileDataParams = {
  filename: string;
  folder?: string;
  type: TFileType;
};

const useGetFileDataQuery = ({
  filename,
  folder = "",
  type,
}: FileDataParams) => {
  return useQuery({
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
