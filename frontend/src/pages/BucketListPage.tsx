import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, BucketInfo } from "../api/client";

export default function BucketListPage() {
  const [buckets, setBuckets] = useState<BucketInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .listBuckets()
      .then((res) => setBuckets(res.buckets))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="card">불러오는 중...</div>;
  if (error) return <div className="card">에러: {error}</div>;

  return (
    <div className="grid">
      {buckets.length === 0 && <div className="card">아직 버킷이 없습니다.</div>}
      {buckets.map((b) => (
        <div className="card" key={b.name}>
          <div className="row">
            <div style={{ flex: 1 }}>
              <div style={{ fontWeight: 700 }}>{b.name}</div>
              <div className="muted">{new Date(b.created_at).toLocaleString()}</div>
            </div>
            <Link to={`/buckets/${b.name}`} className="btn secondary">
              상세
            </Link>
          </div>
        </div>
      ))}
    </div>
  );
}
