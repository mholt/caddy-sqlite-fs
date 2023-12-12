SQLite FS plugin for Caddy
==========================

> [!WARNING]
> I whipped this up quickly as a proof of concept. It is experimental and likely has bugs.

This package implements a virtual file system for Caddy using SQLite.

It expects a path to a SQLite database with at least this table in its schema:

```sql
CREATE TABLE IF NOT EXISTS "files" (
	"name" TEXT PRIMARY KEY, -- full file path compliant with: https://pkg.go.dev/io/fs#ValidPath
	"content" BLOB,          -- file bytes
	"modified" INTEGER,      -- unix timestamp of last modification
	"mode" INTEGER           -- file mode
);
```

It can be used like so in the Caddyfile:

```caddy
file_server /database/* {
	fs sqlite data.sql
}
```

> [!NOTE]
> This is not an official repository of the [Caddy Web Server](https://github.com/caddyserver) organization.
