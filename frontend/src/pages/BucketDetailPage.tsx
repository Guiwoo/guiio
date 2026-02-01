import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api, BucketResponse } from "../api/client";
import UploadWidget from "../components/UploadWidget";
import DownloadButton from "../components/DownloadButton";

export default function BucketDetailPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const [bucket, setBucket] = useState<BucketResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    if (!name) return;
    api
      .getBucket(name)
      .then(setBucket)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [name]);

  const handleDelete = async () => {
    if (!name) return;
    setDeleting(true);
    try {
      await api.deleteBucket(name);
      navigate("/", { replace: true });
    } catch (err: any) {
      setError(err.message);
    } finally {
      setDeleting(false);
    }
  };

  if (loading) return <div className="card">불러오는 중...</div>;
  if (error) return <div className="card">에러: {error}</div>;
  if (!bucket) return <div className="card">버킷 정보를 찾을 수 없습니다.</div>;

  return (
    <div className="grid">
      <div className="card">
        <div className="row start">
          <div style={{ flex: 1 }}>
            <div style={{ fontWeight: 700, fontSize: 18 }}>{bucket.name}</div>
            <div className="muted">생성: {bucket.created_at ? new Date(bucket.created_at).toLocaleString() : "-"}</div>
            <div className="muted">리전: {bucket.region || "기본값"}</div>
          </div>
          <button className="btn secondary" disabled={deleting} onClick={handleDelete}>
            {deleting ? "삭제 중" : "삭제"}
          </button>
        </div>
      </div>

      <UploadWidget bucket={bucket.name} onUploaded={() => navigate(0)} />
      <DownloadButton bucket={bucket.name} />
    </div>
  );
}
