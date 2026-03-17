package sqlquery

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Executor struct {
	db     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc
}

func (e *Executor) Ping() error {
	err := e.db.PingContext(e.ctx)
	if err != nil {
		lowerErr := strings.ToLower(err.Error())
		switch {

		case strings.Contains(lowerErr, "refused") || strings.Contains(lowerErr, "connect") || strings.Contains(lowerErr, "timeout"):
			// No db on host
			return ConnectionFailure
		case strings.Contains(lowerErr, "database") && strings.Contains(lowerErr, "does not exist"):
			// DB doesen't exist
			return DatabaseFailure
		case strings.Contains(lowerErr, "authentication"):
			// Invalid credentials
			return AuthFailure
		default:
			return err
		}
	}
	return nil
}

// TODO: refactoring
func (e *Executor) PrintableRows(rows *sql.Rows) string {
	if rows == nil {
		return "<no rows>"
	}

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Sprintf("Error getting columns: %v", err)
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	var tableRows [][]string
	tableRows = append(tableRows, columns)

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Sprintf("Error scanning row: %v", err)
		}

		rowStrings := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				rowStrings[i] = "NULL"
			} else {
				rowStrings[i] = fmt.Sprintf("%v", val)
			}
		}

		tableRows = append(tableRows, rowStrings)
	}

	colWidths := make([]int, len(columns))
	for _, row := range tableRows {
		for i, col := range row {
			if len(col) > colWidths[i] {
				colWidths[i] = len(col)
			}
		}
	}

	var sb strings.Builder
	for ri, row := range tableRows {
		for i, col := range row {
			sb.WriteString(col)
			spaces := colWidths[i] - len(col) + 2
			sb.WriteString(strings.Repeat(" ", spaces))
		}
		sb.WriteString("\n")

		if ri == 0 {
			for _, w := range colWidths {
				sb.WriteString(strings.Repeat("-", w+2))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// TODO: execute multiple queries with `;`
func (e *Executor) ExecuteQuery(query string) (*sql.Rows, error) {
	query = strings.TrimSpace(query)
	if len(query) >= 6 && query[:6] == "SELECT" {
		rows, err := e.db.Query(query)
		if err != nil {
			return nil, err
		}
		return rows, nil
	} else {
		_, err := e.db.Exec(query)
		return nil, err
	}
}

// TODO: execute query from file

func (e *Executor) Close() {
	e.cancel()
}

func (e *Executor) Escape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `''`)
	return s
}

func NewExecutor(db *sql.DB) *Executor {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	return &Executor{
		db:     db,
		ctx:    ctx,
		cancel: cancel,
	}
}
