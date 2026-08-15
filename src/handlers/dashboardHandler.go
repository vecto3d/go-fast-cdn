package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/kevinanielsen/go-fast-cdn/src/database"
	"github.com/kevinanielsen/go-fast-cdn/src/models"
	"github.com/kevinanielsen/go-fast-cdn/src/util"
)

type DashboardHandler struct {
	DocRepo    models.FileRepository
	ImageRepo  models.FileRepository
	AudioRepo  models.FileRepository
	UserRepo   models.UserRepository
	ConfigRepo *database.ConfigRepo
}

func NewDashboardHandler(docRepo, imageRepo, audioRepo models.FileRepository, userRepo models.UserRepository, configRepo *database.ConfigRepo) *DashboardHandler {
	return &DashboardHandler{
		DocRepo:    docRepo,
		ImageRepo:  imageRepo,
		AudioRepo:  audioRepo,
		UserRepo:   userRepo,
		ConfigRepo: configRepo,
	}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	var cdnSize int64
	_ = filepath.Walk(util.ExPath+"/uploads",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			cdnSize += info.Size()
			return nil
		})

	docs := h.DocRepo.GetAll()
	images := h.ImageRepo.GetAll()
	audio := h.AudioRepo.GetAll()

	recentUploads := []gin.H{}
	for kind, files := range map[string][]models.FileRecord{
		"doc":   docs,
		"image": images,
		"audio": audio,
	} {
		sort.Slice(files, func(i, j int) bool {
			return files[i].CreatedAt.After(files[j].CreatedAt)
		})
		for _, f := range files[:min(5, len(files))] {
			recentUploads = append(recentUploads, gin.H{"filename": f.FileName, "type": kind, "uploaded_at": f.CreatedAt})
		}
	}

	users, _ := h.UserRepo.GetAllUsers()
	totalUsers := len(users)
	admins := 0
	verified := 0
	usersWith2FA := 0

	for _, user := range users {
		if user.Role == "admin" {
			admins++
		}
		if user.IsVerified {
			verified++
		}
		if user.Is2FAEnabled != nil && *user.Is2FAEnabled {
			usersWith2FA++
		}
	}

	sort.Slice(users, func(i, j int) bool { return users[i].CreatedAt.After(users[j].CreatedAt) })
	recentRegs := []gin.H{}
	for _, u := range users[:min(5, len(users))] {
		recentRegs = append(recentRegs, gin.H{"email": u.Email, "role": u.Role, "created_at": u.CreatedAt})
	}

	regEnabled, _ := h.ConfigRepo.Get("registration_enabled")
	accessTokenTTL, err := h.ConfigRepo.Get("access_token_ttl");
	if accessTokenTTL == "" || err != nil {
		accessTokenTTL = "15" // default 15 minutes
	}

	c.JSON(http.StatusOK, gin.H{
		"files": gin.H{
			"total_size_bytes": cdnSize,
			"documents_count":  len(docs),
			"images_count":     len(images),
			"audio_count":      len(audio),
			"recent_uploads":   recentUploads,
		},
		"users": gin.H{
			"total":                totalUsers,
			"admins":               admins,
			"verified":             verified,
			"recent_registrations": recentRegs,
		},
		"config": gin.H{
			"registration_enabled": regEnabled == "true",
			"access_token_ttl": accessTokenTTL,
		},
		"security": gin.H{
			"users_with_2fa": usersWith2FA,
		},
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
