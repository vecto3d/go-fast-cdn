import { constant } from "@/lib/constant";
import { apiSegment, TFileType } from "@/lib/file-types";
import { cdnApiClient } from "@/services/authService";
import { TFile } from "@/types/file";
import { useQuery } from "@tanstack/react-query";

type GetFilesParams = {
  type: TFileType;
};

const useGetFilesQuery = ({ type }: GetFilesParams) => {
  return useQuery({
    queryKey: constant.queryKeys.images(type),
    queryFn: async () => {
      const res = await cdnApiClient.get<TFile[]>(`/${apiSegment(type)}/all`);
      return res.data;
    },
  });
};

export default useGetFilesQuery;
