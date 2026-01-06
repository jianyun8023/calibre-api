import { BaseApiService } from './base-service'
import { UnifiedApiClient, apiClient } from '../api-client-v2'
import { ErrorHandler, errorHandler } from '../error-handler'
import {
  MetadataDraft,
  MetadataChangelog,
  ExtractionSession,
  DraftFilter,
  ChangelogFilter,
  GovernanceStats,
  BatchItem,
  BatchResult,
  DraftListResponse,
  ChangelogListResponse,
} from '@/types/governance'

class GovernanceService extends BaseApiService {
  constructor(client: UnifiedApiClient, handler: ErrorHandler) {
    super(client, handler)
  }

  async listDrafts(filter: DraftFilter = {}): Promise<DraftListResponse> {
    const params: Record<string, unknown> = {}
    if (filter.status) params.status = filter.status
    if (filter.confidence_min !== undefined) params.confidence_min = filter.confidence_min
    if (filter.confidence_max !== undefined) params.confidence_max = filter.confidence_max
    if (filter.has_flags !== undefined) params.has_flags = filter.has_flags
    if (filter.session_id) params.session_id = filter.session_id
    if (filter.book_id) params.book_id = filter.book_id
    if (filter.field) params.field = filter.field
    if (filter.limit) params.limit = filter.limit
    if (filter.offset) params.offset = filter.offset

    const url = this.buildUrl('/api/metadata/drafts', params)
    return this.handleRequest(() => this.client.get<DraftListResponse>(url))
  }

  async getDraft(id: number): Promise<MetadataDraft> {
    return this.handleRequest(() =>
      this.client.get<MetadataDraft>(`/api/metadata/drafts/${id}`)
    )
  }

  async approveDraft(id: number, version: number): Promise<void> {
    return this.handleRequest(() =>
      this.client.post<void>(`/api/metadata/drafts/${id}/approve`, { version })
    )
  }

  async rejectDraft(id: number, version: number): Promise<void> {
    return this.handleRequest(() =>
      this.client.post<void>(`/api/metadata/drafts/${id}/reject`, { version })
    )
  }

  async updateDraft(id: number, newValue: string, version: number): Promise<void> {
    return this.handleRequest(() =>
      this.client.put<void>(`/api/metadata/drafts/${id}`, { new_value: newValue, version })
    )
  }

  async batchApprove(items: BatchItem[]): Promise<BatchResult> {
    return this.handleRequest(() =>
      this.client.post<BatchResult>('/api/metadata/drafts/batch', {
        action: 'approve',
        items,
      })
    )
  }

  async batchReject(items: BatchItem[]): Promise<BatchResult> {
    return this.handleRequest(() =>
      this.client.post<BatchResult>('/api/metadata/drafts/batch', {
        action: 'reject',
        items,
      })
    )
  }

  async applyDrafts(draftIds: number[]): Promise<BatchResult> {
    return this.handleRequest(() =>
      this.client.post<BatchResult>('/api/metadata/apply', { draft_ids: draftIds })
    )
  }

  async applyAll(): Promise<BatchResult> {
    return this.handleRequest(() =>
      this.client.post<BatchResult>('/api/metadata/apply-all', {})
    )
  }

  async listChangelogs(filter: ChangelogFilter = {}): Promise<ChangelogListResponse> {
    const params: Record<string, unknown> = {}
    if (filter.book_id) params.book_id = filter.book_id
    if (filter.field) params.field = filter.field
    if (filter.from_date) params.from_date = filter.from_date
    if (filter.to_date) params.to_date = filter.to_date
    if (filter.reverted !== undefined) params.reverted = filter.reverted
    if (filter.limit) params.limit = filter.limit
    if (filter.offset) params.offset = filter.offset

    const url = this.buildUrl('/api/metadata/changelog', params)
    return this.handleRequest(() => this.client.get<ChangelogListResponse>(url))
  }

  async getChangelog(id: number): Promise<MetadataChangelog> {
    return this.handleRequest(() =>
      this.client.get<MetadataChangelog>(`/api/metadata/changelog/${id}`)
    )
  }

  async revertChangelog(id: number, reason: string): Promise<void> {
    return this.handleRequest(() =>
      this.client.post<void>(`/api/metadata/changelog/${id}/revert`, { reason })
    )
  }

  async getStats(): Promise<GovernanceStats> {
    return this.handleRequest(() =>
      this.client.get<GovernanceStats>('/api/metadata/stats')
    )
  }

  async getSession(id: string): Promise<ExtractionSession> {
    return this.handleRequest(() =>
      this.client.get<ExtractionSession>(`/api/metadata/sessions/${id}`)
    )
  }
}

export const governanceService = new GovernanceService(apiClient, errorHandler)
