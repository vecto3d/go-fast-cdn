import ContentCard from "./content-card";
import useGetFilesQuery from "./hooks/use-get-files-query";
import { Input } from "@/components/ui/input";
import { useCallback, useEffect, useMemo, useState } from "react";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  ChevronRight,
  Folder,
  FolderPlus,
  Home,
  List,
  Trash,
  Trash2,
  X,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { TFileType } from "@/lib/file-types";
import { sanitizeFolder } from "@/utils";
import {
  useCreateFolderMutation,
  useDeleteFolderMutation,
  useFoldersQuery,
} from "./hooks/use-folders";
import { cn } from "@/lib/utils";
import MainContentWrapper from "@/components/layouts/main-content-wrapper";
import { Skeleton } from "@/components/ui/skeleton";
import useDeleteFileMutation from "./hooks/use-delete-file-mutation";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { constant } from "@/lib/constant";
import { AxiosError } from "axios";
import { IErrorResponse } from "@/types/response";
import toast from "react-hot-toast";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import UploadModal from "./upload/upload-modal";

type TFilesProps = {
  type: TFileType;
};

const Files: React.FC<TFilesProps> = ({ type }) => {
  const files = useGetFilesQuery({ type });
  const folderList = useFoldersQuery(type);
  const createFolder = useCreateFolderMutation(type);
  const deleteFolder = useDeleteFolderMutation(type);
  const [newFolder, setNewFolder] = useState("");
  const [isCreatingFolder, setIsCreatingFolder] = useState(false);
  const [search, setSearch] = useState("");
  const [debounceSearch, setDebounceSearch] = useState("");
  const [folder, setFolder] = useState("");

  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
  const [isSelecting, setIsSelecting] = useState(false);

  const deleteMutation = useDeleteFileMutation(type);

  const queryClient = useQueryClient();

  const {
    mutateAsync: deleteFilesMutation,
    isPending: isDeletingFilesLoading,
  } = useMutation({
    mutationFn: () =>
      Promise.all(
        selectedFiles.map((fileName) =>
          deleteMutation.mutateAsync({ filename: fileName, folder })
        )
      ),
    onSuccess: () => {
      setSelectedFiles([]);
      setIsSelecting(false);
      queryClient.invalidateQueries({
        queryKey: constant.queryKeys.images(type),
      });
    },
    onError: (error) => {
      const errorResponse = error as AxiosError<IErrorResponse>;
      const message = errorResponse.response?.data?.error || "Delete failed";
      toast.error(message);
    },
  });

  // Subfolders come from the folders that exist on disk (so empty ones show up)
  // union the folders the files themselves are in.
  const { filteredFiles, subFolders } = useMemo(() => {
    const prefix = folder ? `${folder}/` : "";
    const inFolder: typeof files.data = [];
    const names = new Set<string>();

    const addIfBelow = (path: string) => {
      if (path === folder || !path.startsWith(prefix)) return;
      names.add(path.slice(prefix.length).split("/")[0]);
    };

    for (const path of folderList.data ?? []) {
      addIfBelow(path);
    }

    for (const file of files.data ?? []) {
      const fileFolder = file.folder ?? "";
      if (fileFolder === folder) {
        inFolder.push(file);
      } else {
        addIfBelow(fileFolder);
      }
    }

    return {
      filteredFiles: inFolder.filter((file) =>
        file.file_name.toLowerCase().includes(search.toLowerCase())
      ),
      subFolders: [...names].sort(),
    };
  }, [files.data, folderList.data, folder, search]);

  // How much a folder delete would take with it, for the confirm dialog.
  const filesUnder = useCallback(
    (path: string) =>
      (files.data ?? []).filter(
        (file) =>
          (file.folder ?? "") === path || (file.folder ?? "").startsWith(`${path}/`)
      ).length,
    [files.data]
  );

  const breadcrumbs = useMemo(
    () => (folder ? folder.split("/") : []),
    [folder]
  );

  const handleOpenFolder = useCallback((path: string) => {
    setFolder(path);
    setSelectedFiles([]);
    setIsSelecting(false);
  }, []);

  const handleSearchChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      setDebounceSearch(event.target.value);
    },
    []
  );

  const handleOnSelectFile = useCallback(
    (fileName: string) => {
      if (selectedFiles.includes(fileName)) {
        setSelectedFiles((prev) => prev.filter((name) => name !== fileName));
      } else {
        setSelectedFiles((prev) => [...prev, fileName]);
      }
    },
    [selectedFiles]
  );

  useEffect(() => {
    const handler = setTimeout(() => {
      setSearch(debounceSearch);
    }, 300);

    return () => {
      clearTimeout(handler);
    };
  }, [debounceSearch]);

  return (
    <MainContentWrapper title={type}>
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <section className="flex items-center gap-2">
            <Input
              className="w-full max-w-md min-w-xs"
              placeholder="Search files by name"
              value={debounceSearch}
              onChange={handleSearchChange}
              aria-label="Search users"
            />
            <Button
              variant="outline"
              className={cn({
                hidden: !search,
              })}
              onClick={() => {
                setSearch("");
                setDebounceSearch("");
              }}
              aria-label="Clear search"
            >
              <X />
              Clear Search
            </Button>
          </section>
          <section className="flex items-center gap-2">
            {isSelecting ? (
              <>
                <Button
                  onClick={() => {
                    setIsSelecting(false);
                    setSelectedFiles([]);
                  }}
                  variant="outline"
                  disabled={isDeletingFilesLoading}
                >
                  <X />
                  {selectedFiles.length}
                  {selectedFiles.length === 1
                    ? " File Selected"
                    : " Files Selected"}
                </Button>
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="destructive"
                      disabled={
                        selectedFiles.length === 0 || isDeletingFilesLoading
                      }
                    >
                      <Trash />
                      Delete Selected Files
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>
                        Are you sure you want to delete these files?
                      </AlertDialogTitle>
                      <AlertDialogDescription>
                        This action cannot be undone. All selected files will be
                        permanently deleted.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        className={buttonVariants({
                          variant: "destructive",
                        })}
                        onClick={() => deleteFilesMutation()}
                      >
                        Continue
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </>
            ) : (
              <>
                <Button onClick={() => setIsSelecting(true)} variant="outline">
                  <List />
                  Select
                </Button>
                <Dialog
                  open={isCreatingFolder}
                  onOpenChange={(open) => {
                    setIsCreatingFolder(open);
                    if (!open) setNewFolder("");
                  }}
                >
                  <DialogTrigger asChild>
                    <Button variant="outline">
                      <FolderPlus />
                      New Folder
                    </Button>
                  </DialogTrigger>
                  <DialogContent className="sm:max-w-[425px]">
                    <form
                      onSubmit={(e) => {
                        e.preventDefault();
                        const name = sanitizeFolder(newFolder);
                        if (!name) {
                          toast.error("Folder name empty!");
                          return;
                        }
                        createFolder.mutate(
                          folder ? `${folder}/${name}` : name,
                          {
                            onSuccess: () => {
                              setIsCreatingFolder(false);
                              setNewFolder("");
                            },
                          }
                        );
                      }}
                    >
                      <DialogHeader>
                        <DialogTitle>New folder</DialogTitle>
                        <DialogDescription>
                          Created in {folder ? `${type}/${folder}` : type}. Use
                          "/" to nest deeper.
                        </DialogDescription>
                      </DialogHeader>
                      <div className="py-4">
                        <Input
                          value={newFolder}
                          onChange={(e) => setNewFolder(e.target.value)}
                          placeholder="e.g. logos"
                          aria-label="Folder name"
                          autoFocus
                        />
                      </div>
                      <DialogFooter>
                        <Button
                          type="submit"
                          disabled={createFolder.isPending}
                        >
                          {createFolder.isPending ? "Creating..." : "Create"}
                        </Button>
                      </DialogFooter>
                    </form>
                  </DialogContent>
                </Dialog>
                <UploadModal placement="header" type={type} folder={folder} />
              </>
            )}
          </section>
        </div>
        <nav aria-label="Folder breadcrumb" className="flex items-center gap-1 text-sm">
          <Button
            variant="ghost"
            size="sm"
            className="px-2"
            onClick={() => handleOpenFolder("")}
            disabled={!folder}
          >
            <Home />
            {type}
          </Button>
          {breadcrumbs.map((segment, index) => (
            <span key={segment + index} className="flex items-center gap-1">
              <ChevronRight className="text-muted-foreground" size={14} />
              <Button
                variant="ghost"
                size="sm"
                className="px-2"
                onClick={() =>
                  handleOpenFolder(breadcrumbs.slice(0, index + 1).join("/"))
                }
                disabled={index === breadcrumbs.length - 1}
              >
                {segment}
              </Button>
            </span>
          ))}
        </nav>
        <div className="flex flex-wrap gap-4">
          {subFolders.map((name) => {
            const path = folder ? `${folder}/${name}` : name;
            const count = filesUnder(path);

            return (
              <div
                key={name}
                className="border rounded-lg shadow-lg flex flex-col w-64 max-w-[256px] items-center justify-center gap-2 p-4 relative group"
              >
                <button
                  type="button"
                  onClick={() => handleOpenFolder(path)}
                  className="flex flex-col items-center gap-2 w-full hover:opacity-80 transition-opacity"
                >
                  <Folder size="64" />
                  <span className="truncate w-full text-center">{name}</span>
                  <span className="text-muted-foreground text-sm">
                    {count === 1 ? "1 file" : `${count} files`}
                  </span>
                </button>
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="absolute top-2 right-2 text-destructive"
                      aria-label={`Delete folder ${name}`}
                    >
                      <Trash2 />
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Delete "{name}"?</AlertDialogTitle>
                      <AlertDialogDescription>
                        {count === 0
                          ? "This folder is empty. It will be removed."
                          : `This deletes the folder, its subfolders and ${
                              count === 1 ? "1 file" : `${count} files`
                            }. This cannot be undone.`}
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Cancel</AlertDialogCancel>
                      <AlertDialogAction
                        className={buttonVariants({ variant: "destructive" })}
                        onClick={() => deleteFolder.mutate(path)}
                      >
                        Delete
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            );
          })}
          {files.isLoading ? (
            <>
              <Skeleton className="min-h-[264px] w-64 max-w-[256px]" />
              <Skeleton className="min-h-[264px] w-64 max-w-[256px]" />
              <Skeleton className="min-h-[264px] w-64 max-w-[256px]" />
              <Skeleton className="min-h-[264px] w-64 max-w-[256px]" />
              <Skeleton className="min-h-[264px] w-64 max-w-[256px]" />
              <Skeleton className="min-h-[264px] w-64 max-w-[256px]" />
            </>
          ) : (
            <>
              {filteredFiles?.map((file) => (
                <ContentCard
                  type={type}
                  file_name={file.file_name}
                  folder={file.folder}
                  ID={file.ID}
                  createdAt={file.CreatedAt}
                  updatedAt={file.UpdatedAt}
                  key={file.ID}
                  isSelecting={isSelecting}
                  isSelected={selectedFiles.includes(file.file_name)}
                  onSelect={handleOnSelectFile}
                />
              ))}
            </>
          )}
        </div>
      </div>
    </MainContentWrapper>
  );
};

export default Files;
