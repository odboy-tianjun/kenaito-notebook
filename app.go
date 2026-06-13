package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	stdruntime "runtime"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"kenaito-notebook/models"
)

// 配置后端地址
const defaultServerURL = "http://localhost:8080"

type App struct {
	ctx       context.Context
	serverURL string
	http      *http.Client
	token     string
}

func NewApp() *App {
	return &App{
		serverURL: defaultServerURL,
		http:      &http.Client{Timeout: 30 * time.Second},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("[App] Starting up, server:", a.serverURL)
}

func (a *App) shutdown(ctx context.Context) {
	log.Println("[App] Shutting down")
}

// doRequest 发起 HTTP 请求的通用方法
func (a *App) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, a.serverURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}

	resp, err := a.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server connection failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf(errResp.Error)
		}
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	return data, nil
}

// Register 用户注册
func (a *App) Register(email, password, code string) (string, error) {
	body := map[string]string{
		"email":    email,
		"password": password,
		"code":     code,
	}
	data, err := a.doRequest("POST", "/api/auth/register", body)
	if err != nil {
		return "", err
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	a.token = resp.Token
	return resp.Token, nil
}

// SendVerifyCode 发送验证码
func (a *App) SendVerifyCode(email, purpose string) error {
	body := map[string]string{
		"email":   email,
		"purpose": purpose,
	}
	_, err := a.doRequest("POST", "/api/auth/send-code", body)
	return err
}

// VerifyCode 验证验证码
func (a *App) VerifyCode(email, code, purpose string) error {
	body := map[string]string{
		"email":   email,
		"code":    code,
		"purpose": purpose,
	}
	_, err := a.doRequest("POST", "/api/auth/verify-code", body)
	return err
}

// ResetPassword 重置密码
func (a *App) ResetPassword(email, code, newPassword string) error {
	body := map[string]string{
		"email":       email,
		"code":        code,
		"newPassword": newPassword,
	}
	_, err := a.doRequest("POST", "/api/auth/reset-password", body)
	return err
}

// Login 用户登录
func (a *App) Login(email, password string) (string, error) {
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	data, err := a.doRequest("POST", "/api/auth/login", body)
	if err != nil {
		return "", err
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	a.token = resp.Token
	return resp.Token, nil
}

// Logout 退出登录
func (a *App) Logout() {
	a.token = ""
}

// GetUser 获取当前用户信息
func (a *App) GetUser() (map[string]interface{}, error) {
	data, err := a.doRequest("GET", "/api/auth/me", nil)
	if err != nil {
		return nil, err
	}
	var user map[string]interface{}
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return user, nil
}

// IsLoggedIn 是否已登录
func (a *App) IsLoggedIn() bool {
	return a.token != ""
}

// GetNoteTree 获取笔记树
func (a *App) GetNoteTree() ([]*models.NoteTree, error) {
	data, err := a.doRequest("GET", "/api/notes/tree", nil)
	if err != nil {
		return nil, err
	}
	var result []*models.NoteTree
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetNote 获取单个笔记
func (a *App) GetNote(id int64) (*models.Note, error) {
	data, err := a.doRequest("GET", fmt.Sprintf("/api/notes/%d", id), nil)
	if err != nil {
		return nil, err
	}
	var note models.Note
	if err := json.Unmarshal(data, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

// CreateNote 创建笔记或文件夹
func (a *App) CreateNote(parentID int64, title string, isFolder bool) (*models.Note, error) {
	body := map[string]interface{}{
		"parentId": parentID,
		"title":    title,
		"isFolder": isFolder,
	}
	data, err := a.doRequest("POST", "/api/notes", body)
	if err != nil {
		return nil, err
	}
	var note models.Note
	if err := json.Unmarshal(data, &note); err != nil {
		return nil, err
	}
	return &note, nil
}

// UpdateNote 更新笔记
func (a *App) UpdateNote(id int64, title, content, tag string) error {
	body := map[string]string{
		"title":   title,
		"content": content,
		"tag":     tag,
	}
	_, err := a.doRequest("PUT", fmt.Sprintf("/api/notes/%d", id), body)
	return err
}

// DeleteNote 删除笔记
func (a *App) DeleteNote(id int64) error {
	_, err := a.doRequest("DELETE", fmt.Sprintf("/api/notes/%d", id), nil)
	return err
}

// SearchNotes 搜索笔记
func (a *App) SearchNotes(keyword string) ([]*models.Note, error) {
	data, err := a.doRequest("GET", fmt.Sprintf("/api/notes/search?q=%s", keyword), nil)
	if err != nil {
		return nil, err
	}
	var notes []*models.Note
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

// GetTrashNotes 获取回收站中的笔记
func (a *App) GetTrashNotes() ([]*models.Note, error) {
	data, err := a.doRequest("GET", "/api/trash", nil)
	if err != nil {
		return nil, err
	}
	var notes []*models.Note
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

// RestoreNote 恢复笔记
func (a *App) RestoreNote(id int64) error {
	_, err := a.doRequest("POST", fmt.Sprintf("/api/trash/%d/restore", id), nil)
	return err
}

// PermanentDeleteNote 永久删除笔记
func (a *App) PermanentDeleteNote(id int64) error {
	_, err := a.doRequest("DELETE", fmt.Sprintf("/api/trash/%d", id), nil)
	return err
}

// EmptyTrash 清空回收站
func (a *App) EmptyTrash() error {
	_, err := a.doRequest("DELETE", "/api/trash", nil)
	return err
}

// SaveImage 保存图片
func (a *App) SaveImage(base64Data string) (string, error) {
	body := map[string]string{"data": base64Data}
	data, err := a.doRequest("POST", "/api/files/image", body)
	if err != nil {
		return "", err
	}
	var resp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.URL, nil
}

// SaveFile 保存任意文件
func (a *App) SaveFile(base64Data, filename string) (string, error) {
	body := map[string]string{
		"data":     base64Data,
		"filename": filename,
	}
	data, err := a.doRequest("POST", "/api/files/upload", body)
	if err != nil {
		return "", err
	}
	var resp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.URL, nil
}

// GetFileServerURL 获取文件服务器地址
func (a *App) GetFileServerURL() string {
	data, err := a.doRequest("GET", "/api/files/server-url", nil)
	if err != nil {
		return ""
	}
	var resp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return ""
	}
	return resp.URL
}

// GetAssetsPath 获取资源路径
func (a *App) GetAssetsPath() string {
	data, err := a.doRequest("GET", "/api/files/assets-path", nil)
	if err != nil {
		return ""
	}
	var resp struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return ""
	}
	return resp.Path
}

// SelectDirectory 弹出目录选择对话框
func (a *App) SelectDirectory(title string) (string, error) {
	result, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
	if err != nil {
		return "", err
	}
	if result == "" {
		return "", fmt.Errorf("用户取消选择")
	}
	return result, nil
}

// SelectSaveFile 弹出文件保存对话框
func (a *App) SelectSaveFile(title, defaultFilename string) (string, error) {
	result, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
	})
	if err != nil {
		return "", err
	}
	if result == "" {
		return "", fmt.Errorf("用户取消保存")
	}
	return result, nil
}

// RevealInFinder 在 Finder 中显示文件
func (a *App) RevealInFinder(path string) {
	var cmd *exec.Cmd
	switch stdruntime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	case "windows":
		cmd = exec.Command("explorer", "/select,", path)
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(path))
	}
	cmd.Start()
}

// ExportHTML 导出 HTML
func (a *App) ExportHTML(title, content string) (string, error) {
	body := map[string]string{"title": title, "content": content}
	data, err := a.doRequest("POST", "/api/export/html", body)
	if err != nil {
		return "", err
	}
	var resp struct {
		FilePath string `json:"filePath"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}

	defaultFilename := title + ".html"
	savePath, err := a.SelectSaveFile("导出 HTML 文件", defaultFilename)
	if err != nil {
		return "", err
	}
	if filepath.Ext(savePath) != ".html" {
		savePath += ".html"
	}

	if err := copyFile(resp.FilePath, savePath); err != nil {
		return "", err
	}
	return savePath, nil
}

// ExportMarkdown 导出 Markdown
func (a *App) ExportMarkdown(title, content string) (string, error) {
	body := map[string]string{"title": title, "content": content}
	data, err := a.doRequest("POST", "/api/export/markdown", body)
	if err != nil {
		return "", err
	}
	var resp struct {
		FilePath string `json:"filePath"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}

	defaultFilename := title + ".md"
	savePath, err := a.SelectSaveFile("导出 Markdown 文件", defaultFilename)
	if err != nil {
		return "", err
	}
	if filepath.Ext(savePath) != ".md" {
		savePath += ".md"
	}

	if err := copyFile(resp.FilePath, savePath); err != nil {
		return "", err
	}
	return savePath, nil
}

// ExportFolderMarkdown 导出整个文件夹为 ZIP
func (a *App) ExportFolderMarkdown(folderID int64) (string, error) {
	data, err := a.doRequest("POST", fmt.Sprintf("/api/export/folder-zip/%d", folderID), nil)
	if err != nil {
		return "", err
	}
	var resp struct {
		FilePath string `json:"filePath"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}

	savePath, err := a.SelectSaveFile("导出文件夹为 ZIP", "export.zip")
	if err != nil {
		return "", err
	}
	if filepath.Ext(savePath) != ".zip" {
		savePath += ".zip"
	}

	if err := copyFile(resp.FilePath, savePath); err != nil {
		return "", err
	}
	return savePath, nil
}

// ExportAllMarkdown 导出所有笔记为 ZIP
func (a *App) ExportAllMarkdown() (string, error) {
	data, err := a.doRequest("POST", "/api/export/all-zip", nil)
	if err != nil {
		return "", err
	}
	var resp struct {
		FilePath string `json:"filePath"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}

	savePath, err := a.SelectSaveFile("导出所有笔记为 ZIP", "所有笔记.zip")
	if err != nil {
		return "", err
	}
	if filepath.Ext(savePath) != ".zip" {
		savePath += ".zip"
	}

	if err := copyFile(resp.FilePath, savePath); err != nil {
		return "", err
	}
	return savePath, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
