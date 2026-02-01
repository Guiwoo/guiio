import { useState } from "react";
import { api } from "../api/client";

type Props = {
  bucket: string;
  onUploaded?: () => void;
};

export default function UploadWidget({ bucket, onUploaded }: Props) {
  const [file, setFile] = useState<File | null>(null);
  const [name, setName] = useState("");
  const [meta, setMeta] = useState("meta=demo");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleUpload = async () => {
    if (!file) {
      setError("파일을 선택하세요");
      return;
    }
    setError(null);
    setLoading(true);
    try {
      const metaMap: Record<string, string> = {};
      if (meta) {
        meta.split(",").forEach((p) => {
          const [k, v] = p.split("=");
          if (k && v) metaMap[k.trim()] = v.trim();
        });
      }
      await api.uploadObject(bucket, file, name || undefined, metaMap);
      setFile(null);
      setName("");
      if (onUploaded) onUploaded();
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card">
      <div className="grid">
        <div className="row">
          <input type="file" onChange={(e) => setFile(e.target.files?.[0] || null)} />
          <button className="btn" onClick={handleUpload} disabled={loading}>
            {loading ? "업로드 중" : "업로드"}
          </button>
        </div>
        <input
          placeholder="객체 이름(선택)"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <input
          placeholder="메타데이터 예: author=me,env=dev"
          value={meta}
          onChange={(e) => setMeta(e.target.value)}
        />
        {error && <div className="muted">에러: {error}</div>}
      </div>
    </div>
  );
}
