package sqlitefs

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	caddy.RegisterModule(SQLiteFS{})
}

// SQLiteFS implements a virtual file system with a sqlite database.
type SQLiteFS struct {
	DBPath string `json:"db_path,omitempty"`

	db *sql.DB
}

// Validate the SQLite connection with a ping
func (s *SQLiteFS) Validate() error {
	if s.db == nil {
		return errors.New("sqlitefs: database is not opened")
	}
	return s.db.Ping()
}

// CaddyModule returns the Caddy module information.
func (SQLiteFS) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.fs.sqlite",
		New: func() caddy.Module { return new(SQLiteFS) },
	}
}

func (s SQLiteFS) Provision(ctx caddy.Context) error {
	db, err := sql.Open("sqlite3", s.DBPath)
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

func (s SQLiteFS) Cleanup() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Open implements fs.FS.
func (s SQLiteFS) Open(name string) (fs.File, error) {
	row := s.db.QueryRow("SELECT content, modified, mode FROM files WHERE name=? LIMIT 1")

	var content []byte
	var modified *int64
	var mode *int32
	err := row.Scan(&content, &modified, &mode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fs.ErrNotExist
	}
	if err != nil {
		return nil, fmt.Errorf("querying DB or scanning row: %w", err)
	}

	f := sqliteFile{
		reader: bytes.NewBuffer(content),
		info: sqliteFileInfo{
			size: int64(len(content)),
		},
	}
	if modified != nil {
		f.info.modTime = time.Unix(*modified, 0)
	}
	if mode != nil {
		f.info.mode = fs.FileMode(*mode)
	}

	return f, nil
}

type sqliteFile struct {
	reader *bytes.Buffer
	info   sqliteFileInfo
}

func (f sqliteFile) Stat() (fs.FileInfo, error) { return f.info, nil }
func (f sqliteFile) Read(p []byte) (int, error) { return f.reader.Read(p) }
func (f sqliteFile) Close() error {
	f.reader = nil
	f.info = sqliteFileInfo{}
	return nil
}

type sqliteFileInfo struct {
	name    string // full path
	size    int64
	modTime time.Time
	mode    fs.FileMode
}

func (fi sqliteFileInfo) Name() string       { return path.Base(fi.name) }
func (fi sqliteFileInfo) Size() int64        { return fi.size }
func (fi sqliteFileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi sqliteFileInfo) ModTime() time.Time { return fi.modTime }
func (fi sqliteFileInfo) IsDir() bool        { return fi.mode.IsDir() }
func (fi sqliteFileInfo) Sys() any           { return nil }

// Interface guards
var (
	_ caddy.Provisioner     = (*SQLiteFS)(nil)
	_ caddy.CleanerUpper    = (*SQLiteFS)(nil)
	_ fs.FS                 = (*SQLiteFS)(nil)
	_ caddyfile.Unmarshaler = (*SQLiteFS)(nil)
	_ caddy.Validator       = (*SQLiteFS)(nil)
)
