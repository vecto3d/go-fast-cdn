export type TFile = {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  file_name: string;
  folder: string;
  /** Bytes on disk, read while listing. */
  size?: number;
  checksum: string;
};
