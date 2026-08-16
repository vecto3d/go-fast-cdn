import { TContentCardProps } from "@/types/contentCard";
import { DownloadCloud, FileText, Files, Music, Trash2 } from "lucide-react";
import { toast } from "react-hot-toast";
import FileDataModal from "./file-data-modal";
import RenameModal from "./rename-modal";
import ResizeModal from "./resize-modal";
import useDeleteFileMutation from "./hooks/use-delete-file-mutation";
import { Tooltip, TooltipContent } from "@/components/ui/tooltip";
import { TooltipTrigger } from "@radix-ui/react-tooltip";
import { Dialog, DialogTrigger } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { encodeFilePath } from "@/utils";
import { apiSegment, isResizable } from "@/lib/file-types";
import { cn } from "@/lib/utils";
import { useState } from "react";

const ContentCard: React.FC<TContentCardProps> = ({
  file_name,
  folder = "",
  type = "documents",
  disabled = false,
  isSelected,
  onSelect,
  isSelecting,
  selectionKey,
}) => {
  const [isDetailOpen, setIsDetailOpen] = useState(false);

  const url = `${window.location.protocol}//${
    window.location.host
  }/api/cdn/download/${apiSegment(type)}/${encodeFilePath(folder, file_name)}`;

  const deleteFile = useDeleteFileMutation(type);

  const handleDeleteFile = () => {
    deleteFile.mutate({ filename: file_name, folder });
  };

  // While selecting, the whole card is the hit area — clicking a 16px checkbox
  // 200 times is the kind of thing that makes bulk selection not worth using.
  const handleCardClick = (event: React.MouseEvent) => {
    if (!isSelecting || !onSelect) return;
    onSelect(file_name, { shiftKey: event.shiftKey });
  };

  return (
    <div
      data-selection-key={selectionKey ?? file_name}
      onClick={handleCardClick}
      className={cn(
        "border rounded-lg shadow-lg flex flex-col min-h-[264px] w-64 max-w-[256px] justify-between items-center gap-4 p-4 relative transition-colors",
        {
          "cursor-pointer select-none": isSelecting,
          "ring-2 ring-primary bg-accent": isSelecting && isSelected,
        }
      )}
    >
      {isSelecting && (
        <Checkbox
          className="absolute top-2 right-2 bg-background pointer-events-none"
          checked={isSelected}
          disabled={disabled}
          tabIndex={-1}
          aria-label="Select file"
        />
      )}
      {type === "audio" && (
        // Sits outside the dialog trigger so its controls stay clickable.
        <audio
          controls
          preload="none"
          src={url}
          className="w-full"
          aria-label={`Play ${file_name}`}
        />
      )}
      {type === "video" && (
        <video
          controls
          preload="metadata"
          src={url}
          className="max-h-[150px] max-w-[224px]"
          aria-label={`Play ${file_name}`}
        />
      )}
      <Dialog open={isDetailOpen} onOpenChange={setIsDetailOpen}>
        <DialogTrigger disabled={isSelecting}>
          {type === "images" && (
            <img
              src={url}
              alt={file_name}
              width={224}
              height={150}
              loading="lazy"
              className="object-cover max-h-[150px] max-w-[224px]"
            />
          )}
          {type === "documents" && <FileText size="128" />}
          {type === "audio" && <Music size="96" />}
        </DialogTrigger>
        <FileDataModal
          filename={file_name}
          folder={folder}
          type={type}
          isOpen={isDetailOpen}
        />
      </Dialog>
      <div className="w-full flex flex-col gap-2">
        <p className="truncate">{file_name}</p>
        {/* Non-destructive buttons */}
        <div className={`flex w-full justify-between ${disabled && "sr-only"}`}>
          <div className="flex">
            <Tooltip>
              <TooltipTrigger>
                <Button
                  variant="ghost"
                  size="icon"
                  className="text-sky-600"
                  onClick={() => {
                    if (location.protocol == 'https:') {
                      navigator.clipboard.writeText(url);
                      toast.success("clipboard saved");
                    } else {
                      //https://stackoverflow.com/questions/72237719/not-being-able-to-copy-url-to-clipboard-without-adding-the-protocol-https
                      const textArea = document.createElement("textarea");
                      textArea.value = url;
                      document.body.appendChild(textArea);
                      textArea.focus({ preventScroll: true });
                      textArea.select();
                      document.execCommand('copy');
                      document.body.removeChild(textArea);
                      toast.success("clipboard saved");
                    }
                  }}
                  aria-label="Copy Link"
                  disabled={isSelecting}
                >
                  <Files />
                </Button>
              </TooltipTrigger>
              <TooltipContent side="bottom">
                <p>Copy Link to clipboard</p>
              </TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger>
                <Button
                  variant="ghost"
                  size="icon"
                  className="text-sky-600"
                  disabled={isSelecting}
                  asChild={!isSelecting}
                >
                  <a href={url} download aria-label="Download file">
                    <DownloadCloud />
                  </a>
                </Button>
              </TooltipTrigger>
              <TooltipContent side="bottom">
                <p>Download file</p>
              </TooltipContent>
            </Tooltip>
            <RenameModal
              type={type}
              filename={file_name}
              folder={folder}
              isSelecting={isSelecting}
            />
            {/* Resize re-encodes the image, so it only appears for the formats
                the server has an encoder for. */}
            {type === "images" && isResizable(file_name) && (
              <ResizeModal
                filename={file_name ?? ""}
                folder={folder}
                isSelecting={isSelecting}
              />
            )}
          </div>
          {/* Destructive buttons */}
          <div className="flex gap-2">
            <Tooltip>
              <TooltipTrigger>
                <Button
                  variant="destructive"
                  size="icon"
                  onClick={() => file_name && handleDeleteFile()}
                  aria-label="Delete file"
                  disabled={isSelecting}
                >
                  <Trash2 className="inline" size="24" />
                </Button>
              </TooltipTrigger>
              <TooltipContent side="bottom">
                <p>Delete file</p>
              </TooltipContent>
            </Tooltip>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ContentCard;
