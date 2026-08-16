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
  /** Identifies this card to the drag-select box; defaults to the file name. */
  selectionKey?: string;
  onSelect?: (fileName: string, modifiers?: { shiftKey?: boolean }) => void;
  isSelecting?: boolean;
};
