// Package sqlx 提供 SQL 类型的转换函数
//
// 注意: sql.Null* 类型由于字段名不同 (String/Int64/Float64 等)，
// 无法用单一泛型函数处理，因此保留类型特化函数。
package sqlx

import "database/sql"

// NullStringToPtr 将 sql.NullString 转换为 *string
func NullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// PtrToNullString 将 *string 转换为 sql.NullString
func PtrToNullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{}
}

// NullInt64ToPtr 将 sql.NullInt64 转换为 *int64
func NullInt64ToPtr(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

// PtrToNullInt64 将 *int64 转换为 sql.NullInt64
func PtrToNullInt64(i *int64) sql.NullInt64 {
	if i != nil {
		return sql.NullInt64{Int64: *i, Valid: true}
	}
	return sql.NullInt64{}
}

// NullInt32ToPtr 将 sql.NullInt32 转换为 *int32
func NullInt32ToPtr(n sql.NullInt32) *int32 {
	if n.Valid {
		return &n.Int32
	}
	return nil
}

// PtrToNullInt32 将 *int32 转换为 sql.NullInt32
func PtrToNullInt32(i *int32) sql.NullInt32 {
	if i != nil {
		return sql.NullInt32{Int32: *i, Valid: true}
	}
	return sql.NullInt32{}
}

// NullFloat64ToPtr 将 sql.NullFloat64 转换为 *float64
func NullFloat64ToPtr(n sql.NullFloat64) *float64 {
	if n.Valid {
		return &n.Float64
	}
	return nil
}

// PtrToNullFloat64 将 *float64 转换为 sql.NullFloat64
func PtrToNullFloat64(f *float64) sql.NullFloat64 {
	if f != nil {
		return sql.NullFloat64{Float64: *f, Valid: true}
	}
	return sql.NullFloat64{}
}

// NullBoolToPtr 将 sql.NullBool 转换为 *bool
func NullBoolToPtr(n sql.NullBool) *bool {
	if n.Valid {
		return &n.Bool
	}
	return nil
}

// PtrToNullBool 将 *bool 转换为 sql.NullBool
func PtrToNullBool(b *bool) sql.NullBool {
	if b != nil {
		return sql.NullBool{Bool: *b, Valid: true}
	}
	return sql.NullBool{}
}
