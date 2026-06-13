package services

import (
	"archive/zip"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
)

type ExportService struct{}

func NewExportService() *ExportService {
	return &ExportService{}
}

func (es *ExportService) ExportHTML(title, content, outputPath string) error {
	var htmlContent strings.Builder
	if err := goldmark.Convert([]byte(content), &htmlContent); err != nil {
		return err
	}

	htmlTemplate := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            max-width: 900px;
            margin: 0 auto;
            padding: 20px;
            color: #333;
        }
        h1, h2, h3, h4, h5, h6 {
            margin-top: 1.5em;
            margin-bottom: 0.5em;
        }
        h1 { border-bottom: 2px solid #eee; padding-bottom: 10px; }
        code {
            background-color: #f4f4f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: "Consolas", "Monaco", monospace;
        }
        pre {
            background-color: #f8f8f8;
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 12px;
            overflow-x: auto;
        }
        pre code {
            background: none;
            padding: 0;
        }
        blockquote {
            border-left: 4px solid #ddd;
            margin: 0;
            padding: 0 20px;
            color: #666;
        }
        table {
            border-collapse: collapse;
            width: 100%;
            margin: 1em 0;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px 12px;
            text-align: left;
        }
        th {
            background-color: #f5f5f5;
        }
        img {
            max-width: 100%;
            height: auto;
        }
        a {
            color: #0066cc;
        }
    </style>
</head>
<body>
    <h1>{{.Title}}</h1>
    {{.Content}}
    <hr>
    <p style="color: #999; font-size: 12px;">
        导出于 {{.ExportTime}} | Kenaito Notebook
    </p>
</body>
</html>`

	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, map[string]string{
		"Title":      title,
		"Content":    htmlContent.String(),
		"ExportTime": fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")),
	})
}

func (es *ExportService) ExportMarkdown(title, content, outputPath string) error {
	fullContent := fmt.Sprintf("# %s\n\n%s", title, content)
	return os.WriteFile(outputPath, []byte(fullContent), 0644)
}

func (es *ExportService) ExportFolderMarkdown(folderTitle string, notes []NoteInfo, outputPath string) error {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s\n\n", folderTitle))

	for _, note := range notes {
		content.WriteString(fmt.Sprintf("## %s\n\n", note.Title))
		content.WriteString(note.Content)
		content.WriteString("\n\n---\n\n")
	}

	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}

func (es *ExportService) ExportAllMarkdown(notes []NoteInfo, outputPath string) error {
	var content strings.Builder
	content.WriteString("# 所有笔记\n\n")

	for _, note := range notes {
		content.WriteString(fmt.Sprintf("## %s\n\n", note.Title))
		content.WriteString(note.Content)
		content.WriteString("\n\n---\n\n")
	}

	return os.WriteFile(outputPath, []byte(content.String()), 0644)
}

func (es *ExportService) ExportFolderZip(folderTitle string, notes []*NoteInfo, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	rootFolder := es.GetSafeFilename(folderTitle) + "/"
	_, err = w.Create(rootFolder)
	if err != nil {
		return err
	}

	es.addNotesToZip(w, notes, rootFolder)

	return nil
}

func (es *ExportService) ExportAllZip(notes []*NoteInfo, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	rootFolder := "所有笔记/"
	_, err = w.Create(rootFolder)
	if err != nil {
		return err
	}

	es.addNotesToZip(w, notes, rootFolder)

	return nil
}

func (es *ExportService) addNotesToZip(w *zip.Writer, notes []*NoteInfo, parentPath string) {
	for _, note := range notes {
		safeName := es.GetSafeFilename(note.Title)

		if note.IsFolder {
			folderPath := parentPath + safeName + "/"
			_, _ = w.Create(folderPath)

			if note.Children != nil {
				es.addNotesToZip(w, note.Children, folderPath)
			}
		} else {
			filePath := parentPath + safeName + ".md"
			fullContent := fmt.Sprintf("# %s\n\n%s", note.Title, note.Content)

			f, err := w.Create(filePath)
			if err != nil {
				continue
			}
			_, _ = f.Write([]byte(fullContent))
		}
	}
}

type NoteInfo struct {
	Title    string
	Content  string
	IsFolder bool
	Children []*NoteInfo
}

func (es *ExportService) GetDefaultExportPath(title, ext string) string {
	homeDir, _ := os.UserHomeDir()
	exportDir := filepath.Join(homeDir, "Desktop")
	safeName := es.GetSafeFilename(title)
	return filepath.Join(exportDir, fmt.Sprintf("%s%s", safeName, ext))
}

func (es *ExportService) GetSafeFilename(title string) string {
	safeName := strings.ReplaceAll(title, "/", "-")
	safeName = strings.ReplaceAll(safeName, "\\", "-")
	safeName = strings.ReplaceAll(safeName, ":", "-")
	safeName = strings.ReplaceAll(safeName, "*", "")
	safeName = strings.ReplaceAll(safeName, "?", "")
	safeName = strings.ReplaceAll(safeName, "\"", "")
	safeName = strings.ReplaceAll(safeName, "<", "")
	safeName = strings.ReplaceAll(safeName, ">", "")
	safeName = strings.ReplaceAll(safeName, "|", "")

	if strings.TrimSpace(safeName) == "" {
		safeName = "未命名笔记"
	}

	return safeName
}
