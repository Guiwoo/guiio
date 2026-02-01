import { useState } from "react";

type Props = {
  bucket: string;
};

const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

export default function DownloadButton({ bucket }: Props) {
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handleDownload = async () => {
    if (!name) {
      setError("객체 이름을 입력하세요");
      return;
    }
    setError(null);
    const url = `${API_BASE}/buckets/${encodeURIComponent(bucket)}/objects/${encodeURIComponent(name)}`;
    try {
      const res = await fetch(url);
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || res.statusText);
      }
      const blob = await res.blob();
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = name;
      a.click();
      URL.revokeObjectURL(a.href);
    } catch (e: any) {
      setError(e.message);
    }
  };

  return (
    <div className="card">
      <div className="grid">
        <input
          placeholder="다운로드할 객체 이름"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <button className="btn" onClick={handleDownload}>
          다운로드
        </button>
        {error && <div className="muted">에러: {error}</div>}
      </div>
    </div>
  );
}
