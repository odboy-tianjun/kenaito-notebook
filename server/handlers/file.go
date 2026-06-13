package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"kenaito-notebook-server/services"
)

type FileHandler struct {
	fileSvc *services.FileService
}

func NewFileHandler(fileSvc *services.FileService) *FileHandler {
	return &FileHandler{fileSvc: fileSvc}
}

func (h *FileHandler) SaveImage(c *gin.Context) {
	var req struct {
		Data string `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := h.fileSvc.SaveImage(req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *FileHandler) SaveFile(c *gin.Context) {
	var req struct {
		Data     string `json:"data" binding:"required"`
		Filename string `json:"filename" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := h.fileSvc.SaveFile(req.Data, req.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *FileHandler) GetFileServerURL(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"url": h.fileSvc.GetBaseURL()})
}

func (h *FileHandler) GetAssetsPath(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"path": h.fileSvc.GetAssetsPath()})
}
