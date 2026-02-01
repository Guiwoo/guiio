import { Link, Outlet, useLocation } from "react-router-dom";

export default function AppShell() {
  const location = useLocation();
  return (
    <div className="shell">
      <nav className="nav">
        <div>
          <h1>guiio 콘솔</h1>
          <div className="muted">AWS S3 느낌의 버킷 관리</div>
        </div>
        <div className="row">
          <Link to="/" className="btn secondary">
            버킷 목록
          </Link>
          <Link to="/buckets/new" className="btn">
            새 버킷
          </Link>
        </div>
      </nav>
      <main>
        <Outlet key={location.pathname} />
      </main>
    </div>
  );
}
