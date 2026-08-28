package handlers

import (
	"crypto/sha256"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

// HandleSync reconciles the database with what is actually on disk for one file
// type: it adds a row for every stored file that has none, and removes every row
// whose file is gone.
//
// Folders are listed from disk but files from the database, so a file copied
// straight into uploads/ is invisible and a row whose file was deleted by hand
// serves a 404. That makes bulk work painful — re-rendering tens of thousands of
// images means one HTTP upload per file, where a directory copy plus one sync
// does the same job in seconds.
//
//	POST /api/admin/sync/images
//	POST /api/admin/sync/images?dry_run=1     report only, change nothing
//	POST /api/admin/sync/images?prune=0       add missing rows, keep orphans
//	POST /api/admin/sync/images?force=1       allow a prune that empties the table
func (h *FileHandler) HandleSync(c *gin.Context) {
	fileType, repo, ok := h.resolve(c)
	if !ok {
		return
	}

	dryRun := isTrue(c.Query("dry_run"))
	prune := c.Query("prune") == "" || isTrue(c.Query("prune"))
	force := isTrue(c.Query("force"))

	root := filepath.Join(util.ExPath, "uploads", fileType.Name)
	onDisk, err := scanUploads(root, fileType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows := repo.GetAll()

	// An unmounted volume looks exactly like "every file was deleted". Pruning
	// then would wipe the table on a misconfiguration rather than a real change,
	// so it takes an explicit force.
	if prune && len(onDisk) == 0 && len(rows) > 0 && !force {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Refusing to prune: no files found on disk but rows exist",
			"directory": root,
			"rows":      len(rows),
			"hint":      "pass force=1 if the directory really is empty",
		})
		return
	}

	inDB := make(map[string]models.FileRecord, len(rows))
	for _, row := range rows {
		inDB[row.Folder+"/"+row.FileName] = row
	}

	added, removed := []string{}, []string{}
	failed := []gin.H{}

	for key, file := range onDisk {
		if _, exists := inDB[key]; exists {
			continue
		}
		if dryRun {
			added = append(added, key)
			continue
		}

		checksum, err := checksumFile(file.path)
		if err != nil {
			failed = append(failed, gin.H{"file": key, "error": err.Error()})
			continue
		}
		if _, err := repo.Add(models.FileRecord{
			FileName: file.name,
			Folder:   file.folder,
			Checksum: checksum,
		}); err != nil {
			failed = append(failed, gin.H{"file": key, "error": err.Error()})
			continue
		}

		added = append(added, key)
	}

	if prune {
		for key, row := range inDB {
			if _, exists := onDisk[key]; exists {
				continue
			}
			if !dryRun {
				if _, found := repo.Delete(row.Folder, row.FileName); !found {
					failed = append(failed, gin.H{"file": key, "error": "row vanished"})
					continue
				}
			}
			removed = append(removed, key)
		}
	}

	// Only removals need a purge: an added file was not previously served, so no
	// edge holds a response for it. A replaced one keeps its path and is purged
	// by whatever wrote it.
	if !dryRun && len(removed) > 0 {
		urls := make([]string, 0, len(removed))
		for _, row := range inDB {
			if _, exists := onDisk[row.Folder+"/"+row.FileName]; !exists {
				urls = append(urls, purgeURL(c, fileType.Name, row.Folder, row.FileName))
			}
		}
		util.PurgeURLs(urls)
	}

	c.JSON(http.StatusOK, gin.H{
		"type":        fileType.Name,
		"dry_run":     dryRun,
		"pruned":      prune,
		"on_disk":     len(onDisk),
		"rows_before": len(rows),
		"added":       len(added),
		"removed":     len(removed),
		"failed":      failed,
	})
}

type scannedFile struct {
	path   string
	folder string
	name   string
}

// scanUploads walks a type's upload directory and returns every file whose
// extension the type accepts, keyed the same way rows are: "<folder>/<name>",
// with an empty folder for the root.
func scanUploads(root string, fileType models.FileType) (map[string]scannedFile, error) {
	found := make(map[string]scannedFile)

	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return found, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return found, nil
	}

	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		if !fileType.Extensions[strings.ToLower(filepath.Ext(entry.Name()))] {
			return nil
		}

		relative, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		folder := filepath.ToSlash(relative)
		if folder == "." {
			folder = ""
		}

		found[folder+"/"+entry.Name()] = scannedFile{path: path, folder: folder, name: entry.Name()}
		return nil
	})

	return found, err
}

func checksumFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}

func isTrue(value string) bool {
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
