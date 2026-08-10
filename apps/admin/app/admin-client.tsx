'use client';

import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react';

type TokenPair = { access_token: string; refresh_token: string; expires_at: string };
type Driver = { id: string; phone?: string; full_name: string; service_type: string; approval_status: string; availability_status: string };
type Trip = { id: string; rider_id: string; driver_id?: string; service_type: string; status: string; estimated_fare_minor: number; final_fare_minor: number; created_at: string };
type Metrics = Record<string, number>;

type ApiError = { error?: { message?: string; code?: string } };

class ApiHttpError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

async function api<T>(path: string, options: RequestInit = {}, token?: string): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set('content-type', 'application/json');
  if (token) headers.set('authorization', `Bearer ${token}`);
  const response = await fetch(`/api/flashx${path}`, { ...options, headers, cache: 'no-store' });
  const payload = (await response.json()) as ApiError & { data?: T };
  if (!response.ok) throw new ApiHttpError(response.status, payload.error?.message || `HTTP ${response.status}`);
  return payload.data as T;
}

function money(value?: number) {
  return new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND', maximumFractionDigits: 0 }).format(value || 0);
}

export default function AdminClient() {
  const [token, setToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [phone, setPhone] = useState(process.env.NEXT_PUBLIC_ADMIN_PHONE || '');
  const [challenge, setChallenge] = useState('');
  const [debugCode, setDebugCode] = useState('');
  const [code, setCode] = useState('');
  const [metrics, setMetrics] = useState<Metrics>({});
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [trips, setTrips] = useState<Trip[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    setToken(localStorage.getItem('flashx_admin_access_token'));
    setRefreshToken(localStorage.getItem('flashx_admin_refresh_token'));
  }, []);

  const clearSession = useCallback(() => {
    localStorage.removeItem('flashx_admin_access_token');
    localStorage.removeItem('flashx_admin_refresh_token');
    setToken(null);
    setRefreshToken(null);
  }, []);

  const refreshSession = useCallback(async (): Promise<string | null> => {
    const currentRefresh = refreshToken || localStorage.getItem('flashx_admin_refresh_token');
    if (!currentRefresh) return null;
    try {
      const tokens = await api<TokenPair>('/v1/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: currentRefresh }),
      });
      localStorage.setItem('flashx_admin_access_token', tokens.access_token);
      localStorage.setItem('flashx_admin_refresh_token', tokens.refresh_token);
      setToken(tokens.access_token);
      setRefreshToken(tokens.refresh_token);
      return tokens.access_token;
    } catch {
      clearSession();
      return null;
    }
  }, [clearSession, refreshToken]);

  const authedApi = useCallback(async <T,>(path: string, options: RequestInit = {}, access = token): Promise<T> => {
    if (!access) throw new ApiHttpError(401, 'Phiên đăng nhập đã hết hạn');
    try {
      return await api<T>(path, options, access);
    } catch (e) {
      if (!(e instanceof ApiHttpError) || e.status !== 401) throw e;
      const refreshed = await refreshSession();
      if (!refreshed) throw e;
      return api<T>(path, options, refreshed);
    }
  }, [refreshSession, token]);

  const load = useCallback(async (access = token) => {
    if (!access) return;
    setBusy(true);
    try {
      const [dashboard, pending, tripData] = await Promise.all([
        authedApi<Metrics>('/v1/admin/dashboard', {}, access),
        authedApi<Driver[]>('/v1/admin/drivers?approval=pending', {}, access),
        authedApi<Trip[]>('/v1/admin/trips', {}, access),
      ]);
      setMetrics(dashboard);
      setDrivers(pending);
      setTrips(tripData);
      setError('');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Không thể tải dữ liệu');
    } finally {
      setBusy(false);
    }
  }, [authedApi, token]);

  useEffect(() => { if (token) void load(token); }, [token, load]);

  async function requestOtp(e: FormEvent) {
    e.preventDefault(); setBusy(true); setError('');
    try {
      const result = await api<{ challenge_id: string; debug_code?: string }>('/v1/auth/otp/request', {
        method: 'POST', body: JSON.stringify({ phone, role: 'admin' }),
      });
      setChallenge(result.challenge_id); setDebugCode(result.debug_code || '');
    } catch (e) { setError(e instanceof Error ? e.message : 'Không thể gửi OTP'); }
    finally { setBusy(false); }
  }

  async function verifyOtp(e: FormEvent) {
    e.preventDefault(); setBusy(true); setError('');
    try {
      const result = await api<{ tokens: TokenPair }>('/v1/auth/otp/verify', {
        method: 'POST', body: JSON.stringify({ challenge_id: challenge, role: 'admin', code }),
      });
      const tokens = result.tokens;
      localStorage.setItem('flashx_admin_access_token', tokens.access_token);
      localStorage.setItem('flashx_admin_refresh_token', tokens.refresh_token);
      setToken(tokens.access_token); setRefreshToken(tokens.refresh_token);
      setChallenge(''); setCode('');
    } catch (e) { setError(e instanceof Error ? e.message : 'OTP không hợp lệ'); }
    finally { setBusy(false); }
  }

  async function approve(driver: Driver, status: 'approved' | 'rejected') {
    if (!token) return;
    setBusy(true); setError('');
    try {
      await authedApi(`/v1/admin/drivers/${driver.id}/approval`, {
        method: 'POST', body: JSON.stringify({ status, reason: status === 'rejected' ? 'Từ chối bởi vận hành' : '' }),
      }, token);
      await load();
    } catch (e) { setError(e instanceof Error ? e.message : 'Không thể cập nhật tài xế'); }
    finally { setBusy(false); }
  }

  async function logout() {
    if (refreshToken) {
      try { await api('/v1/auth/logout', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) }); } catch {}
    }
    clearSession();
  }

  const cards = useMemo(() => [
    ['Tổng chuyến', metrics.trips_total || 0, `${metrics.trips_completed || 0} hoàn thành`],
    ['Tài xế online', metrics.drivers_online || 0, `${metrics.drivers_total || 0} tổng tài xế`],
    ['Đang tìm tài xế', metrics.trips_searching || 0, `${metrics.trips_active || 0} chuyến active`],
    ['Chờ duyệt', metrics.drivers_pending || 0, `${metrics.trips_cancelled || 0} chuyến đã hủy`],
  ], [metrics]);

  if (!token) {
    return <main className="loginShell"><section className="loginCard">
      <div className="brand loginBrand"><div className="brandMark">G</div><div className="brandText"><strong>FlashX</strong><span>Operations</span></div></div>
      <h1>Đăng nhập quản trị</h1><p className="muted">OTP chỉ được gửi tới số điện thoại Admin đã bootstrap trên backend.</p>
      <form onSubmit={challenge ? verifyOtp : requestOtp} className="loginForm">
        <label>Số điện thoại<input value={phone} onChange={e => setPhone(e.target.value)} disabled={!!challenge || busy} placeholder="+84..." /></label>
        {challenge && <label>Mã OTP<input value={code} onChange={e => setCode(e.target.value)} maxLength={6} inputMode="numeric" placeholder="000000" /></label>}
        {debugCode && <div className="devOtp">DEV OTP: <strong>{debugCode}</strong></div>}
        {error && <div className="errorBox">{error}</div>}
        <button className="primaryButton" disabled={busy}>{busy ? 'Đang xử lý…' : challenge ? 'Xác nhận OTP' : 'Gửi OTP'}</button>
      </form>
    </section></main>;
  }

  return <div className="shell">
    <aside className="sidebar">
      <div className="brand"><div className="brandMark">G</div><div className="brandText"><strong>FlashX</strong><span>Operations</span></div></div>
      <nav className="nav" aria-label="Điều hướng quản trị">
        {['▦ Tổng quan','⇄ Chuyến đi','◉ Tài xế','◎ Khách hàng','₫ Giá cước','% Khuyến mại','⌕ Audit log','⚙ Cài đặt'].map((item,i)=><div className={`navItem${i===0?' active':''}`} key={item}>{item}</div>)}
      </nav>
      <div className="sidebarFooter"><strong style={{color:'white'}}>MVP Operations</strong><br/>Dữ liệu trực tiếp từ FlashX API.</div>
    </aside>
    <main className="main">
      <header className="topbar"><h1>Trung tâm vận hành</h1><div className="topbarRight"><button className="linkButton" onClick={()=>void load()} disabled={busy}>Làm mới</button><span className="env">LIVE</span><button className="avatar" onClick={()=>void logout()}>AD</button></div></header>
      <div className="content">
        <section className="pageHeading"><div><h2>Tổng quan hoạt động</h2><p>Giám sát chuyến đi và duyệt tài xế trước khi vận hành.</p></div><div className="liveBadge"><span className="liveDot"/> Backend connected</div></section>
        {error && <div className="errorBox">{error}</div>}
        <section className="metrics">{cards.map(([label,value,foot])=><article className="metricCard" key={label}><div className="metricTop"><span>{label}</span></div><div className="metricValue">{value}</div><div className="metricFoot good">{foot}</div></article>)}</section>
        <section className="grid">
          <article className="card"><div className="cardHeader"><h3>Bản đồ vận hành</h3><span className="muted">Map adapter chờ API key khách hàng</span></div><div className="opsMap"><div className="mapRoad r1"/><div className="mapRoad r2"/><div className="mapRoad r3"/><div className="mapRoad r4"/><div className="mapLabel">{metrics.drivers_online || 0} tài xế online · {metrics.trips_active || 0} chuyến active</div></div></article>
          <article className="card"><div className="cardHeader"><h3>Chờ duyệt tài xế</h3><span className="muted">{drivers.length} hồ sơ</span></div><div className="queue">{drivers.length===0?<div className="emptyState">Không có hồ sơ chờ duyệt.</div>:drivers.map(d=><div className="queueItem" key={d.id}><div className="queueAvatar">{(d.full_name||d.phone||'TX').slice(0,2).toUpperCase()}</div><div><div className="queueTitle">{d.full_name||d.phone||d.id}</div><div className="queueSub">{d.service_type} · {d.id}</div></div><div className="queueActions"><button className="rejectButton" onClick={()=>void approve(d,'rejected')}>Từ chối</button><button className="approveButton" onClick={()=>void approve(d,'approved')}>Duyệt</button></div></div>)}</div></article>
        </section>
        <section className="card tableCard"><div className="cardHeader"><h3>Chuyến gần đây</h3><span className="muted">{trips.length} chuyến</span></div><div className="tableWrap"><table><thead><tr><th>Mã chuyến</th><th>Rider</th><th>Driver</th><th>Dịch vụ</th><th>Giá</th><th>Trạng thái</th></tr></thead><tbody>{trips.map(t=><tr key={t.id}><td className="tripId">{t.id}</td><td>{t.rider_id}</td><td>{t.driver_id||'—'}</td><td>{t.service_type}</td><td><strong>{money(t.final_fare_minor||t.estimated_fare_minor)}</strong></td><td><span className={`status ${t.status==='cancelled'?'cancelled':t.status==='completed'?'active':'pending'}`}>{t.status}</span></td></tr>)}</tbody></table></div></section>
      </div>
    </main>
  </div>;
}
