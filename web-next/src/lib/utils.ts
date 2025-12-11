import { toast } from "sonner"
import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatFileSize(size: number): string {
    if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB'
    return (size / 1024 / 1024).toFixed(2) + ' MB'
}

export function copyToClipboard(text: string): void {
    navigator.clipboard
        .writeText(text)
        .then(() => {
            toast.success("Copied to clipboard", {
                description: text,
            })
        })
        .catch((err) => {
            toast.error("Failed to copy", {
                description: "Could not copy text to clipboard.",
            })
        })
}
