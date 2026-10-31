package pgcompat

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"

	"github.com/lib/pq"
)

var (
	intervalNumRegex   = regexp.MustCompile(`(?i)\bINTERVAL\s+(\d+)\s+([a-zA-Z]+)\b`)
	intervalParamRegex = regexp.MustCompile(`(?i)\bINTERVAL\s+\?\s+([a-zA-Z]+)\b`)
	timestampDiffRegex = regexp.MustCompile(`(?i)\bTIMESTAMPDIFF\s*\(\s*SECOND\s*,\s*([^,]+?)\s*,\s*([^)]+?)\s*\)`)
)

func Rebind(q string) string {
	q = timestampDiffRegex.ReplaceAllString(q, "EXTRACT(EPOCH FROM (($2) - ($1)))::bigint")
	q = intervalParamRegex.ReplaceAllString(q, "(? * INTERVAL '1 $1')")
	q = intervalNumRegex.ReplaceAllString(q, "INTERVAL '$1 $2'")

	var b strings.Builder
	idx := 1
	inQuote := false
	for i := 0; i < len(q); i++ {
		ch := q[i]
		if ch == '\'' {
			inQuote = !inQuote
			b.WriteByte(ch)
		} else if ch == '`' && !inQuote {
			b.WriteByte('"')
		} else if ch == '?' && !inQuote {
			b.WriteString(fmt.Sprintf("$%d", idx))
			idx++
		} else {
			b.WriteByte(ch)
		}
	}
	return b.String()
}

type pgResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (r *pgResult) LastInsertId() (int64, error) {
	return r.lastInsertID, nil
}

func (r *pgResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

type pgCompatDriver struct {
	underlying driver.Driver
}

func (d *pgCompatDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.underlying.Open(name)
	if err != nil {
		return nil, err
	}
	return &pgCompatConn{conn: conn}, nil
}

type pgCompatConn struct {
	conn driver.Conn
}

func (c *pgCompatConn) Prepare(query string) (driver.Stmt, error) {
	return c.conn.Prepare(Rebind(query))
}

func (c *pgCompatConn) Close() error {
	return c.conn.Close()
}

func (c *pgCompatConn) Begin() (driver.Tx, error) {
	return c.conn.Begin()
}

var insertRegexp = regexp.MustCompile(`(?i)^\s*INSERT\s+INTO\s+`)
var returningRegexp = regexp.MustCompile(`(?i)\sRETURNING\s`)

func (c *pgCompatConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if qc, ok := c.conn.(driver.QueryerContext); ok {
		return qc.QueryContext(ctx, Rebind(query), args)
	}
	return nil, driver.ErrSkip
}

func (c *pgCompatConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	reb := Rebind(query)

	if insertRegexp.MatchString(reb) && !returningRegexp.MatchString(reb) {
		qWithReturning := reb + " RETURNING id"
		if qc, ok := c.conn.(driver.QueryerContext); ok {
			rows, err := qc.QueryContext(ctx, qWithReturning, args)
			if err == nil {
				defer rows.Close()
				dest := make([]driver.Value, 1)
				if err := rows.Next(dest); err == nil && len(dest) > 0 {
					var lastID int64
					switch v := dest[0].(type) {
					case int64:
						lastID = v
					case int32:
						lastID = int64(v)
					case int:
						lastID = int64(v)
					}
					return &pgResult{lastInsertID: lastID, rowsAffected: 1}, nil
				}
			}
		}
	}

	if ec, ok := c.conn.(driver.ExecerContext); ok {
		return ec.ExecContext(ctx, reb, args)
	}
	return nil, driver.ErrSkip
}

func init() {
	sql.Register("pg_compat", &pgCompatDriver{underlying: &pq.Driver{}})
}
