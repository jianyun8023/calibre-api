package calibre

import (
	"database/sql"
)
import _ "github.com/mattn/go-sqlite3"

const (
	queryAllBooks = `SELECT 
    books.id AS id,
    books.timestamp AS last_modified,
    books.pubdate AS pubdate,
    books.title AS title,
    group_concat(DISTINCT authors.name) AS authors,
    group_concat(DISTINCT authors.sort) AS authors_sort,
    group_concat(DISTINCT publishers.name) AS publisher,
    books.path || '/' || data.name || '.epub' AS file_path,
    books.path || '/' || 'cover.jpg' AS cover,
    data.uncompressed_size AS size,
    identifiers.val AS isbn
FROM
    books
LEFT JOIN
    books_authors_link ON books.id = books_authors_link.book
LEFT JOIN
    authors ON books_authors_link.author = authors.id
LEFT JOIN
    books_publishers_link ON books.id = books_publishers_link.book
LEFT JOIN
    publishers ON books_publishers_link.publisher = publishers.id
LEFT JOIN
    data ON books.id = data.book AND data.format = 'EPUB'
LEFT JOIN
    identifiers ON books.id = identifiers.book AND identifiers.type = 'isbn'
GROUP BY
    books.id`
)

type Db struct {
	dbPath string
	db     *sql.DB
}

func NewDb(dbPath string) (*Db, error) {
	db, err := getDb(dbPath)
	if err != nil {
		return nil, err
	}
	return &Db{dbPath: dbPath, db: db}, nil
}

func getDb(sqlitePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", sqlitePath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (d *Db) Close() error {
	return d.db.Close()
}

func (d Db) queryBooks() (books []Book, err error) {
	rows, err := d.db.Query(queryAllBooks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// rows to books
	for rows.Next() {
		var book BookRaw
		if err := rows.Scan(&book.ID, &book.LastModified, &book.Pubdate, &book.Title, &book.Authors, &book.AuthorSort, &book.Publisher, &book.FilePath, &book.Cover, &book.Size, &book.Isbn); err != nil {
			return nil, err
		}

		// convert BookRaw to Book
		newBook := Book{
			ID:           book.ID,
			AuthorSort:   book.AuthorSort,
			Authors:      book.Authors,
			Cover:        book.Cover,
			FilePath:     book.FilePath,
			Isbn:         book.Isbn.String,
			LastModified: book.LastModified,
			Pubdate:      book.Pubdate,
			Publisher:    book.Publisher.String,
			Size:         book.Size,
			Tags:         []string{},
			Title:        book.Title,
		}
		books = append(books, newBook)
	}
	return books, nil
}
