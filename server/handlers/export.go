package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"kenaito-notebook-server/models"
	"kenaito-notebook-server/services"
)

type ExportHandler struct {
	db        *services.Database
	exportSvc *services.ExportService
}

func NewExportHandler(db *services.Database, exportSvc *services.ExportService) *ExportHandler {
	return &ExportHandler{db: db, exportSvc: exportSvc}
}

func (h *ExportHandler) ExportHTML(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	outputPath, err := h.getTempExportPath(req.Title, ".html")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.exportSvc.ExportHTML(req.Title, req.Content, outputPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"filePath": outputPath})
}

func (h *ExportHandler) ExportMarkdown(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	outputPath, err := h.getTempExportPath(req.Title, ".md")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.exportSvc.ExportMarkdown(req.Title, req.Content, outputPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"filePath": outputPath})
}

func (h *ExportHandler) ExportFolderZip(c *gin.Context) {
	userID := c.GetInt64("userID")

	folderID, err := strconv.ParseInt(c.Param("folderId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}

	folderTree, err := h.db.GetFolderTree(userID, folderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	noteInfoTree := convertNoteToNoteInfo(folderTree)

	outputPath, err := h.getTempExportPath(folderTree.Title, ".zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.exportSvc.ExportFolderZip(folderTree.Title, noteInfoTree.Children, outputPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"filePath": outputPath})
}

func (h *ExportHandler) ExportAllZip(c *gin.Context) {
	userID := c.GetInt64("userID")

	allNotesTree, err := h.db.GetAllNotesTree(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	noteInfos := make([]*services.NoteInfo, len(allNotesTree))
	for i, noteTree := range allNotesTree {
		noteInfos[i] = convertNoteTreeToNoteInfo(noteTree)
	}

	outputPath, err := h.getTempExportPath("所有笔记", ".zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.exportSvc.ExportAllZip(noteInfos, outputPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"filePath": outputPath})
}

func (h *ExportHandler) getTempExportPath(title, ext string) (string, error) {
	tmpDir := filepath.Join(os.TempDir(), "kenaito-notebook-export")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp export dir: %w", err)
	}

	safeName := h.exportSvc.GetSafeFilename(title)
	return filepath.Join(tmpDir, fmt.Sprintf("%s%s", safeName, ext)), nil
}

func convertNoteToNoteInfo(note *models.Note) *services.NoteInfo {
	if note == nil {
		return nil
	}

	info := &services.NoteInfo{
		Title:    note.Title,
		Content:  note.Content,
		IsFolder: note.IsFolder,
	}

	if note.Children != nil && len(note.Children) > 0 {
		info.Children = make([]*services.NoteInfo, len(note.Children))
		for i, child := range note.Children {
			info.Children[i] = convertNoteToNoteInfo(child)
		}
	}

	return info
}

func convertNoteTreeToNoteInfo(noteTree *models.NoteTree) *services.NoteInfo {
	if noteTree == nil {
		return nil
	}

	info := &services.NoteInfo{
		Title:    noteTree.Title,
		Content:  noteTree.Content,
		IsFolder: noteTree.IsFolder,
	}

	if noteTree.Children != nil && len(noteTree.Children) > 0 {
		info.Children = make([]*services.NoteInfo, len(noteTree.Children))
		for i, child := range noteTree.Children {
			info.Children[i] = convertNoteTreeToNoteInfo(child)
		}
	}

	return info
}
