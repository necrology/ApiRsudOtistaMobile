package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrTableNotFound = errors.New("table not found")

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type ResourceRepository struct {
	db           *sql.DB
	databaseName string
	mu           sync.RWMutex
	catalog      *catalog
}

type ListOptions struct {
	Page          int
	Limit         int
	Search        string
	Columns       []string
	SearchColumns []string
	WithTotal     bool
}

type Pagination struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	HasNext    bool   `json:"has_next"`
	TotalRows  *int64 `json:"total_rows,omitempty"`
	TotalPages *int   `json:"total_pages,omitempty"`
}

type TableSummary struct {
	Name          string   `json:"name"`
	PrimaryKey    string   `json:"primary_key,omitempty"`
	SortColumn    string   `json:"sort_column"`
	ColumnCount   int      `json:"column_count"`
	SearchColumns []string `json:"search_columns"`
}

type TableInfo struct {
	Name          string       `json:"name"`
	PrimaryKey    string       `json:"primary_key,omitempty"`
	SortColumn    string       `json:"sort_column"`
	ColumnCount   int          `json:"column_count"`
	SearchColumns []string     `json:"search_columns"`
	Columns       []ColumnInfo `json:"columns"`
	columnSet     map[string]bool
	searchSet     map[string]bool
}

type ColumnInfo struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	ColumnType string `json:"column_type"`
	IsNullable bool   `json:"is_nullable"`
	Key        string `json:"key,omitempty"`
	Extra      string `json:"extra,omitempty"`
	Searchable bool   `json:"searchable"`
}

type catalog struct {
	tables    map[string]TableInfo
	summaries []TableSummary
}

func NewResourceRepository(db *sql.DB, databaseName string) *ResourceRepository {
	return &ResourceRepository{
		db:           db,
		databaseName: databaseName,
	}
}

func (r *ResourceRepository) Tables(ctx context.Context) ([]TableSummary, error) {
	if err := r.ensureCatalog(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	summaries := make([]TableSummary, len(r.catalog.summaries))
	copy(summaries, r.catalog.summaries)

	return summaries, nil
}

func (r *ResourceRepository) Table(ctx context.Context, tableName string) (TableInfo, error) {
	if err := r.ensureCatalog(ctx); err != nil {
		return TableInfo{}, err
	}

	return r.table(tableName)
}

func (r *ResourceRepository) Search(ctx context.Context, tableName string, opts ListOptions) ([]map[string]any, Pagination, error) {
	if err := r.ensureCatalog(ctx); err != nil {
		return nil, Pagination{}, err
	}

	table, err := r.table(tableName)
	if err != nil {
		return nil, Pagination{}, err
	}

	opts = normalizeListOptions(opts)

	selectedColumns, err := validatedColumns(table, opts.Columns)
	if err != nil {
		return nil, Pagination{}, err
	}

	searchColumns, err := validatedSearchColumns(table, opts.SearchColumns)
	if err != nil {
		return nil, Pagination{}, err
	}

	where, args, err := buildSearchClause(searchColumns, opts.Search)
	if err != nil {
		return nil, Pagination{}, err
	}

	var totalRows *int64
	var totalPages *int
	if opts.WithTotal {
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", quoteIdentifier(table.Name), where)

		var count int64
		if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&count); err != nil {
			return nil, Pagination{}, err
		}

		pages := int((count + int64(opts.Limit) - 1) / int64(opts.Limit))
		totalRows = &count
		totalPages = &pages
	}

	offset := (opts.Page - 1) * opts.Limit
	query := fmt.Sprintf(
		"SELECT %s FROM %s%s ORDER BY %s ASC LIMIT ? OFFSET ?",
		quoteIdentifiers(selectedColumns),
		quoteIdentifier(table.Name),
		where,
		quoteIdentifier(table.SortColumn),
	)

	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, opts.Limit+1, offset)

	rows, err := r.db.QueryContext(ctx, query, listArgs...)
	if err != nil {
		return nil, Pagination{}, err
	}
	defer rows.Close()

	records, err := scanRows(rows)
	if err != nil {
		return nil, Pagination{}, err
	}

	hasNext := len(records) > opts.Limit
	if hasNext {
		records = records[:opts.Limit]
	}

	pagination := Pagination{
		Page:       opts.Page,
		Limit:      opts.Limit,
		HasNext:    hasNext,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}

	return records, pagination, nil
}

func (r *ResourceRepository) FindByID(ctx context.Context, tableName string, id string) (map[string]any, error) {
	if err := r.ensureCatalog(ctx); err != nil {
		return nil, err
	}

	table, err := r.table(tableName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE %s = ? LIMIT 1",
		quoteIdentifier(table.Name),
		quoteIdentifier(table.SortColumn),
	)

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, sql.ErrNoRows
	}

	return records[0], nil
}

func (r *ResourceRepository) ensureCatalog(ctx context.Context) error {
	r.mu.RLock()
	loaded := r.catalog != nil
	r.mu.RUnlock()

	if loaded {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.catalog != nil {
		return nil
	}

	loadedCatalog, err := r.loadCatalog(ctx)
	if err != nil {
		return err
	}

	r.catalog = loadedCatalog
	return nil
}

func (r *ResourceRepository) table(tableName string) (TableInfo, error) {
	tableName = strings.TrimSpace(tableName)

	r.mu.RLock()
	defer r.mu.RUnlock()

	table, ok := r.catalog.tables[tableName]
	if !ok {
		return TableInfo{}, ErrTableNotFound
	}

	return cloneTableInfo(table), nil
}

func (r *ResourceRepository) loadCatalog(ctx context.Context) (*catalog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			c.TABLE_NAME,
			c.COLUMN_NAME,
			c.DATA_TYPE,
			c.COLUMN_TYPE,
			c.IS_NULLABLE,
			c.COLUMN_KEY,
			c.EXTRA
		FROM information_schema.COLUMNS c
		INNER JOIN information_schema.TABLES t
			ON t.TABLE_SCHEMA = c.TABLE_SCHEMA
			AND t.TABLE_NAME = c.TABLE_NAME
		WHERE c.TABLE_SCHEMA = ?
			AND t.TABLE_TYPE = 'BASE TABLE'
		ORDER BY c.TABLE_NAME, c.ORDINAL_POSITION
	`, r.databaseName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make(map[string]TableInfo)
	for rows.Next() {
		var (
			tableName  string
			columnName string
			dataType   string
			columnType string
			nullable   string
			key        sql.NullString
			extra      sql.NullString
		)

		if err := rows.Scan(&tableName, &columnName, &dataType, &columnType, &nullable, &key, &extra); err != nil {
			return nil, err
		}

		info := tables[tableName]
		if info.Name == "" {
			info = TableInfo{
				Name:      tableName,
				columnSet: make(map[string]bool),
				searchSet: make(map[string]bool),
			}
		}

		columnKey := key.String
		columnExtra := extra.String
		searchable := isSearchableDataType(dataType)

		column := ColumnInfo{
			Name:       columnName,
			DataType:   dataType,
			ColumnType: columnType,
			IsNullable: nullable == "YES",
			Key:        columnKey,
			Extra:      columnExtra,
			Searchable: searchable,
		}

		info.Columns = append(info.Columns, column)
		info.columnSet[columnName] = true

		if columnKey == "PRI" && info.PrimaryKey == "" {
			info.PrimaryKey = columnName
			info.SortColumn = columnName
		}

		if searchable {
			info.SearchColumns = append(info.SearchColumns, columnName)
			info.searchSet[columnName] = true
		}

		if info.SortColumn == "" {
			info.SortColumn = columnName
		}

		info.ColumnCount = len(info.Columns)
		tables[tableName] = info
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	summaries := make([]TableSummary, 0, len(tables))
	for _, table := range tables {
		summaries = append(summaries, TableSummary{
			Name:          table.Name,
			PrimaryKey:    table.PrimaryKey,
			SortColumn:    table.SortColumn,
			ColumnCount:   table.ColumnCount,
			SearchColumns: append([]string{}, table.SearchColumns...),
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})

	return &catalog{
		tables:    tables,
		summaries: summaries,
	}, nil
}

func normalizeListOptions(opts ListOptions) ListOptions {
	if opts.Page < 1 {
		opts.Page = 1
	}

	if opts.Limit < 1 {
		opts.Limit = 20
	}

	if opts.Limit > 100 {
		opts.Limit = 100
	}

	opts.Search = strings.TrimSpace(opts.Search)

	return opts
}

func validatedColumns(table TableInfo, requested []string) ([]string, error) {
	if len(requested) == 0 {
		columns := make([]string, 0, len(table.Columns))
		for _, column := range table.Columns {
			columns = append(columns, column.Name)
		}

		return columns, nil
	}

	columns := make([]string, 0, len(requested))
	seen := make(map[string]bool)

	for _, column := range requested {
		column = strings.TrimSpace(column)
		if column == "" {
			continue
		}

		if !table.columnSet[column] {
			return nil, &ValidationError{Message: "kolom tidak valid: " + column}
		}

		if !seen[column] {
			columns = append(columns, column)
			seen[column] = true
		}
	}

	if len(columns) == 0 {
		return nil, &ValidationError{Message: "parameter columns tidak berisi kolom valid"}
	}

	return columns, nil
}

func validatedSearchColumns(table TableInfo, requested []string) ([]string, error) {
	if len(requested) == 0 {
		return append([]string{}, table.SearchColumns...), nil
	}

	columns := make([]string, 0, len(requested))
	seen := make(map[string]bool)

	for _, column := range requested {
		column = strings.TrimSpace(column)
		if column == "" {
			continue
		}

		if !table.columnSet[column] {
			return nil, &ValidationError{Message: "kolom pencarian tidak valid: " + column}
		}

		if !table.searchSet[column] {
			return nil, &ValidationError{Message: "kolom tidak mendukung pencarian teks: " + column}
		}

		if !seen[column] {
			columns = append(columns, column)
			seen[column] = true
		}
	}

	if len(columns) == 0 {
		return nil, &ValidationError{Message: "parameter search_columns tidak berisi kolom valid"}
	}

	return columns, nil
}

func buildSearchClause(columns []string, search string) (string, []any, error) {
	search = strings.TrimSpace(search)
	if search == "" {
		return "", nil, nil
	}

	if len(columns) == 0 {
		return "", nil, &ValidationError{Message: "tabel ini tidak memiliki kolom pencarian teks"}
	}

	conditions := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))

	for _, column := range columns {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", quoteIdentifier(column)))
		args = append(args, "%"+search+"%")
	}

	return " WHERE " + strings.Join(conditions, " OR "), args, nil
}

func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	records := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		valuePointers := make([]any, len(columns))
		for i := range values {
			valuePointers[i] = &values[i]
		}

		if err := rows.Scan(valuePointers...); err != nil {
			return nil, err
		}

		record := make(map[string]any, len(columns))
		for i, column := range columns {
			record[column] = normalizeSQLValue(values[i])
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func normalizeSQLValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case []byte:
		return normalizeDisplayString(string(typed))
	case time.Time:
		return typed.Format(time.RFC3339)
	case string:
		return normalizeDisplayString(typed)
	default:
		return typed
	}
}

func normalizeDisplayString(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "null" {
		return ""
	}

	return timePattern.ReplaceAllString(trimmed, `${1}:${2}`)
}

var timePattern = regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2})`)

func cloneTableInfo(table TableInfo) TableInfo {
	cloned := table
	cloned.Columns = append([]ColumnInfo{}, table.Columns...)
	cloned.SearchColumns = append([]string{}, table.SearchColumns...)
	cloned.columnSet = make(map[string]bool, len(table.columnSet))
	cloned.searchSet = make(map[string]bool, len(table.searchSet))

	for key, value := range table.columnSet {
		cloned.columnSet[key] = value
	}

	for key, value := range table.searchSet {
		cloned.searchSet[key] = value
	}

	return cloned
}

func isSearchableDataType(dataType string) bool {
	switch strings.ToLower(dataType) {
	case "char", "varchar", "tinytext", "text", "mediumtext", "longtext", "enum", "set":
		return true
	default:
		return false
	}
}

func quoteIdentifiers(identifiers []string) string {
	quoted := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		quoted = append(quoted, quoteIdentifier(identifier))
	}

	return strings.Join(quoted, ", ")
}

func quoteIdentifier(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}
