import { TFileType } from "@/lib/file-types";

export type TContentCardProps = {
  file_name: string;
  folder?: string;
  type?: TFileType;
  ID?: number;
  createdAt: string;
  updatedAt: string;
  disabled?: boolean;
  isSelected?: boolean;
  onSelect?: (fileName: string) => void;
  isSelecting?: boolean;
};
