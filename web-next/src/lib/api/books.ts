import { apiRequest } from "@/lib/api-client"
import { Book, PaginatedResponse } from "@/types/book"

// 补充定义 PaginatedResponse 泛型，如果 types/book.ts 里没有定义
interface BookListResponse {
    total: number
    records: Book[]
    next_cursor?: string
}

export async function fetchPublishers() {
    return apiRequest<string[]>('/api/publisher');
}

export async function fetchRandomBooks() {
    // /api/random 返回数组，不是 BookListResponse 格式
    const books = await apiRequest<Book[]>('/api/random?limit=5')
    return { total: books.length, records: books }
}

export async function fetchRecentBooks(limit: number, offset: number) {
    return apiRequest<BookListResponse>(`/api/recently?limit=${limit}&offset=${offset}`)
}

export async function fetchAllBooks(limit: number, cursor: string = '') {
    const url = cursor 
        ? `/api/books/all?limit=${limit}&cursor=${encodeURIComponent(cursor)}`
        : `/api/books/all?limit=${limit}`
    return apiRequest<BookListResponse>(url)
}

export async function fetchBooks(keyword: string, filter: string[], limit: number, offset: number, sort?: string[], mode?: string) {
    return apiRequest<BookListResponse>('/api/search?q=' + keyword + '&mode=' + (mode || 'hybrid'), {
        method: 'POST',
        body: JSON.stringify({
            Filter: filter,
            Limit: limit,
            Offset: offset,
            Sort: sort || [],
        }),
    });
}

export async function searchSemantic(query: string, limit: number = 12) {
    return apiRequest<BookListResponse>(`/api/search/semantic?q=${encodeURIComponent(query)}&limit=${limit}`);
}

export async function deleteBook(bookId: number) {
    return apiRequest(`/api/book/${bookId}/delete`, {
        method: 'POST',
    });
}

export async function fetchBook(id: string) {
    return apiRequest<Book>(`/api/book/${id}`);
}

export async function updateBook(id: string | number, body: any) {
    return apiRequest<Book>(`/api/book/${id}/update`, {
        method: 'POST',
        body: JSON.stringify(body),
    });
}

export async function fetchBookToc(id: string) {
    // TOC API returns data directly, not wrapped in standard response format
    const response = await fetch(`/api/read/${id}/toc`);
    
    if (!response.ok) {
        throw new Error(`Failed to fetch TOC: ${response.status} ${response.statusText}`);
    }
    
    return response.json();
}

export async function fetchChapterContent(bookId: string, filePath: string) {
    // Remove the /read/{id}/file/ prefix if present
    const cleanPath = filePath.replace(`/read/${bookId}/file/`, '');
    const response = await fetch(`/api/read/${bookId}/file/${cleanPath}`);
    
    if (!response.ok) {
        throw new Error(`Failed to fetch chapter content: ${response.status} ${response.statusText}`);
    }
    
    return response.text();
}

