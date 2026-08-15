import { TFileType } from "./file-types";

export const constant = {
  queryKeys: {
    all: [{ entity: "cdn" }] as const,
    size: () => [{ ...constant.queryKeys.all[0], scope: "size" }] as const,
    image: (filename: string) =>
      [{ ...constant.queryKeys.all[0], scope: "file-data", filename }] as const,
    dimensions: (height: number, width: number) =>
      [{ ...constant.queryKeys.all[0], scope: "aaaa", height, width }] as const,
    images: (type: TFileType) =>
      [{ ...constant.queryKeys.all[0], scope: "files", type }] as const,
    folders: (type: TFileType) =>
      [{ ...constant.queryKeys.all[0], scope: "folders", type }] as const,
    users: "users",
    registrationEnabled: "registration-enabled",
    dashboard: "dashboard",
  },
};
