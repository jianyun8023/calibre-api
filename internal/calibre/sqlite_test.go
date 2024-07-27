package calibre

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestDb_Close(t *testing.T) {
	type fields struct {
		dbPath string
		db     *sql.DB
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Db{
				dbPath: tt.fields.dbPath,
				db:     tt.fields.db,
			}
			if err := d.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDb_queryBooks(t *testing.T) {
	type fields struct {
		dbPath string
		db     *sql.DB
	}
	tests := []struct {
		name      string
		fields    fields
		wantBooks []Book
		wantErr   bool
	}{
		// TODO:
		{
			name: "test queryBooks",
			fields: fields{
				dbPath: "/Users/zhaojianyun/Downloads/metadata.db",
				db:     nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _ := NewDb(tt.fields.dbPath)
			gotBooks, err := d.queryBooks()

			// 打印gotBooks
			for _, book := range gotBooks {
				t.Log(book)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("queryBooks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotBooks, tt.wantBooks) {
				t.Errorf("queryBooks() gotBooks = %v, want %v", gotBooks, tt.wantBooks)
			}
		})
	}
}

func TestNewCalibreDb(t *testing.T) {
	type args struct {
		dbPath string
	}
	tests := []struct {
		name    string
		args    args
		want    *Db
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDb(tt.args.dbPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCalibreDb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCalibreDb() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDb(t *testing.T) {
	type args struct {
		sqlite_path string
	}
	tests := []struct {
		name    string
		args    args
		want    *sql.DB
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDb(tt.args.sqlite_path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDb() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDb() got = %v, want %v", got, tt.want)
			}
		})
	}
}
