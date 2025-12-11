// src/lib/api-client.ts
import { toast } from "sonner"

interface StandardResponse<T> {
    code: number
    data: T
    message: string
}

export async function apiRequest<T>(
  url: string, 
  options?: RequestInit
): Promise<T> {
  try {
      const res = await fetch(url, {
        headers: { 
            'Content-Type': 'application/json',
        },
        ...options
      })

      const json = await res.json() as StandardResponse<T>

      if (!res.ok) {
          throw new Error(json.message || `HTTP ${res.status}`)
      }

      if (json.code === 200) {
          return json.data
      } else {
          toast.error(`ERROR: ${json.message}`)
          throw new Error(json.message)
      }
  } catch (error) {
      console.error("API Request Failed:", error)
      const message = error instanceof Error ? error.message : "Network Error"
      toast.error(message)
      throw error
  }
}
