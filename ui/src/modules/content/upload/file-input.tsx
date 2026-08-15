import { TabsContent } from "@/components/ui/tabs";
import { FILE_TYPES, TFileType } from "@/lib/file-types";
import React from "react";

interface FileInputProps {
  type: TFileType;
  fileRef: React.RefObject<HTMLInputElement | null>;
  onFileChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const FileInput = ({ type, fileRef, onFileChange }: FileInputProps) => {
  const { accept, dropMessage } = FILE_TYPES[type];

  return (
    <TabsContent value={type}>
      <p className="text-muted-foreground text-center">{dropMessage}</p>
      <input
        type="file"
        accept={accept}
        multiple
        name={type}
        id={type}
        aria-label={`Select ${type}`}
        ref={fileRef}
        className="hidden"
        onChange={onFileChange}
      />
    </TabsContent>
  );
};

export default FileInput;
