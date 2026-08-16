import { Loader2Icon, Plus } from "lucide-react";
import { useCallback, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { sanitizeFileName, sanitizeFolder } from "@/utils";
import { FILE_TYPES, TFileType } from "@/lib/file-types";
import UploadForm from "./upload-form";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SidebarGroupAction } from "@/components/ui/sidebar";
import { uploadAll } from "../hooks/use-upload-file-mutation";
import toast from "react-hot-toast";
import { constant } from "@/lib/constant";

const ALLOW_DUPLICATES_KEY = "upload:allow-duplicates";

type ConditionalUploadModalProps = { folder?: string } & (
  | { placement: "header"; type: TFileType }
  | { placement?: "sidebar"; type?: TFileType }
);

const UploadModal = ({
  placement = "sidebar",
  type,
  folder: currentFolder = "",
}: ConditionalUploadModalProps) => {
  const [open, setOpen] = useState(false);
  const [files, setFiles] = useState<File[]>([]);
  const [folder, setFolder] = useState(currentFolder);

  // Set initial tab based on type when placement is header, otherwise default to documents
  const [tab, setTab] = useState<TFileType>(
    placement === "header" && type ? type : "documents"
  );

  const [progress, setProgress] = useState({ done: 0, total: 0 });
  // Remembered, because someone who turns the check off usually has a whole
  // batch of legitimately identical files to get through.
  const [allowDuplicates, setAllowDuplicates] = useState(
    () => localStorage.getItem(ALLOW_DUPLICATES_KEY) === "true"
  );

  const handleAllowDuplicates = useCallback((allow: boolean) => {
    setAllowDuplicates(allow);
    localStorage.setItem(ALLOW_DUPLICATES_KEY, String(allow));
  }, []);

  const handleReset = useCallback(() => {
    setFiles([]);
    setFolder(currentFolder);
    setProgress({ done: 0, total: 0 });

    // Reset tab to initial value based on placement and type
    const initialTab = placement === "header" && type ? type : "documents";
    setTab(initialTab);
    setOpen(false);
  }, [placement, type, currentFolder]);

  const queryClient = useQueryClient();

  const { mutate: uploadFileMutate, isPending: isUploadPending } = useMutation({
    mutationFn: async () => {
      setProgress({ done: 0, total: files.length });

      return uploadAll(
        files.map(sanitizeFileName),
        tab,
        sanitizeFolder(folder),
        (done, total) => setProgress({ done, total }),
        allowDuplicates
      );
    },
    onSuccess: async (outcomes) => {
      const failed = outcomes.filter((outcome) => !outcome.ok);
      const uploaded = outcomes.length - failed.length;

      if (uploaded > 0) {
        toast.success(
          uploaded === 1
            ? "Successfully uploaded file!"
            : `Successfully uploaded ${uploaded} files!`
        );
      }

      // Report the failures individually rather than losing the whole batch to
      // the first one: with 100 files, which ones failed is the useful part.
      failed.slice(0, 3).forEach((outcome) => {
        toast.error(`${outcome.file.name}: ${outcome.error}`);
      });
      if (failed.length > 3) {
        toast.error(`${failed.length - 3} more files failed to upload.`);
      }

      Promise.all([
        queryClient.invalidateQueries({ queryKey: constant.queryKeys.all }),
        queryClient.invalidateQueries({
          queryKey: [constant.queryKeys.dashboard],
        }),
      ]);

      if (failed.length === 0) {
        handleReset();
      } else {
        setFiles(failed.map((outcome) => outcome.file));
        setProgress({ done: 0, total: 0 });
      }
    },
  });

  const handleUpload = useCallback(() => {
    if (files.length === 0) {
      return;
    }
    uploadFileMutate();
  }, [files, uploadFileMutate]);

  return (
    <Dialog open={open} onOpenChange={setOpen} modal>
      <DialogTrigger asChild>
        {placement === "header" ? (
          <Button onClick={() => {}} variant="default">
            <Plus />
            Add {type ? FILE_TYPES[type].label : "Content"}
          </Button>
        ) : (
          <SidebarGroupAction title="Add Content">
            <Plus /> <span className="sr-only">Add Content</span>
          </SidebarGroupAction>
        )}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Upload Files</DialogTitle>
          <DialogDescription>
            Upload your files here and manage your content easily.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-2">
          <Label htmlFor="upload-folder">Folder</Label>
          <Input
            id="upload-folder"
            value={folder}
            onChange={(e) => setFolder(e.target.value)}
            placeholder="Root — or e.g. logos/dark"
            disabled={isUploadPending}
          />
        </div>

        <div className="flex items-start gap-2">
          <Checkbox
            id="allow-duplicates"
            checked={allowDuplicates}
            onCheckedChange={(checked) => handleAllowDuplicates(checked === true)}
            disabled={isUploadPending}
          />
          <div className="grid gap-1 leading-none">
            <Label htmlFor="allow-duplicates">Allow duplicate files</Label>
            <p className="text-muted-foreground text-sm">
              Uploads files whose contents already exist in this folder. A name
              that is taken gets a numbered suffix.
            </p>
          </div>
        </div>

        <UploadForm
          isLoading={isUploadPending}
          tab={tab}
          onChangeTab={setTab}
          files={files}
          onChangeFiles={setFiles}
          disableTabSwitching={placement === "header"}
        />

        <DialogFooter className="sm:justify-end">
          <DialogClose asChild>
            <Button
              onClick={handleReset}
              type="button"
              variant="secondary"
              disabled={isUploadPending}
            >
              Cancel
            </Button>
          </DialogClose>
          <Button
            disabled={isUploadPending}
            type="button"
            variant="default"
            onClick={handleUpload}
          >
            {isUploadPending ? (
              <>
                <Loader2Icon className="animate-spin" />
                {progress.total > 0
                  ? `Uploading ${progress.done}/${progress.total}`
                  : "Please wait"}
              </>
            ) : (
              "Upload"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default UploadModal;
