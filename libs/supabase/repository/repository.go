package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/supabase/config"
	"github.com/himdhiman/dashboard-backend/libs/supabase/errors"
)

// IRepository defines the interface for database operations
type IRepository[T any] interface {
	// Basic CRUD operations
	Insert(ctx context.Context, data T) (T, error)
	InsertMany(ctx context.Context, data []T) ([]T, error)
	Select(ctx context.Context, columns ...string) IRepository[T]
	Update(ctx context.Context, data T) (T, error)
	UpdateMany(ctx context.Context, data map[string]interface{}) error
	Delete(ctx context.Context) error
	DeleteMany(ctx context.Context, filter map[string]interface{}) error

	// Query operations
	Where(column, operator, value string) IRepository[T]
	Or(column, operator, value string) IRepository[T]
	In(column string, values []interface{}) IRepository[T]
	NotIn(column string, values []interface{}) IRepository[T]
	Like(column, pattern string) IRepository[T]
	ILike(column, pattern string) IRepository[T]
	Is(column, value string) IRepository[T]
	IsNot(column, value string) IRepository[T]
	Contains(column string, value interface{}) IRepository[T]
	ContainedBy(column string, value interface{}) IRepository[T]
	RangeGt(column string, value interface{}) IRepository[T]
	RangeGte(column string, value interface{}) IRepository[T]
	RangeLt(column string, value interface{}) IRepository[T]
	RangeLte(column string, value interface{}) IRepository[T]
	RangeAdjacent(column string, value interface{}) IRepository[T]
	RangeOverlaps(column string, value interface{}) IRepository[T]
	TextSearch(column, query string) IRepository[T]
	TextSearchType(column, query, textSearchType string) IRepository[T]
	Match(query map[string]interface{}) IRepository[T]
	Not(column, operator, value string) IRepository[T]

	// Ordering and pagination
	Order(column, direction string) IRepository[T]
	Limit(count int) IRepository[T]
	Offset(count int) IRepository[T]
	Range(from, to int) IRepository[T]

	// Aggregation
	Count(ctx context.Context) (int64, error)
	Single(ctx context.Context) (T, error)

	// Execute queries
	Execute(ctx context.Context) ([]T, error)
	ExecuteSingle(ctx context.Context) (T, error)

	// Filter operations
	Filter(filters map[string]interface{}) IRepository[T]
	Eq(column string, value interface{}) IRepository[T]
	Neq(column string, value interface{}) IRepository[T]
	Gt(column string, value interface{}) IRepository[T]
	Gte(column string, value interface{}) IRepository[T]
	Lt(column string, value interface{}) IRepository[T]
	Lte(column string, value interface{}) IRepository[T]
}

// Repository represents a database repository
type Repository[T any] struct {
	IRepository[T]
	config     *config.Config
	httpClient *http.Client
	logger     logger.ILogger
	table      string
	query      *QueryBuilder
}

// QueryBuilder represents a query builder
type QueryBuilder struct {
	selectColumns []string
	whereClauses  []string
	orderBy       []string
	limit         *int
	offset        *int
	rangeFrom     *int
	rangeTo       *int
	filters       map[string]interface{}
}

// NewRepository creates a new repository
func NewRepository[T any](cfg *config.Config, httpClient *http.Client, logger logger.ILogger, table string) IRepository[T] {
	return &Repository[T]{
		config:     cfg,
		httpClient: httpClient,
		logger:     logger,
		table:      table,
		query:      &QueryBuilder{filters: make(map[string]interface{})},
	}
}

// Insert inserts a single record
func (r *Repository[T]) Insert(ctx context.Context, data T) (T, error) {
	url := r.config.URL + "/rest/v1/" + r.table

	body, err := json.Marshal(data)
	if err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to marshal insert data")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to create insert request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", r.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.Key)
	httpReq.Header.Set("Prefer", "return=representation")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to execute insert request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var zero T
		return zero, errors.NewErrorf("insert failed with status: %d", resp.StatusCode)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to decode insert response")
	}

	r.logger.Info("Record inserted successfully", "table", r.table)
	return result, nil
}

// InsertMany inserts multiple records
func (r *Repository[T]) InsertMany(ctx context.Context, data []T) ([]T, error) {
	url := r.config.URL + "/rest/v1/" + r.table

	body, err := json.Marshal(data)
	if err != nil {
		return nil, errors.NewError(err, "failed to marshal insert many data")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.NewError(err, "failed to create insert many request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", r.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.Key)
	httpReq.Header.Set("Prefer", "return=representation")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute insert many request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.NewErrorf("insert many failed with status: %d", resp.StatusCode)
	}

	var result []T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewError(err, "failed to decode insert many response")
	}

	r.logger.Info("Records inserted successfully", "table", r.table, "count", len(result))
	return result, nil
}

// Select specifies columns to select
func (r *Repository[T]) Select(ctx context.Context, columns ...string) IRepository[T] {
	r.query.selectColumns = columns
	return r
}

// Update updates a record
func (r *Repository[T]) Update(ctx context.Context, data T) (T, error) {
	url := r.config.URL + "/rest/v1/" + r.table + r.buildQueryString()

	body, err := json.Marshal(data)
	if err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to marshal update data")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to create update request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", r.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.Key)
	httpReq.Header.Set("Prefer", "return=representation")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to execute update request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var zero T
		return zero, errors.NewErrorf("update failed with status: %d", resp.StatusCode)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to decode update response")
	}

	r.logger.Info("Record updated successfully", "table", r.table)
	return result, nil
}

// UpdateMany updates multiple records
func (r *Repository[T]) UpdateMany(ctx context.Context, data map[string]interface{}) error {
	url := r.config.URL + "/rest/v1/" + r.table + r.buildQueryString()

	body, err := json.Marshal(data)
	if err != nil {
		return errors.NewError(err, "failed to marshal update many data")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.NewError(err, "failed to create update many request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", r.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.Key)

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute update many request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("update many failed with status: %d", resp.StatusCode)
	}

	r.logger.Info("Records updated successfully", "table", r.table)
	return nil
}

// Delete deletes records
func (r *Repository[T]) Delete(ctx context.Context) error {
	url := r.config.URL + "/rest/v1/" + r.table + r.buildQueryString()

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return errors.NewError(err, "failed to create delete request")
	}

	httpReq.Header.Set("apikey", r.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.Key)

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute delete request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("delete failed with status: %d", resp.StatusCode)
	}

	r.logger.Info("Records deleted successfully", "table", r.table)
	return nil
}

// DeleteMany deletes multiple records with filter
func (r *Repository[T]) DeleteMany(ctx context.Context, filter map[string]interface{}) error {
	// Add filter to query
	for k, v := range filter {
		r.query.filters[k] = v
	}

	return r.Delete(ctx)
}

// Where adds a where clause
func (r *Repository[T]) Where(column, operator, value string) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.%s.%s", column, operator, value))
	return r
}

// Or adds an OR clause
func (r *Repository[T]) Or(column, operator, value string) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("or(%s.%s.%s)", column, operator, value))
	return r
}

// In adds an IN clause
func (r *Repository[T]) In(column string, values []interface{}) IRepository[T] {
	valueStr := strings.Join(r.stringifyValues(values), ",")
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.in.(%s)", column, valueStr))
	return r
}

// NotIn adds a NOT IN clause
func (r *Repository[T]) NotIn(column string, values []interface{}) IRepository[T] {
	valueStr := strings.Join(r.stringifyValues(values), ",")
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.not.in.(%s)", column, valueStr))
	return r
}

// Like adds a LIKE clause
func (r *Repository[T]) Like(column, pattern string) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.like.%s", column, pattern))
	return r
}

// ILike adds an ILIKE clause
func (r *Repository[T]) ILike(column, pattern string) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.ilike.%s", column, pattern))
	return r
}

// Is adds an IS clause
func (r *Repository[T]) Is(column, value string) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.is.%s", column, value))
	return r
}

// IsNot adds an IS NOT clause
func (r *Repository[T]) IsNot(column, value string) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.not.is.%s", column, value))
	return r
}

// Contains adds a CONTAINS clause
func (r *Repository[T]) Contains(column string, value interface{}) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.cs.%v", column, value))
	return r
}

// ContainedBy adds a CONTAINED BY clause
func (r *Repository[T]) ContainedBy(column string, value interface{}) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.cd.%v", column, value))
	return r
}

// RangeGt adds a range greater than clause
func (r *Repository[T]) RangeGt(column string, value interface{}) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.sr.gt.%v", column, value))
	return r
}

// RangeGte adds a range greater than or equal clause
func (r *Repository[T]) RangeGte(column string, value interface{}) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.sr.gte.%v", column, value))
	return r
}

// RangeLt adds a range less than clause
func (r *Repository[T]) RangeLt(column string, value interface{}) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.sr.lt.%v", column, value))
	return r
}

// RangeLte adds a range less than or equal clause
func (r *Repository[T]) RangeLte(column string, value interface{}) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.sr.lte.%v", column, value))
	return r
}

// RangeAdjacent adds a range adjacent clause
func (r *Repository[T]) RangeAdjacent(column string, value interface{}) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.sr.adj.%v", column, value))
	return r
}

// RangeOverlaps adds a range overlaps clause
func (r *Repository[T]) RangeOverlaps(column string, value interface{}) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.sr.ov.%v", column, value))
	return r
}

// TextSearch adds a text search clause
func (r *Repository[T]) TextSearch(column, query string) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.textsearch.%s", column, query))
	return r
}

// TextSearchType adds a text search with type clause
func (r *Repository[T]) TextSearchType(column, query, textSearchType string) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("%s.textsearch.%s.%s", column, textSearchType, query))
	return r
}

// Match adds a match clause
func (r *Repository[T]) Match(query map[string]interface{}) IRepository[T] {
	for k, v := range query {
		r.query.filters[k] = v
	}
	return r
}

// Not adds a NOT clause
func (r *Repository[T]) Not(column, operator, value string) IRepository[T] {
	r.query.whereClauses = append(r.query.whereClauses, fmt.Sprintf("not(%s.%s.%s)", column, operator, value))
	return r
}

// Order adds an ORDER BY clause
func (r *Repository[T]) Order(column, direction string) IRepository[T] {
	r.query.orderBy = append(r.query.orderBy, fmt.Sprintf("%s.%s", column, direction))
	return r
}

// Limit adds a LIMIT clause
func (r *Repository[T]) Limit(count int) IRepository[T] {
	r.query.limit = &count
	return r
}

// Offset adds an OFFSET clause
func (r *Repository[T]) Offset(count int) IRepository[T] {
	r.query.offset = &count
	return r
}

// Range adds a RANGE clause
func (r *Repository[T]) Range(from, to int) IRepository[T] {
	r.query.rangeFrom = &from
	r.query.rangeTo = &to
	return r
}

// Count returns the count of records
func (r *Repository[T]) Count(ctx context.Context) (int64, error) {
	url := r.config.URL + "/rest/v1/" + r.table + r.buildQueryString()

	httpReq, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, errors.NewError(err, "failed to create count request")
	}

	httpReq.Header.Set("apikey", r.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.Key)

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return 0, errors.NewError(err, "failed to execute count request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.NewErrorf("count failed with status: %d", resp.StatusCode)
	}

	countStr := resp.Header.Get("Content-Range")
	if countStr == "" {
		return 0, errors.NewError(err, "no content range header found")
	}

	// Parse content range header
	parts := strings.Split(countStr, "/")
	if len(parts) != 2 {
		return 0, errors.NewError(err, "invalid content range header format")
	}

	count, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, errors.NewError(err, "failed to parse count")
	}

	return count, nil
}

// Single returns a single record
func (r *Repository[T]) Single(ctx context.Context) (T, error) {
	url := r.config.URL + "/rest/v1/" + r.table + r.buildQueryString()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to create single request")
	}

	httpReq.Header.Set("apikey", r.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.Key)
	httpReq.Header.Set("Prefer", "count=exact")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to execute single request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var zero T
		return zero, errors.NewErrorf("single failed with status: %d", resp.StatusCode)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		var zero T
		return zero, errors.NewError(err, "failed to decode single response")
	}

	return result, nil
}

// Execute executes the query and returns results
func (r *Repository[T]) Execute(ctx context.Context) ([]T, error) {
	url := r.config.URL + "/rest/v1/" + r.table + r.buildQueryString()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewError(err, "failed to create execute request")
	}

	httpReq.Header.Set("apikey", r.config.Key)
	httpReq.Header.Set("Authorization", "Bearer "+r.config.Key)
	httpReq.Header.Set("Prefer", "count=exact")

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("execute failed with status: %d", resp.StatusCode)
	}

	var result []T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewError(err, "failed to decode execute response")
	}

	return result, nil
}

// ExecuteSingle executes the query and returns a single result
func (r *Repository[T]) ExecuteSingle(ctx context.Context) (T, error) {
	return r.Single(ctx)
}

// Filter adds filters to the query
func (r *Repository[T]) Filter(filters map[string]interface{}) IRepository[T] {
	for k, v := range filters {
		r.query.filters[k] = v
	}
	return r
}

// Eq adds an equals filter
func (r *Repository[T]) Eq(column string, value interface{}) IRepository[T] {
	r.query.filters[column] = value
	return r
}

// Neq adds a not equals filter
func (r *Repository[T]) Neq(column string, value interface{}) IRepository[T] {
	r.query.filters[column+"[neq]"] = value
	return r
}

// Gt adds a greater than filter
func (r *Repository[T]) Gt(column string, value interface{}) IRepository[T] {
	r.query.filters[column+"[gt]"] = value
	return r
}

// Gte adds a greater than or equal filter
func (r *Repository[T]) Gte(column string, value interface{}) IRepository[T] {
	r.query.filters[column+"[gte]"] = value
	return r
}

// Lt adds a less than filter
func (r *Repository[T]) Lt(column string, value interface{}) IRepository[T] {
	r.query.filters[column+"[lt]"] = value
	return r
}

// Lte adds a less than or equal filter
func (r *Repository[T]) Lte(column string, value interface{}) IRepository[T] {
	r.query.filters[column+"[lte]"] = value
	return r
}

// buildQueryString builds the query string from the query builder
func (r *Repository[T]) buildQueryString() string {
	var params []string

	// Add select columns
	if len(r.query.selectColumns) > 0 {
		params = append(params, "select="+strings.Join(r.query.selectColumns, ","))
	}

	// Add where clauses
	if len(r.query.whereClauses) > 0 {
		params = append(params, strings.Join(r.query.whereClauses, "&"))
	}

	// Add filters
	for k, v := range r.query.filters {
		params = append(params, fmt.Sprintf("%s=%v", k, v))
	}

	// Add order by
	if len(r.query.orderBy) > 0 {
		params = append(params, "order="+strings.Join(r.query.orderBy, ","))
	}

	// Add limit
	if r.query.limit != nil {
		params = append(params, fmt.Sprintf("limit=%d", *r.query.limit))
	}

	// Add offset
	if r.query.offset != nil {
		params = append(params, fmt.Sprintf("offset=%d", *r.query.offset))
	}

	// Add range
	if r.query.rangeFrom != nil && r.query.rangeTo != nil {
		params = append(params, fmt.Sprintf("range=%d-%d", *r.query.rangeFrom, *r.query.rangeTo))
	}

	if len(params) == 0 {
		return ""
	}

	return "?" + strings.Join(params, "&")
}

// stringifyValues converts a slice of interface{} to a slice of strings
func (r *Repository[T]) stringifyValues(values []interface{}) []string {
	var result []string
	for _, v := range values {
		result = append(result, fmt.Sprintf("%v", v))
	}
	return result
}
