// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mysql

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/cloudspannerecosystem/harbourbridge/internal"
	"github.com/cloudspannerecosystem/harbourbridge/profiles"
	"github.com/cloudspannerecosystem/harbourbridge/schema"
	"github.com/cloudspannerecosystem/harbourbridge/sources/common"
	"github.com/cloudspannerecosystem/harbourbridge/spanner/ddl"
)

type mockSpec struct {
	query string
	args  []driver.Value   // Query args.
	cols  []string         // Columns names for returned rows.
	rows  [][]driver.Value // Set of rows returned.
}

func TestProcessSchemaMYSQL(t *testing.T) {
	ms := []mockSpec{

		{
			query: "SELECT (.+) FROM information_schema.tables where table_type = 'BASE TABLE'  and (.+)",
			args:  []driver.Value{"test"},
			cols:  []string{"table_name"},
			rows: [][]driver.Value{
				{"user"},
				{"cart"},
				{"product"},
				{"test"},
				{"test_ref"}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "user"},
			cols:  []string{"column_name", "constraint_type"},
			rows: [][]driver.Value{
				{"user_id", "PRIMARY KEY"},
				{"ref", "FOREIGN KEY"}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "user"},
			cols:  []string{"REFERENCED_TABLE_NAME", "COLUMN_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"},
			rows: [][]driver.Value{
				{"test", "ref", "id", "fk_test"},
			},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.STATISTICS (.+)",
			args:  []driver.Value{"test", "user"},
			cols:  []string{"INDEX_NAME", "COLUMN_NAME", "SEQ_IN_INDEX", "COLLATION", "NON_UNIQUE"},
		}, {
			query: "SELECT (.+) FROM information_schema.COLUMNS (.+)",
			args:  []driver.Value{"test", "user"},
			cols:  []string{"column_name", "data_type", "column_type", "is_nullable", "column_default", "character_maximum_length", "numeric_precision", "numeric_scale", "extra"},
			rows: [][]driver.Value{
				{"user_id", "text", "text", "NO", nil, nil, nil, nil, nil},
				{"name", "text", "text", "NO", nil, nil, nil, nil, nil},
				{"ref", "bigint", "bigint", "NO", nil, nil, nil, nil, nil}},
		},
		{
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "cart"},
			cols:  []string{"column_name", "constraint_type"},
			rows: [][]driver.Value{
				{"productid", "PRIMARY KEY"},
				{"userid", "PRIMARY KEY"}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "cart"},
			cols:  []string{"REFERENCED_TABLE_NAME", "COLUMN_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"},
			rows: [][]driver.Value{
				{"product", "productid", "product_id", "fk_test2"},
				{"user", "userid", "user_id", "fk_test3"}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.STATISTICS (.+)",
			args:  []driver.Value{"test", "cart"},
			cols:  []string{"INDEX_NAME", "COLUMN_NAME", "SEQ_IN_INDEX", "COLLATION", "NON_UNIQUE"},
			rows: [][]driver.Value{
				{"index1", "userid", 1, sql.NullString{Valid: false}, "0"},
				{"index2", "userid", 1, "A", "1"},
				{"index2", "productid", 2, "D", "1"},
				{"index3", "productid", 1, "A", "0"},
				{"index3", "userid", 2, "D", "0"}},
		}, {
			query: "SELECT (.+) FROM information_schema.COLUMNS (.+)",
			args:  []driver.Value{"test", "cart"},
			cols:  []string{"column_name", "data_type", "column_type", "is_nullable", "column_default", "character_maximum_length", "numeric_precision", "numeric_scale", "extra"},
			rows: [][]driver.Value{
				{"productid", "text", "text", "NO", nil, nil, nil, nil, nil},
				{"userid", "text", "text", "NO", nil, nil, nil, nil, nil},
				{"quantity", "bigint", "bigint", "YES", nil, nil, 64, 0, nil}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "product"},
			cols:  []string{"column_name", "constraint_type"},
			rows: [][]driver.Value{
				{"product_id", "PRIMARY KEY"}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "product"},
			cols:  []string{"REFERENCED_TABLE_NAME", "COLUMN_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.STATISTICS (.+)",
			args:  []driver.Value{"test", "product"},
			cols:  []string{"INDEX_NAME", "COLUMN_NAME", "SEQ_IN_INDEX", "COLLATION", "NON_UNIQUE"},
		}, {
			query: "SELECT (.+) FROM information_schema.COLUMNS (.+)",
			args:  []driver.Value{"test", "product"},
			cols:  []string{"column_name", "data_type", "column_type", "is_nullable", "column_default", "character_maximum_length", "numeric_precision", "numeric_scale", "extra"},
			rows: [][]driver.Value{
				{"product_id", "text", "text", "NO", nil, nil, nil, nil, nil},
				{"product_name", "text", "text", "NO", nil, nil, nil, nil, nil}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "test"},
			cols:  []string{"column_name", "constraint_type"},
			rows:  [][]driver.Value{{"id", "PRIMARY KEY"}, {"id", "FOREIGN KEY"}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "test"},
			cols:  []string{"REFERENCED_TABLE_NAME", "COLUMN_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"},
			rows: [][]driver.Value{{"test_ref", "id", "ref_id", "fk_test4"},
				{"test_ref", "txt", "ref_txt", "fk_test4"}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.STATISTICS (.+)",
			args:  []driver.Value{"test", "test"},
			cols:  []string{"INDEX_NAME", "COLUMN_NAME", "SEQ_IN_INDEX", "COLLATION", "NON_UNIQUE"},
		}, {
			query: "SELECT (.+) FROM information_schema.COLUMNS (.+)",
			args:  []driver.Value{"test", "test"},
			cols:  []string{"column_name", "data_type", "column_type", "is_nullable", "column_default", "character_maximum_length", "numeric_precision", "numeric_scale", "extra"},
			rows: [][]driver.Value{
				{"id", "bigint", "bigint", "NO", nil, nil, 64, 0, nil},
				{"s", "set", "set", "YES", nil, nil, nil, nil, nil},
				{"txt", "text", "text", "NO", nil, nil, nil, nil, nil},
				{"b", "boolean", "boolean", "YES", nil, nil, nil, nil, nil},
				{"bs", "bigint", "bigint", "NO", "nextval('test11_bs_seq'::regclass)", nil, 64, 0, nil},
				{"bl", "blob", "blob", "YES", nil, nil, nil, nil, nil},
				{"c", "char", "char(1)", "YES", nil, 1, nil, nil, nil},
				{"c8", "char", "char(8)", "YES", nil, 8, nil, nil, nil},
				{"d", "date", "date", "YES", nil, nil, nil, nil, nil},
				{"dec", "decimal", "decimal(20,5)", "YES", nil, nil, 20, 5, nil},
				{"f8", "double", "double", "YES", nil, nil, 53, nil, nil},
				{"f4", "float", "float", "YES", nil, nil, 24, nil, nil},
				{"i8", "bigint", "bigint", "YES", nil, nil, 64, 0, nil},
				{"i4", "integer", "integer", "YES", nil, nil, 32, 0, "auto_increment"},
				{"i2", "smallint", "smallint", "YES", nil, nil, 16, 0, nil},
				{"si", "integer", "integer", "NO", "nextval('test11_s_seq'::regclass)", nil, 32, 0, nil},
				{"ts", "datetime", "datetime", "YES", nil, nil, nil, nil, nil},
				{"tz", "timestamp", "timestamp", "YES", nil, nil, nil, nil, nil},
				{"vc", "varchar", "varchar", "YES", nil, nil, nil, nil, nil},
				{"vc6", "varchar", "varchar(6)", "YES", nil, 6, nil, nil, nil}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "test_ref"},
			cols:  []string{"column_name", "constraint_type"},
			rows: [][]driver.Value{
				{"ref_id", "PRIMARY KEY"},
				{"ref_txt", "PRIMARY KEY"}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "test_ref"},
			cols:  []string{"REFERENCED_TABLE_NAME", "COLUMN_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.STATISTICS (.+)",
			args:  []driver.Value{"test", "test_ref"},
			cols:  []string{"INDEX_NAME", "COLUMN_NAME", "SEQ_IN_INDEX", "COLLATION", "NON_UNIQUE"},
		}, {
			query: "SELECT (.+) FROM information_schema.COLUMNS (.+)",
			args:  []driver.Value{"test", "test_ref"},
			cols:  []string{"column_name", "data_type", "column_type", "is_nullable", "column_default", "character_maximum_length", "numeric_precision", "numeric_scale", "extra"},
			rows: [][]driver.Value{
				{"ref_id", "bigint", "bigint", "NO", nil, nil, 64, 0, nil},
				{"ref_txt", "text", "text", "NO", nil, nil, nil, nil, nil},
				{"abc", "text", "text", "NO", nil, nil, nil, nil, nil}},
		},
	}
	db := mkMockDB(t, ms)
	conv := internal.MakeConv()
	isi := InfoSchemaImpl{"test", db, profiles.SourceProfile{}, profiles.TargetProfile{}}
	err := common.ProcessSchema(conv, isi, 1)
	assert.Nil(t, err)
	expectedSchema := map[string]ddl.CreateTable{
		"user": ddl.CreateTable{
			Name:     "user",
			ColNames: []string{"user_id", "name", "ref"},
			ColDefs: map[string]ddl.ColumnDef{
				"user_id": ddl.ColumnDef{Name: "user_id", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
				"name":    ddl.ColumnDef{Name: "name", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
				"ref":     ddl.ColumnDef{Name: "ref", T: ddl.Type{Name: ddl.Int64}, NotNull: true},
			},
			Pks: []ddl.IndexKey{ddl.IndexKey{Col: "user_id"}},
			Fks: []ddl.Foreignkey{ddl.Foreignkey{Name: "fk_test", Columns: []string{"ref"}, ReferTable: "test", ReferColumns: []string{"id"}}}},
		"cart": ddl.CreateTable{
			Name:     "cart",
			ColNames: []string{"productid", "userid", "quantity"},
			ColDefs: map[string]ddl.ColumnDef{
				"productid": ddl.ColumnDef{Name: "productid", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
				"userid":    ddl.ColumnDef{Name: "userid", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
				"quantity":  ddl.ColumnDef{Name: "quantity", T: ddl.Type{Name: ddl.Int64}},
			},
			Pks: []ddl.IndexKey{ddl.IndexKey{Col: "productid"}, ddl.IndexKey{Col: "userid"}},
			Fks: []ddl.Foreignkey{ddl.Foreignkey{Name: "fk_test2", Columns: []string{"productid"}, ReferTable: "product", ReferColumns: []string{"product_id"}},
				ddl.Foreignkey{Name: "fk_test3", Columns: []string{"userid"}, ReferTable: "user", ReferColumns: []string{"user_id"}}},
			Indexes: []ddl.CreateIndex{ddl.CreateIndex{Name: "index1", Table: "cart", Unique: true, Keys: []ddl.IndexKey{ddl.IndexKey{Col: "userid", Desc: false}}},
				ddl.CreateIndex{Name: "index2", Table: "cart", Unique: false, Keys: []ddl.IndexKey{ddl.IndexKey{Col: "userid", Desc: false}, ddl.IndexKey{Col: "productid", Desc: true}}},
				ddl.CreateIndex{Name: "index3", Table: "cart", Unique: true, Keys: []ddl.IndexKey{ddl.IndexKey{Col: "productid", Desc: false}, ddl.IndexKey{Col: "userid", Desc: true}}}}},
		"product": ddl.CreateTable{
			Name:     "product",
			ColNames: []string{"product_id", "product_name"},
			ColDefs: map[string]ddl.ColumnDef{
				"product_id":   ddl.ColumnDef{Name: "product_id", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
				"product_name": ddl.ColumnDef{Name: "product_name", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
			},
			Pks: []ddl.IndexKey{ddl.IndexKey{Col: "product_id"}}},
		"test": ddl.CreateTable{
			Name:     "test",
			ColNames: []string{"id", "s", "txt", "b", "bs", "bl", "c", "c8", "d", "dec", "f8", "f4", "i8", "i4", "i2", "si", "ts", "tz", "vc", "vc6"},
			ColDefs: map[string]ddl.ColumnDef{
				"id":  ddl.ColumnDef{Name: "id", T: ddl.Type{Name: ddl.Int64}, NotNull: true},
				"s":   ddl.ColumnDef{Name: "s", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength, IsArray: true}},
				"txt": ddl.ColumnDef{Name: "txt", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
				"b":   ddl.ColumnDef{Name: "b", T: ddl.Type{Name: ddl.Bool}},
				"bs":  ddl.ColumnDef{Name: "bs", T: ddl.Type{Name: ddl.Int64}, NotNull: true},
				"bl":  ddl.ColumnDef{Name: "bl", T: ddl.Type{Name: ddl.Bytes, Len: ddl.MaxLength}},
				"c":   ddl.ColumnDef{Name: "c", T: ddl.Type{Name: ddl.String, Len: int64(1)}},
				"c8":  ddl.ColumnDef{Name: "c8", T: ddl.Type{Name: ddl.String, Len: int64(8)}},
				"d":   ddl.ColumnDef{Name: "d", T: ddl.Type{Name: ddl.Date}},
				"dec": ddl.ColumnDef{Name: "dec", T: ddl.Type{Name: ddl.Numeric}},
				"f8":  ddl.ColumnDef{Name: "f8", T: ddl.Type{Name: ddl.Float64}},
				"f4":  ddl.ColumnDef{Name: "f4", T: ddl.Type{Name: ddl.Float64}},
				"i8":  ddl.ColumnDef{Name: "i8", T: ddl.Type{Name: ddl.Int64}},
				"i4":  ddl.ColumnDef{Name: "i4", T: ddl.Type{Name: ddl.Int64}},
				"i2":  ddl.ColumnDef{Name: "i2", T: ddl.Type{Name: ddl.Int64}},
				"si":  ddl.ColumnDef{Name: "si", T: ddl.Type{Name: ddl.Int64}, NotNull: true},
				"ts":  ddl.ColumnDef{Name: "ts", T: ddl.Type{Name: ddl.Timestamp}},
				"tz":  ddl.ColumnDef{Name: "tz", T: ddl.Type{Name: ddl.Timestamp}},
				"vc":  ddl.ColumnDef{Name: "vc", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}},
				"vc6": ddl.ColumnDef{Name: "vc6", T: ddl.Type{Name: ddl.String, Len: int64(6)}},
			},
			Pks: []ddl.IndexKey{ddl.IndexKey{Col: "id"}},
			Fks: []ddl.Foreignkey{ddl.Foreignkey{Name: "fk_test4", Columns: []string{"id", "txt"}, ReferTable: "test_ref", ReferColumns: []string{"ref_id", "ref_txt"}}}},
		"test_ref": ddl.CreateTable{
			Name:     "test_ref",
			ColNames: []string{"ref_id", "ref_txt", "abc"},
			ColDefs: map[string]ddl.ColumnDef{
				"ref_id":  ddl.ColumnDef{Name: "ref_id", T: ddl.Type{Name: ddl.Int64}, NotNull: true},
				"ref_txt": ddl.ColumnDef{Name: "ref_txt", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
				"abc":     ddl.ColumnDef{Name: "abc", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
			},
			Pks: []ddl.IndexKey{ddl.IndexKey{Col: "ref_id"}, ddl.IndexKey{Col: "ref_txt"}}},
	}
	assert.Equal(t, expectedSchema, stripSchemaComments(conv.SpSchema))
	assert.Equal(t, len(conv.Issues["cart"]), 0)
	expectedIssues := map[string][]internal.SchemaIssue{
		"bs": []internal.SchemaIssue{internal.DefaultValue},
		"f4": []internal.SchemaIssue{internal.Widened},
		"i4": []internal.SchemaIssue{internal.Widened, internal.AutoIncrement},
		"i2": []internal.SchemaIssue{internal.Widened},
		"si": []internal.SchemaIssue{internal.Widened, internal.DefaultValue},
		"ts": []internal.SchemaIssue{internal.Datetime},
	}
	assert.Equal(t, expectedIssues, conv.Issues["test"])
	assert.Equal(t, int64(0), conv.Unexpecteds())
}

func TestProcessData(t *testing.T) {
	ms := []mockSpec{
		{
			query: "SELECT (.+) FROM `test`.`te st`",
			cols:  []string{"a a", " b", " c "},
			rows: [][]driver.Value{
				{42.3, 3, "cat"},
				{6.6, 22, "dog"},
				{6.6, "2006-01-02", "dog"}}, // Test bad row logic.
		},
	}
	db := mkMockDB(t, ms)
	conv := buildConv(
		ddl.CreateTable{
			Name:     "te_st",
			ColNames: []string{"a a", " b", " c "},
			ColDefs: map[string]ddl.ColumnDef{
				"a_a": ddl.ColumnDef{Name: "a_a", T: ddl.Type{Name: ddl.Float64}},
				"Ab":  ddl.ColumnDef{Name: "Ab", T: ddl.Type{Name: ddl.Int64}},
				"Ac_": ddl.ColumnDef{Name: "Ac_", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}},
			}},
		schema.Table{
			Name:     "te st",
			Schema:   "test",
			ColNames: []string{"a_a", "_b", "_c_"},
			ColDefs: map[string]schema.Column{
				"a a": schema.Column{Name: "a a", Type: schema.Type{Name: "float"}},
				" b":  schema.Column{Name: " b", Type: schema.Type{Name: "int"}},
				" c ": schema.Column{Name: " c ", Type: schema.Type{Name: "text"}},
			}})

	conv.SetDataMode()
	var rows []spannerData
	conv.SetDataSink(
		func(table string, cols []string, vals []interface{}) {
			rows = append(rows, spannerData{table: table, cols: cols, vals: vals})
		})
	isi := InfoSchemaImpl{"test", db, profiles.SourceProfile{}, profiles.TargetProfile{}}
	common.ProcessData(conv, isi)
	assert.Equal(t,
		[]spannerData{
			spannerData{table: "te_st", cols: []string{"a_a", "Ab", "Ac_"}, vals: []interface{}{float64(42.3), int64(3), "cat"}},
			spannerData{table: "te_st", cols: []string{"a_a", "Ab", "Ac_"}, vals: []interface{}{float64(6.6), int64(22), "dog"}},
		},
		rows)
	assert.Equal(t, conv.BadRows(), int64(1))
	assert.Equal(t, conv.SampleBadRows(10), []string{"table=te st cols=[a a  b  c ] data=[6.6 2006-01-02 dog]\n"})
	assert.Equal(t, int64(1), conv.Unexpecteds()) // Bad row generates an entry in unexpected.
}

func TestProcessData_MultiCol(t *testing.T) {
	// Tests multi-column behavior of ProcessSQLData (including
	// handling of null columns and synthetic keys). Also tests
	// the combination of ProcessInfoSchema and ProcessSQLData
	// i.e. ProcessSQLData uses the schemas built by
	// ProcessInfoSchema.
	ms := []mockSpec{
		{
			query: "SELECT table_name FROM information_schema.tables where table_type = 'BASE TABLE' and (.+)",
			args:  []driver.Value{"test"},
			cols:  []string{"table_name"},
			rows:  [][]driver.Value{{"test"}},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "test"},
			cols:  []string{"column_name", "constraint_type"},
			rows:  [][]driver.Value{}, // No primary key --> force generation of synthetic key.
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS (.+)",
			args:  []driver.Value{"test", "test"},
			cols:  []string{"REFERENCED_TABLE_NAME", "COLUMN_NAME", "REFERENCED_COLUMN_NAME", "CONSTRAINT_NAME"},
		}, {
			query: "SELECT (.+) FROM INFORMATION_SCHEMA.STATISTICS (.+)",
			args:  []driver.Value{"test", "test"},
			cols:  []string{"INDEX_NAME", "COLUMN_NAME", "SEQ_IN_INDEX", "COLLATION", "NON_UNIQUE"},
		}, {
			query: "SELECT (.+) FROM information_schema.COLUMNS (.+)",
			args:  []driver.Value{"test", "test"},
			cols:  []string{"column_name", "data_type", "column_type", "is_nullable", "column_default", "character_maximum_length", "numeric_precision", "numeric_scale", "extra"},
			rows: [][]driver.Value{
				{"a", "text", "text", "NO", nil, nil, nil, nil, nil},
				{"b", "double", "double", "YES", nil, nil, 53, nil, nil},
				{"c", "bigint", "bigint", "YES", nil, nil, 64, 0, nil}},
		},
		{
			query: "SELECT (.+) FROM `test`.`test`",
			cols:  []string{"a", "b", "c"},
			rows: [][]driver.Value{
				{"cat", 42.3, nil},
				{"dog", nil, 22}},
		},
	}
	db := mkMockDB(t, ms)
	conv := internal.MakeConv()
	isi := InfoSchemaImpl{"test", db, profiles.SourceProfile{}, profiles.TargetProfile{}}
	err := common.ProcessSchema(conv, isi, 1)
	assert.Nil(t, err)
	expectedSchema := map[string]ddl.CreateTable{
		"test": ddl.CreateTable{
			Name:     "test",
			ColNames: []string{"a", "b", "c", "synth_id"},
			ColDefs: map[string]ddl.ColumnDef{
				"a":        ddl.ColumnDef{Name: "a", T: ddl.Type{Name: ddl.String, Len: ddl.MaxLength}, NotNull: true},
				"b":        ddl.ColumnDef{Name: "b", T: ddl.Type{Name: ddl.Float64}},
				"c":        ddl.ColumnDef{Name: "c", T: ddl.Type{Name: ddl.Int64}},
				"synth_id": ddl.ColumnDef{Name: "synth_id", T: ddl.Type{Name: ddl.String, Len: 50}},
			},
			Pks: []ddl.IndexKey{ddl.IndexKey{Col: "synth_id"}}},
	}
	assert.Equal(t, expectedSchema, stripSchemaComments(conv.SpSchema))
	expectedIssues := map[string][]internal.SchemaIssue{}
	assert.Equal(t, expectedIssues, conv.Issues["test"])
	assert.Equal(t, int64(0), conv.Unexpecteds())
	conv.SetDataMode()
	var rows []spannerData
	conv.SetDataSink(
		func(table string, cols []string, vals []interface{}) {
			rows = append(rows, spannerData{table: table, cols: cols, vals: vals})
		})
	common.ProcessData(conv, isi)
	assert.Equal(t, []spannerData{
		{table: "test", cols: []string{"a", "b", "synth_id"}, vals: []interface{}{"cat", float64(42.3), "0"}},
		{table: "test", cols: []string{"a", "c", "synth_id"}, vals: []interface{}{"dog", int64(22), "-9223372036854775808"}}},
		rows)
	assert.Equal(t, int64(0), conv.Unexpecteds())
}

func TestSetRowStats(t *testing.T) {
	ms := []mockSpec{
		{
			query: "SELECT table_name FROM information_schema.tables where table_type = 'BASE TABLE' and (.+)",
			args:  []driver.Value{"test"},
			cols:  []string{"table_name"},
			rows:  [][]driver.Value{{"test1"}, {"test2"}},
		}, {
			query: "SELECT COUNT[(][*][)] FROM `test`.`test1`",
			cols:  []string{"count"},
			rows:  [][]driver.Value{{5}},
		}, {
			query: "SELECT COUNT[(][*][)] FROM `test`.`test2`",
			cols:  []string{"count"},
			rows:  [][]driver.Value{{142}},
		},
	}
	db := mkMockDB(t, ms)
	conv := internal.MakeConv()
	conv.SetDataMode()
	isi := InfoSchemaImpl{"test", db, profiles.SourceProfile{}, profiles.TargetProfile{}}
	common.SetRowStats(conv, isi)
	assert.Equal(t, int64(5), conv.Stats.Rows["test1"])
	assert.Equal(t, int64(142), conv.Stats.Rows["test2"])
	assert.Equal(t, int64(0), conv.Unexpecteds())
}

func mkMockDB(t *testing.T, ms []mockSpec) *sql.DB {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	for _, m := range ms {
		rows := sqlmock.NewRows(m.cols)
		for _, r := range m.rows {
			rows.AddRow(r...)
		}
		if len(m.args) > 0 {
			mock.ExpectQuery(m.query).WithArgs(m.args...).WillReturnRows(rows)
		} else {
			mock.ExpectQuery(m.query).WillReturnRows(rows)
		}
	}
	return db
}
