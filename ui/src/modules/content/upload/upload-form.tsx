import { sanitizeFileName } from "@/utils";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import ImageCardUpload from "./image-card-upload";
import FileInput from "./file-input";
import toast from "react-hot-toast";
import DocCardUpload from "./doc-card-upload";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  FILE_TYPE_NAMES,
  FILE_TYPES,
  isAccepted,
  TFileType,
} from "@/lib/file-types";

interface UploadProps {
  files: File[];
  onChangeFiles: (files: File[]) => void;
  isLoading: boolean;
  tab: TFileType;
  onChangeTab: (tab: TFileType) => void;
  disableTabSwitching?: boolean;
}

const UploadForm = ({
  isLoading,
  files,
  onChangeFiles,
  onChangeTab,
  tab,
  disableTabSwitching = false,
}: UploadProps) => {
  const fileRef = useRef<HTMLInputElement>(null);
  const [isDragOver, setIsDragOver] = useState(false);

  // One object URL per file, released when the selection changes. Creating
  // these inline during render leaked one per render per file, which is what
  // made a large batch heavy before it even started uploading.
  const previews = useMemo(
    () => (tab === "images" ? files.map((file) => URL.createObjectURL(file)) : []),
    [files, tab]
  );

  useEffect(
    () => () => previews.forEach((preview) => URL.revokeObjectURL(preview)),
    [previews]
  );

  const handleOnChangeFiles = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFiles = e.target.files;
      if (selectedFiles) {
        const fileArray = Array.from(selectedFiles);
        onChangeFiles(fileArray);
      }
    },
    [onChangeFiles]
  );

  const handleDeleteFile = useCallback(
    (index: number) => {
      onChangeFiles(files.filter((_, i) => i !== index));
    },
    [files, onChangeFiles]
  );

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragOver(false);

      if (isLoading) return;

      const droppedFiles = Array.from(e.dataTransfer.files);

      // Check by extension: browsers report nothing useful for heic, avif or
      // some svg files, and the server validates by extension too.
      const isValidFiles = droppedFiles.every((file) =>
        isAccepted(tab, file.name)
      );
      if (!isValidFiles) {
        toast.error(`Invalid file type. Please upload ${tab} only.`);
        return;
      }
      if (droppedFiles.length === 0) return;
      onChangeFiles([...files, ...droppedFiles]);
    },
    [isLoading, onChangeFiles, files, tab]
  );

  const handleClick = useCallback(() => {
    if (fileRef.current && files.length === 0 && !isLoading) {
      fileRef.current.click();
    }
  }, [files.length, isLoading]);

  return (
    <div
      id="drop-zone"
      className={cn(
        "w-full h-96 border-2 border-dashed rounded-md transition-colors",
        {
          "cursor-pointer": files.length === 0 && !isLoading,
          "border-zinc-300": !isDragOver,
          "border-blue-400 bg-blue-50": isDragOver,
          "opacity-50": isLoading,
        }
      )}
      onClick={handleClick}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      {files.length > 0 ? (
        <div className="flex gap-2 flex-wrap p-2 overflow-y-auto max-h-80">
          {files.map((file, index) =>
            tab === "images" ? (
              <ImageCardUpload
                key={file.name + index}
                fileName={sanitizeFileName(file).name}
                onClickDelete={() => handleDeleteFile(index)}
                imageUrl={previews[index]}
              />
            ) : (
              <DocCardUpload
                key={file.name + index}
                fileName={sanitizeFileName(file).name}
                onClickDelete={() => handleDeleteFile(index)}
              />
            )
          )}
        </div>
      ) : (
        <div className="flex flex-col justify-center items-center h-full">
          {disableTabSwitching ? (
            // When tab switching is disabled, show content without tabs
            <>
              {isDragOver ? (
                <div className="text-center">
                  <p className="text-blue-600 font-medium">
                    Drop files here to upload
                  </p>
                </div>
              ) : (
                <div className="text-center">
                  <p className="text-muted-foreground text-center">
                    {FILE_TYPES[tab].dropMessage}
                  </p>
                  <input
                    type="file"
                    accept={FILE_TYPES[tab].accept}
                    multiple
                    name={tab}
                    id={tab}
                    aria-label={`Select ${tab}`}
                    ref={fileRef}
                    className="hidden"
                    onChange={handleOnChangeFiles}
                  />
                </div>
              )}
            </>
          ) : (
            // When tab switching is enabled, show tabs with content
            <Tabs
              onValueChange={(value) => {
                onChangeTab(value as TFileType);
              }}
              value={tab}
            >
              <TabsList
                className="self-center mb-4"
                onClick={(e) => e.stopPropagation()}
              >
                {FILE_TYPE_NAMES.map((name) => (
                  <TabsTrigger key={name} value={name}>
                    {FILE_TYPES[name].label}
                  </TabsTrigger>
                ))}
              </TabsList>

              {isDragOver ? (
                <div className="text-center">
                  <p className="text-blue-600 font-medium">
                    Drop files here to upload
                  </p>
                </div>
              ) : (
                <>
                  {FILE_TYPE_NAMES.map((name) => (
                    <FileInput
                      key={name}
                      type={name}
                      fileRef={fileRef}
                      onFileChange={handleOnChangeFiles}
                    />
                  ))}
                </>
              )}
            </Tabs>
          )}
        </div>
      )}
    </div>
  );
};

export default UploadForm;
