package services

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"kenaito-notebook-server/models"
)

type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dbDir := filepath.Join(homeDir, ".kenaito-notebook")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "notes.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Database{db: db}, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL DEFAULT 0,
		parent_id INTEGER DEFAULT 0,
		title TEXT NOT NULL DEFAULT '',
		content TEXT DEFAULT '',
		sort_order INTEGER DEFAULT 0,
		is_folder BOOLEAN DEFAULT FALSE,
		tag TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME DEFAULT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_notes_parent_id ON notes(parent_id);
	CREATE INDEX IF NOT EXISTS idx_notes_sort_order ON notes(sort_order);
	CREATE INDEX IF NOT EXISTS idx_notes_deleted_at ON notes(deleted_at);
	CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		hashed_password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	// 迁移：给已有表加 deleted_at 列
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('notes') WHERE name='deleted_at'").Scan(&count)
	if err == nil && count == 0 {
		_, err = db.Exec("ALTER TABLE notes ADD COLUMN deleted_at DATETIME DEFAULT NULL")
		if err != nil {
			return err
		}
	}

	// 迁移：给已有表加 user_id 列
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('notes') WHERE name='user_id'").Scan(&count)
	if err == nil && count == 0 {
		_, err = db.Exec("ALTER TABLE notes ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}

	_, err = db.Exec(query)
	return err
}

func (d *Database) Close() error {
	return d.db.Close()
}

// --- 笔记相关方法 ---

func (d *Database) CreateNote(userID int64, parentID int64, title string, isFolder bool) (*models.Note, error) {
	now := time.Now()
	result, err := d.db.Exec(`
		INSERT INTO notes (user_id, parent_id, title, is_folder, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, parentID, title, isFolder, now, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.Note{
		ID:        id,
		ParentID:  parentID,
		Title:     title,
		IsFolder:  isFolder,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (d *Database) GetNote(id int64) (*models.Note, error) {
	note := &models.Note{}
	err := d.db.QueryRow(`
		SELECT id, parent_id, title, content, sort_order, is_folder, tag, created_at, updated_at, deleted_at
		FROM notes WHERE id = ?
	`, id).Scan(
		&note.ID, &note.ParentID, &note.Title, &note.Content,
		&note.SortOrder, &note.IsFolder, &note.Tag,
		&note.CreatedAt, &note.UpdatedAt, &note.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return note, nil
}

func (d *Database) GetNotesByParentID(userID, parentID int64) ([]*models.Note, error) {
	rows, err := d.db.Query(`
		SELECT id, parent_id, title, content, sort_order, is_folder, tag, created_at, updated_at, deleted_at
		FROM notes WHERE user_id = ? AND parent_id = ? AND deleted_at IS NULL
		ORDER BY sort_order, created_at
	`, userID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note := &models.Note{}
		err := rows.Scan(
			&note.ID, &note.ParentID, &note.Title, &note.Content,
			&note.SortOrder, &note.IsFolder, &note.Tag,
			&note.CreatedAt, &note.UpdatedAt, &note.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (d *Database) GetAllNotes(userID int64) ([]*models.Note, error) {
	rows, err := d.db.Query(`
		SELECT id, parent_id, title, content, sort_order, is_folder, tag, created_at, updated_at, deleted_at
		FROM notes WHERE user_id = ? AND deleted_at IS NULL AND is_folder = FALSE
		ORDER BY sort_order, created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note := &models.Note{}
		err := rows.Scan(
			&note.ID, &note.ParentID, &note.Title, &note.Content,
			&note.SortOrder, &note.IsFolder, &note.Tag,
			&note.CreatedAt, &note.UpdatedAt, &note.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (d *Database) GetAllNotesWithFolders(userID int64) ([]*models.Note, error) {
	rows, err := d.db.Query(`
		SELECT id, parent_id, title, content, sort_order, is_folder, tag, created_at, updated_at, deleted_at
		FROM notes WHERE user_id = ? AND deleted_at IS NULL
		ORDER BY sort_order, created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note := &models.Note{}
		err := rows.Scan(
			&note.ID, &note.ParentID, &note.Title, &note.Content,
			&note.SortOrder, &note.IsFolder, &note.Tag,
			&note.CreatedAt, &note.UpdatedAt, &note.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (d *Database) GetFolderTree(userID, folderID int64) (*models.Note, error) {
	folder, err := d.GetNote(folderID)
	if err != nil {
		return nil, err
	}

	children, err := d.GetNotesByParentID(userID, folderID)
	if err != nil {
		return nil, err
	}

	folder.Children = make([]*models.Note, len(children))
	for i, child := range children {
		if child.IsFolder {
			childFolder, err := d.GetFolderTree(userID, child.ID)
			if err == nil {
				folder.Children[i] = childFolder
			} else {
				folder.Children[i] = child
			}
		} else {
			folder.Children[i] = child
		}
	}

	return folder, nil
}

func (d *Database) GetAllNotesTree(userID int64) ([]*models.NoteTree, error) {
	notes, err := d.GetAllNotesWithFolders(userID)
	if err != nil {
		return nil, err
	}

	return buildTree(notes, 0), nil
}

func (d *Database) UpdateNote(id int64, title, content, tag string) error {
	now := time.Now()
	_, err := d.db.Exec(`
		UPDATE notes SET title = ?, content = ?, tag = ?, updated_at = ?
		WHERE id = ?
	`, title, content, tag, now, id)
	return err
}

func (d *Database) DeleteNote(id int64) error {
	now := time.Now()

	err := d.softDeleteChildren(id, now)
	if err != nil {
		return err
	}

	_, err = d.db.Exec("UPDATE notes SET deleted_at = ? WHERE id = ?", now, id)
	return err
}

func (d *Database) softDeleteChildren(parentID int64, deletedAt time.Time) error {
	rows, err := d.db.Query("SELECT id FROM notes WHERE parent_id = ? AND deleted_at IS NULL", parentID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := d.softDeleteChildren(id, deletedAt); err != nil {
			return err
		}
	}

	_, err = d.db.Exec("UPDATE notes SET deleted_at = ? WHERE parent_id = ? AND deleted_at IS NULL", deletedAt, parentID)
	return err
}

func (d *Database) GetNoteTree(userID int64) ([]*models.NoteTree, error) {
	notes, err := d.getAllNotes(userID)
	if err != nil {
		return nil, err
	}

	return buildTree(notes, 0), nil
}

func (d *Database) getAllNotes(userID int64) ([]*models.Note, error) {
	rows, err := d.db.Query(`
		SELECT id, parent_id, title, content, sort_order, is_folder, tag, created_at, updated_at, deleted_at
		FROM notes
		WHERE user_id = ? AND deleted_at IS NULL
		ORDER BY sort_order, created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note := &models.Note{}
		err := rows.Scan(
			&note.ID, &note.ParentID, &note.Title, &note.Content,
			&note.SortOrder, &note.IsFolder, &note.Tag,
			&note.CreatedAt, &note.UpdatedAt, &note.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func buildTree(notes []*models.Note, parentID int64) []*models.NoteTree {
	var result []*models.NoteTree
	for _, note := range notes {
		if note.ParentID == parentID {
			node := &models.NoteTree{
				Note: *note,
			}
			node.Children = buildTree(notes, note.ID)
			result = append(result, node)
		}
	}
	return result
}

func (d *Database) SearchNotes(userID int64, keyword string) ([]*models.Note, error) {
	rows, err := d.db.Query(`
		SELECT id, parent_id, title, content, sort_order, is_folder, tag, created_at, updated_at, deleted_at
		FROM notes
		WHERE user_id = ? AND (title LIKE ? OR content LIKE ?) AND deleted_at IS NULL
		ORDER BY updated_at DESC
	`, userID, "%"+keyword+"%", "%"+keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note := &models.Note{}
		err := rows.Scan(
			&note.ID, &note.ParentID, &note.Title, &note.Content,
			&note.SortOrder, &note.IsFolder, &note.Tag,
			&note.CreatedAt, &note.UpdatedAt, &note.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (d *Database) UpdateSortOrder(id int64, sortOrder int) error {
	_, err := d.db.Exec("UPDATE notes SET sort_order = ? WHERE id = ?", sortOrder, id)
	return err
}

func (d *Database) GetTrashNotes(userID int64) ([]*models.Note, error) {
	rows, err := d.db.Query(`
		SELECT id, parent_id, title, content, sort_order, is_folder, tag, created_at, updated_at, deleted_at
		FROM notes
		WHERE user_id = ? AND deleted_at IS NOT NULL
		ORDER BY deleted_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*models.Note
	for rows.Next() {
		note := &models.Note{}
		err := rows.Scan(
			&note.ID, &note.ParentID, &note.Title, &note.Content,
			&note.SortOrder, &note.IsFolder, &note.Tag,
			&note.CreatedAt, &note.UpdatedAt, &note.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (d *Database) RestoreNote(id int64) error {
	note, err := d.GetNote(id)
	if err != nil {
		return err
	}

	if note.IsFolder {
		err = d.restoreChildrenRecursive(id)
		if err != nil {
			return err
		}
		_, err = d.db.Exec("UPDATE notes SET deleted_at = NULL WHERE id = ?", id)
	} else {
		_, err = d.db.Exec("UPDATE notes SET deleted_at = NULL, parent_id = 0 WHERE id = ?", id)
	}
	return err
}

func (d *Database) restoreChildrenRecursive(parentID int64) error {
	var parentDeletedAt *time.Time
	err := d.db.QueryRow("SELECT deleted_at FROM notes WHERE id = ?", parentID).Scan(&parentDeletedAt)
	if err != nil || parentDeletedAt == nil {
		return nil
	}

	rows, err := d.db.Query(
		"SELECT id, is_folder FROM notes WHERE parent_id = ? AND deleted_at = ?",
		parentID, *parentDeletedAt,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type childInfo struct {
		id       int64
		isFolder bool
	}
	var children []childInfo

	for rows.Next() {
		var child childInfo
		if err := rows.Scan(&child.id, &child.isFolder); err != nil {
			return err
		}
		children = append(children, child)
	}

	for _, child := range children {
		if child.isFolder {
			if err := d.restoreChildrenRecursive(child.id); err != nil {
				return err
			}
		}
	}

	if len(children) > 0 {
		_, err = d.db.Exec("UPDATE notes SET deleted_at = NULL WHERE parent_id = ? AND deleted_at = ?", parentID, *parentDeletedAt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) PermanentDeleteNote(id int64) error {
	err := d.permanentDeleteChildren(id)
	if err != nil {
		return err
	}

	_, err = d.db.Exec("DELETE FROM notes WHERE id = ?", id)
	return err
}

func (d *Database) permanentDeleteChildren(parentID int64) error {
	rows, err := d.db.Query("SELECT id FROM notes WHERE parent_id = ?", parentID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := d.permanentDeleteChildren(id); err != nil {
			return err
		}
	}

	_, err = d.db.Exec("DELETE FROM notes WHERE parent_id = ?", parentID)
	return err
}

func (d *Database) EmptyTrash(userID int64) error {
	_, err := d.db.Exec("DELETE FROM notes WHERE user_id = ? AND deleted_at IS NOT NULL", userID)
	return err
}

// --- 用户相关方法 ---

type UserRecord struct {
	ID             int64     `json:"id"`
	HashedPassword string    `json:"-"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (d *Database) CreateUser(email, hashedPassword string) (*UserRecord, error) {
	now := time.Now()
	result, err := d.db.Exec(`
		INSERT INTO users (email, hashed_password, created_at)
		VALUES (?, ?, ?)
	`, email, hashedPassword, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &UserRecord{
		ID:             id,
		HashedPassword: hashedPassword,
		Email:          email,
		CreatedAt:      now,
	}, nil
}

func (d *Database) GetUserByEmail(email string) (*UserRecord, error) {
	user := &UserRecord{}
	err := d.db.QueryRow(`
		SELECT id, hashed_password, email, created_at
		FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.HashedPassword, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *Database) GetUserByID(id int64) (*UserRecord, error) {
	user := &UserRecord{}
	err := d.db.QueryRow(`
		SELECT id, hashed_password, email, created_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.HashedPassword, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (d *Database) UpdatePassword(id int64, hashedPassword string) error {
	_, err := d.db.Exec("UPDATE users SET hashed_password = ? WHERE id = ?", hashedPassword, id)
	return err
}
