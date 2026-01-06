export type DraftStatus = 'pending' | 'approved' | 'rejected' | 'applied' | 'skipped'
export type SuggestedAction = 'auto_apply' | 'review' | 'skip'
export type MetadataSource = 'copyright_extract' | 'douban' | 'manual'
export type MetadataField = 'isbn' | 'title' | 'authors' | 'publisher' | 'pubdate'
export type DraftFlag =
  | 'collection_suspected'
  | 'multiple_isbn'
  | 'isbn_invalid_checksum'
  | 'title_too_long'
  | 'multiple_authors'
  | 'magazine_suspected'

export interface ConfidenceBreakdown {
  isbn_score: number
  context_score: number
  complexity_penalty: number
  final_score: number
  details?: string
}

export interface MetadataDraft {
  id: number
  book_id: number
  book_title: string
  field: MetadataField
  old_value: string
  new_value: string
  source: MetadataSource
  confidence: number
  confidence_breakdown?: ConfidenceBreakdown
  flags?: DraftFlag[]
  status: DraftStatus
  suggested_action: SuggestedAction
  session_id?: string
  version: number
  created_at: string
  reviewed_at?: string
  reviewed_by?: string
  applied_at?: string
}

export interface MetadataChangelog {
  id: number
  book_id: number
  book_title: string
  field: MetadataField
  old_value: string
  new_value: string
  source: MetadataSource
  draft_id?: number
  applied_at: string
  applied_by?: string
  reverted_at?: string
  reverted_by?: string
  revert_reason?: string
}

export type SessionState = 'running' | 'completed' | 'failed' | 'cancelled'

export interface ExtractionSession {
  id: string
  task_type: string
  mode: string
  total_books: number
  processed: number
  success: number
  failed: number
  skipped: number
  auto_approved: number
  pending_review: number
  state: SessionState
  error_message?: string
  started_at: string
  completed_at?: string
}

export interface DraftFilter {
  status?: DraftStatus
  confidence_min?: number
  confidence_max?: number
  has_flags?: boolean
  session_id?: string
  book_id?: number
  field?: MetadataField
  limit?: number
  offset?: number
}

export interface ChangelogFilter {
  book_id?: number
  field?: MetadataField
  from_date?: string
  to_date?: string
  reverted?: boolean
  limit?: number
  offset?: number
}

export interface GovernanceStats {
  drafts: {
    pending: number
    approved: number
    rejected: number
    applied: number
  }
  confidence_distribution: {
    high: number
    medium: number
    low: number
  }
  by_source: Record<MetadataSource, number>
  flags_count: Record<DraftFlag, number>
}

export interface BatchItem {
  id: number
  version: number
}

export interface BatchResult {
  success: number[]
  conflicts: number[]
  errors: number[]
}

export interface DraftListResponse {
  drafts: MetadataDraft[]
  total: number
}

export interface ChangelogListResponse {
  changelogs: MetadataChangelog[]
  total: number
}
