export type BucketInfo = {
  name: string;
  created_at: string;
};

export type BucketListResponse = {
  buckets: BucketInfo[];
};

export type BucketResponse = {
  name: string;
  region?: string;
  created_at?: string;
};

export type DeleteBucketResponse = {
  deleted: string;
};

export type UploadObjectResponse = {
  bucket: string;
  object: string;
  content_type: string;
  size: number;
  etag: string;
  storage_path: string;
};

export type ErrorResponse = {
  error: string;
};

const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...(init.headers || {}) },
    ...init,
  });
  if (!res.ok) {
    const err = (await res.json().catch(() => ({}))) as Partial<ErrorResponse>;
    throw new Error(err.error || res.statusText);
  }
  return res.json();
}

export const api = {
  listBuckets: () => request<BucketListResponse>("/buckets"),
  getBucket: (name: string) => request<BucketResponse>(`/buckets/${encodeURIComponent(name)}`),
  createBucket: (name: string, region?: string) =>
    request<BucketResponse>("/buckets", {
      method: "POST",
      body: JSON.stringify({ name, region }),
    }),
  deleteBucket: (name: string) =>
    request<DeleteBucketResponse>(`/buckets/${encodeURIComponent(name)}`, { method: "DELETE" }),
  uploadObject: (bucket: string, file: File, objectName?: string, meta?: Record<string, string>) => {
    const form = new FormData();
    form.append("file", file);
    if (objectName) form.append("objectName", objectName);
    if (meta) {
      Object.entries(meta).forEach(([k, v]) => form.append(`meta-${k}`, v));
    }
    return fetch(`${API_BASE}/buckets/${encodeURIComponent(bucket)}/objects`, {
      method: "POST",
      body: form,
    }).then(async (res) => {
      if (!res.ok) {
        const err = (await res.json().catch(() => ({}))) as Partial<ErrorResponse>;
        throw new Error(err.error || res.statusText);
      }
      return res.json() as Promise<UploadObjectResponse>;
    });
  },
};
