// src/types/book.ts
export interface Book {
    id: number
    title: string
    authors: string[]
    isbn: string
    publisher: string
    pubdate: string
    rating: number
    tags: string[]
    comments: string
    cover: string
}

export interface MetaBook {
    id: string;
    title: string;
    author: string[];
    translator: string[];
    summary: string;
    publisher: string;
    tags: Tag[];
    rating: Rating;
    series: null;
    image: string;
    url: string;
    isbn13: string;
    isbn10: null;
    pages: string;
    binding: string;
    price: string;
    catalog: null;
    origin_title: null;
    sub_title: string;
    pubdate: string;
    author_intro: null;
    ebook_url: null;
    ebook_price: null;
}

export interface Rating {
    average: string;
}

export interface Tag {
    name: string;
    title: string;
}

export function mapMetaBookToBook(metaBook: MetaBook): Book {
    return {
        id: 0,
        title: joinTitle(metaBook.title, metaBook.sub_title),
        authors: metaBook.author,
        isbn: metaBook.isbn13,
        publisher: metaBook.publisher,
        pubdate: parseDateString(metaBook.pubdate).toISOString(),
        rating: parseFloat(metaBook.rating.average),
        tags: metaBook.tags.map(tag => tag.name),
        comments: metaBook.summary,
        cover: metaBook.image.replace('subject/l/public', 'subject/s/public')
    };
}

export function updateBook(source: Book, target: Book){
    target.title = source.title
    target.authors = source.authors
    target.isbn = source.isbn
    target.publisher = source.publisher
    target.pubdate = source.pubdate
    target.rating = source.rating
    target.tags = source.tags
    target.comments = source.comments
    target.cover = source.cover
    target.id = source.id
}


function joinTitle(title: string, subTitle: string) {
    if (!subTitle) {
        return title
    }
    if (subTitle.length > 16) {
        return title
    }
    return title + "：" + subTitle
}

function parseDateString(dateString: string) {
    const dateParts = dateString.split('-')
    const year = parseInt(dateParts[0], 10)
    const month = parseInt(dateParts[1], 10) - 1 // JavaScript months are 0-based
    const day = dateParts.length === 3 ? parseInt(dateParts[2], 10) : 1 // Default to the first day of the month if day is not provided
    return new Date(year, month, day)
}