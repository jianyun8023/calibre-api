// src/utils/utils.ts
import { ElNotification } from 'element-plus'
import { h } from 'vue'

export function formatFileSize(size: number): string {
    if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB'
    return (size / 1024 / 1024).toFixed(2) + ' MB'
}

export function copyToClipboard(text: string): void {
    navigator.clipboard
        .writeText(text)
        .then(() => {
            ElNotification({
                title: 'ID copied ' + text,
                message: h('i', { style: 'color: teal' }, 'ID copied to clipboard'),
                type: 'success'
            })
        })
        .catch((err) => {
            ElNotification({
                title: 'ID copied ' + text,
                message: h('i', { style: 'color: red' }, 'Oops, Could not copy text.'),
                type: 'error'
            })
        })
}