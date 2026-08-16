import ContentCard from "./content-card";
import useGetFilesQuery from "./hooks/use-get-files-query";
import { Input } from "@/components/ui/input";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  CheckSquare,
  ChevronRight,
  Folder,
  FolderPlus,
  Home,
  List,
  Trash,
  Trash2,
  X,
} from "lucide-react";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import UploadModal from "./upload/upload-modal";
import { FILE_TYPES, TFileType } from "@/lib/file-types";
import { TFile } from "@/types/file";
import { sanitizeFolder, slugify } from "@/utils";
import {
  useCreateFolderMutation,
  useDeleteFolderMutation,
  useFoldersQuery,
} from "./hooks/use-folders";
import { useDragSelect } from "./hooks/use-drag-select";
import { useFolderRoute } from "./hooks/use-folder-route";

type TFilesProps = {
  type: TFileType;
};

const Files: React.FC<TFilesProps> = ({ type }) => {
  // The folder lives in the URL, so it is linkable and the back button walks
  // back up the tree instead of leaving the page.
  const { folder, openFolder } = useFolderRoute(type);

  const files = useGetFilesQuery({ type });
  const folderList = useFoldersQuery(type);
  const createFolder = useCreateFolderMutation(type);
  const deleteFolder = useDeleteFolderMutation(type);

  const [search, setSearch] = useState("");
  const [debounceSearch, setDebounceSearch] = useState("");
  const [newFolder, setNewFolder] = useState("");
  const [isCreatingFolder, setIsCreatingFolder] = useState(false);

  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);
  const [isSelecting, setIsSelecting] = useState(false);
  // Anchor for shift-range selection: the last file clicked without shift.
  const lastClicked = useRef<string | null>(null);
  const selectionAtDragStart = useRef<string[]>([]);

  // Leaving a folder invalidates the selection, which was folder-scoped.
  // Adjusting during render (React's documented pattern for state derived from
  // a changing input) avoids a second pass with the old folder's selection
  // still applied. The shift anchor needs no reset: a name from another folder
  // simply won't be found in the visible list.
  const [renderedFolder, setRenderedFolder] = useState(folder);
  if (renderedFolder !== folder) {
    setRenderedFolder(folder);
    setSelectedFiles([]);
    setIsSelecting(false);
  }

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
    const inFolder: TFile[] = [];
    const names = new Map<string, string>();

    const addIfBelow = (path: string, displayName?: string) => {
      if (path === folder || !path.startsWith(prefix)) return;
      const segment = path.slice(prefix.length).split("/")[0];
      const isDirectChild = path.slice(prefix.length) === segment;
      // Only a direct child's stored name describes this tile; a deeper folder
      // just proves the tile exists.
      if (isDirectChild && displayName) {
        names.set(segment, displayName);
      } else if (!names.has(segment)) {
        names.set(segment, segment);
      }
    };

    for (const entry of folderList.data ?? []) {
      addIfBelow(entry.path, entry.name);
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
      subFolders: [...names.entries()]
        .map(([path, name]) => ({ path, name }))
        .sort((a, b) => a.name.localeCompare(b.name)),
    };
  }, [files.data, folderList.data, folder, search]);

  // Breadcrumb segments, each with the display name of the folder it points at.
  const breadcrumbs = useMemo(() => {
    if (!folder) return [];

    const names = new Map(
      (folderList.data ?? []).map((entry) => [entry.path, entry.name])
    );

    return folder.split("/").map((segment, index, all) => {
      const path = all.slice(0, index + 1).join("/");
      return { path, name: names.get(path) ?? segment };
    });
  }, [folder, folderList.data]);

  const visibleNames = useMemo(
    () => filteredFiles?.map((file) => file.file_name) ?? [],
    [filteredFiles]
  );

  const allSelected =
    visibleNames.length > 0 && selectedFiles.length === visibleNames.length;

  const handleSearchChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      setDebounceSearch(event.target.value);
    },
    []
  );

  const handleOnSelectFile = useCallback(
    (fileName: string, modifiers?: { shiftKey?: boolean }) => {
      // Shift extends from the last plain click through this one, which is how
      // every file manager behaves.
      if (modifiers?.shiftKey && lastClicked.current) {
        const from = visibleNames.indexOf(lastClicked.current);
        const to = visibleNames.indexOf(fileName);
        if (from !== -1 && to !== -1) {
          const range = visibleNames.slice(
            Math.min(from, to),
            Math.max(from, to) + 1
          );
          setSelectedFiles((prev) => [...new Set([...prev, ...range])]);
          return;
        }
      }

      lastClicked.current = fileName;
      setSelectedFiles((prev) =>
        prev.includes(fileName)
          ? prev.filter((name) => name !== fileName)
          : [...prev, fileName]
      );
    },
    [visibleNames]
  );

  const handleToggleAll = useCallback(() => {
    setSelectedFiles(allSelected ? [] : visibleNames);
  }, [allSelected, visibleNames]);

  const handleDragSelect = useCallback((keys: string[]) => {
    // Additive: a drag adds to whatever was already selected when it started.
    setSelectedFiles([...new Set([...selectionAtDragStart.current, ...keys])]);
  }, []);

  const handleDragStart = useCallback(() => {
    selectionAtDragStart.current = selectedFiles;
  }, [selectedFiles]);

  const { containerRef, dragProps, box } = useDragSelect({
    enabled: isSelecting,
    onSelect: handleDragSelect,
    onStart: handleDragStart,
  });

  const handleStopSelecting = useCallback(() => {
    setIsSelecting(false);
    setSelectedFiles([]);
    lastClicked.current = null;
  }, []);

  useEffect(() => {
    const handler = setTimeout(() => {
      setSearch(debounceSearch);
    }, 300);

    return () => {
      clearTimeout(handler);
    };
  }, [debounceSearch]);

  const filesUnder = useCallback(
    (path: string) =>
      (files.data ?? []).filter(
        (file) =>
          (file.folder ?? "") === path ||
          (file.folder ?? "").startsWith(`${path}/`)
      ).length,
    [files.data]
  );

  return (
    <MainContentWrapper title={FILE_TYPES[type].label}>
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <section className="flex items-center gap-2">
            <Input
              className="w-full max-w-md min-w-xs"
              placeholder="Search files by name"
              value={debounceSearch}
              onChange={handleSearchChange}
              aria-label="Search files"
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
                  onClick={handleStopSelecting}
                  variant="outline"
                  disabled={isDeletingFilesLoading}
                >
                  <X />
                  {selectedFiles.length}
                  {selectedFiles.length === 1
                    ? " File Selected"
                    : " Files Selected"}
                </Button>
                <Button
                  onClick={handleToggleAll}
                  variant="outline"
                  disabled={isDeletingFilesLoading || visibleNames.length === 0}
                >
                  <CheckSquare />
                  {allSelected ? "Deselect All" : "Select All"}
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
                        This action cannot be undone.{" "}
                        {selectedFiles.length === 1
                          ? "1 file"
                          : `${selectedFiles.length} files`}{" "}
                        will be permanently deleted.
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
                        const slug = sanitizeFolder(slugify(newFolder));
                        if (!slug) {
                          toast.error("Folder name empty!");
                          return;
                        }
                        createFolder.mutate(
                          {
                            folder: folder ? `${folder}/${slug}` : slug,
                            displayName: newFolder.trim(),
                          },
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
                          Created in {folder ? `${type}/${folder}` : type}.
                        </DialogDescription>
                      </DialogHeader>
                      <div className="py-4 grid gap-2">
                        <Input
                          value={newFolder}
                          onChange={(e) => setNewFolder(e.target.value)}
                          placeholder="e.g. Vehicle Icons"
                          aria-label="Folder name"
                          autoFocus
                        />
                        {/* The URL uses the slug, so show what it will be. */}
                        <p className="text-muted-foreground text-sm">
                          URL:{" "}
                          <code>
                            {FILE_TYPES[type].route}
                            {folder ? `/${folder}` : ""}/
                            {slugify(newFolder) || "…"}
                          </code>
                        </p>
                      </div>
                      <DialogFooter>
                        <Button type="submit" disabled={createFolder.isPending}>
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
        <nav
          aria-label="Folder breadcrumb"
          className="flex items-center gap-1 text-sm flex-wrap"
        >
          <Button
            variant="ghost"
            size="sm"
            className="px-2"
            onClick={() => openFolder("")}
            disabled={!folder}
          >
            <Home />
            {FILE_TYPES[type].label}
          </Button>
          {breadcrumbs.map((crumb, index) => (
            <span key={crumb.path} className="flex items-center gap-1">
              <ChevronRight className="text-muted-foreground" size={14} />
              <Button
                variant="ghost"
                size="sm"
                className="px-2"
                onClick={() => openFolder(crumb.path)}
                disabled={index === breadcrumbs.length - 1}
              >
                {crumb.name}
              </Button>
            </span>
          ))}
        </nav>
        {isSelecting && (
          <p className="text-muted-foreground text-sm">
            Click a card to select, shift-click for a range, or drag a box over
            several.
          </p>
        )}
        <div
          ref={containerRef}
          {...dragProps}
          className="flex flex-wrap gap-4 relative"
        >
          {subFolders.map((subFolder) => {
            const path = folder
              ? `${folder}/${subFolder.path}`
              : subFolder.path;
            const count = filesUnder(path);

            return (
              <div
                key={subFolder.path}
                className="border rounded-lg shadow-lg flex flex-col w-64 max-w-[256px] items-center justify-center gap-2 p-4 relative"
              >
                <button
                  type="button"
                  onClick={() => openFolder(path)}
                  className="flex flex-col items-center gap-2 w-full hover:opacity-80 transition-opacity"
                >
                  <Folder size="64" />
                  <span className="truncate w-full text-center">
                    {subFolder.name}
                  </span>
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
                      aria-label={`Delete folder ${subFolder.name}`}
                    >
                      <Trash2 />
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>
                        Delete "{subFolder.name}"?
                      </AlertDialogTitle>
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
          {box && (
            <div
              className="fixed border-2 border-primary bg-primary/10 pointer-events-none z-50"
              style={{
                left: box.left,
                top: box.top,
                width: box.width,
                height: box.height,
              }}
            />
          )}
        </div>
      </div>
    </MainContentWrapper>
  );
};

export default Files;
