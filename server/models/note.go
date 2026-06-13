package models

import "time"

type Note struct {
	ID        int64      `json:"id"`
	ParentID  int64      `json:"parentId"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	SortOrder int        `json:"sortOrder"`
	IsFolder  bool       `json:"isFolder"`
	Tag       string     `json:"tag"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	Children  []*Note    `json:"children,omitempty"`
}

type NoteTree struct {
	Note
	Children []*NoteTree `json:"children"`
}
