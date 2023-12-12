package sqlitefs

import "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"

// UnmarshalCaddyfile implements caddyfile.Unmarshaler. Syntax:
//
//	sqlite <db_path>
func (s *SQLiteFS) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.AllArgs(&s.DBPath) {
			return d.ArgErr()
		}
	}
	return nil
}
