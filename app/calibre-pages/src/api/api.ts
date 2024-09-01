// src/api/api.ts
import {handleApiResponse} from "@/api/apiUtils";

export async function fetchPublishers() {
    const response = await fetch('/api/publisher');
    if (!response.ok) {
        throw new Error('Failed to fetch publishers');
    }
    return handleApiResponse(response);
}

export async function fetchRandomBooks() {
    const response = await fetch('/api/random?limit=12')
    if (!response.ok) {
        throw new Error('Failed to random');
    }
    return handleApiResponse(response);
}

export async function fetchRecentBooks(limit: number, offset: number) {
    const response = await fetch(`/api/recently?limit=${limit}&offset=${offset}`)
    if (!response.ok) {
        throw new Error('Failed to random');
    }
    return handleApiResponse(response);
}

export async function fetchBooks(keyword: string, filter: string[], limit: number, offset: number) {
    const response = await fetch('/api/search?q=' + keyword, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            Filter: filter,
            Limit: limit,
            Offset: offset,
        }),
    });
    if (!response.ok) {
        throw new Error('Failed to fetch books');
    }
    return handleApiResponse(response);
}

export async function deleteBook(bookId: number) {
    const response = await fetch(`/api/book/${bookId}/delete`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
    });
    if (!response.ok) {
        throw new Error('Failed to delete book');
    }
    return handleApiResponse(response);
}

export async function fetchBook(id: string) {
    try {
        const response = await fetch(`/api/book/${id}`);
        if (!response.ok) throw new Error('Network response was not ok');
        return handleApiResponse(response);
    } catch (error) {
        console.error('There was a problem with the fetch operation:', error);
        throw error;
    }
}

export async function updateBook(id: string, body: any) {
    try {
        const response = await fetch(`/api/book/${id}/update`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(body),
        });
        if (!response.ok) throw new Error('Network response was not ok');
        return handleApiResponse(response);
    } catch (error) {
        console.error('There was a problem with the fetch operation:', error);
        throw error;
    }
}

