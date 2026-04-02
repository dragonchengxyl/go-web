export const POST_DRAFT_STORAGE_KEY = "post_draft";

export interface StoredPostDraft {
  title?: string;
  content?: string;
  tags?: string;
  visibility?: string;
  source?: string;
  agent_run_id?: string;
}

export function readStoredPostDraft(): StoredPostDraft | null {
  if (typeof window === "undefined") return null;
  const raw = window.localStorage.getItem(POST_DRAFT_STORAGE_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as StoredPostDraft;
    return parsed && typeof parsed === "object" ? parsed : null;
  } catch {
    return null;
  }
}

export function writeStoredPostDraft(value: StoredPostDraft) {
  if (typeof window === "undefined") return;
  window.localStorage.setItem(POST_DRAFT_STORAGE_KEY, JSON.stringify(value));
}

export function clearStoredPostDraft() {
  if (typeof window === "undefined") return;
  window.localStorage.removeItem(POST_DRAFT_STORAGE_KEY);
}
