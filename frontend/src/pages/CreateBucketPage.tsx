import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";

export default function CreateBucketPage() {
  const [name, setName] = useState("");
  const [region, setRegion] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await api.createBucket(name, region || undefined);
      navigate(`/buckets/${res.name}`);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card">
      <form onSubmit={handleSubmit} className="grid">
        <div>
          <label>버킷 이름</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="ex) photos-kr-1"
            required
            minLength={3}
            maxLength={63}
          />
        </div>
        <div>
          <label>리전 (선택)</label>
          <input value={region} onChange={(e) => setRegion(e.target.value)} placeholder="ap-northeast-2" />
        </div>
        {error && <div className="muted">에러: {error}</div>}
        <div className="row">
          <button className="btn" type="submit" disabled={loading}>
            {loading ? "생성 중" : "생성"}
          </button>
        </div>
      </form>
    </div>
  );
}
