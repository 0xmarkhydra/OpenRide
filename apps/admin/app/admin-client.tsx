'use client';

import { FormEvent, useCallback, useEffect, useMemo, useState } from 'react';

type TokenPair = { access_token: string; refresh_token: string; expires_at: string };
type Driver = { id: string; phone?: string; full_name: string; service_type: string; capabilities?: string[]; approval_status: string; availability_status: string };
type Trip = { id: string; rider_id: string; driver_id?: string; customer_vehicle_id?: string; service_type: string; booking_mode?: string; scheduled_at?: string; status: string; estimated_fare_minor: number; final_fare_minor: number; inspection_result?: string; incident_open?: boolean; created_at: string };
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

function tripStatusLabel(status: string) {
  return ({
    scheduled: 'Đã hẹn lịch',
    searching: 'Đang tìm tài xế',
    accepted: 'Đã nhận việc',
    arriving: 'Đang đến nhận xe',
    arrived: 'Đã tới điểm nhận',
    arriving_for_pickup: 'Đang đến nhận xe',
    arrived_for_pickup: 'Đã tới điểm nhận',
    vehicle_received: 'Đã nhận xe khách',
    in_progress: 'Đang thực hiện',
    en_route_to_inspection: 'Đang tới nơi đăng kiểm',
    arrived_at_inspection_center: 'Đã tới nơi đăng kiểm',
    inspection_in_progress: 'Đang đăng kiểm',
    inspection_completed: 'Đã có kết quả đăng kiểm',
    returning_vehicle: 'Đang trả xe',
    arrived_for_return: 'Đã tới điểm trả',
    handover: 'Đang bàn giao',
    completed: 'Hoàn thành',
    cancelled: 'Đã hủy',
  } as Record<string, string>)[status] || status;
}

function serviceLabel(service: string) {
  return ({
    car: 'Lái hộ ô tô',
    designated_driver_car: 'Lái hộ ô tô',
    bike: 'Lái hộ xe máy',
    designated_driver_bike: 'Lái hộ xe máy',
    vehicle_inspection_assist: 'Đăng kiểm hộ',
  } as Record<string, string>)[service] || service;
}

function inspectionResultLabel(result?: string) {
  return ({ passed: 'Đạt', failed: 'Không đạt', deferred: 'Cần thực hiện lại', unavailable: 'Chưa có kết quả' } as Record<string, string>)[result || ''] || result || '';
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

  useEffect(() => {
    if (!token) return;
    void load(token);
    const timer = window.setInterval(() => void load(token), 5000);
    return () => window.clearInterval(timer);
  }, [token, load]);

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
    ['Tổng công việc', metrics.trips_total || 0, `${metrics.trips_completed || 0} hoàn thành`, '↗'],
    ['Tài xế online', metrics.drivers_online || 0, `${metrics.drivers_total || 0} tổng tài xế`, '●'],
    ['Đang vận hành', metrics.trips_active || 0, `${metrics.trips_searching || 0} đang tìm · ${metrics.trips_scheduled || 0} đã hẹn`, '⌖'],
    ['Cần chú ý', metrics.trips_incident || 0, `${metrics.drivers_pending || 0} tài xế chờ duyệt`, '!'],
  ], [metrics]);

  if (!token) {
    return <main className="loginShell"><section className="loginCard">
      <div className="brand loginBrand"><div className="brandMark">⚡</div><div className="brandText"><strong>FlashX</strong><span>Operations</span></div></div>
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
      <div className="brand"><div className="brandMark">⚡</div><div className="brandText"><strong>FlashX</strong><span>Operations</span></div></div>
      <nav className="nav" aria-label="Điều hướng quản trị">
        {['▦ Tổng quan','⇄ Công việc','◉ Tài xế','◎ Khách hàng','▣ Xe khách','₫ Bảng giá','! Sự cố','⌕ Nhật ký','⚙ Cài đặt'].map((item,i)=><div className={`navItem${i===0?' active':''}`} key={item}>{item}</div>)}
      </nav>
      <div className="sidebarFooter"><strong style={{color:'white'}}>MVP Operations</strong><br/>Dữ liệu trực tiếp từ FlashX API.</div>
    </aside>
    <main className="main">
      <header className="topbar"><h1>Trung tâm vận hành</h1><div className="topbarRight"><button className="linkButton" onClick={()=>void load()} disabled={busy}>Làm mới</button><span className="env">LIVE</span><button className="avatar" onClick={()=>void logout()}>AD</button></div></header>
      <div className="content" aria-busy={busy}>
        {busy && <div className="loadingBar" aria-label="Đang đồng bộ dữ liệu"><span /></div>}
        <section className="pageHeading"><div><div className="eyebrow">FLASHX OPERATIONS</div><h2>Trung tâm điều hành dịch vụ</h2><p>Giám sát Lái hộ ô tô, Lái hộ xe máy và Đăng kiểm hộ trên cùng một hệ thống.</p></div><div className="liveBadge"><span className="liveDot"/> Backend connected</div></section>
        {error && <div className="errorBox" role="alert">{error}</div>}
        <section className="metrics">{cards.map(([label,value,foot,icon])=><article className="metricCard" key={label}><div className="metricTop"><span>{label}</span><span className="metricIcon">{icon}</span></div><div className="metricValue">{value}</div><div className="metricFoot good">{foot}</div></article>)}</section>
        <section className="grid">
          <article className="card"><div className="cardHeader"><h3>Bản đồ vận hành</h3><span className="muted">Map adapter chờ API key khách hàng</span></div><div className="opsMap"><div className="mapRoad r1"/><div className="mapRoad r2"/><div className="mapRoad r3"/><div className="mapRoad r4"/><div className="mapLabel">{metrics.drivers_online || 0} tài xế online · {metrics.trips_active || 0} công việc đang thực hiện</div></div></article>
          <article className="card"><div className="cardHeader"><div><h3>Chờ duyệt tài xế</h3><span className="cardHint">Hồ sơ cần Operations xử lý</span></div><span className="countBadge">{drivers.length}</span></div><div className="queue">{drivers.length===0?<div className="emptyState"><span className="emptyIcon">✓</span><strong>Đã xử lý hết hồ sơ</strong><span>Hiện không có tài xế nào đang chờ duyệt.</span></div>:drivers.map(d=><div className="queueItem" key={d.id}><div className="queueAvatar">{(d.full_name||d.phone||'TX').slice(0,2).toUpperCase()}</div><div className="queueIdentity"><div className="queueTitle">{d.full_name||d.phone||d.id}</div><div className="queueSub">{serviceLabel(d.service_type)} · {d.phone||d.id}</div></div><div className="queueActions"><button className="rejectButton" disabled={busy} onClick={()=>void approve(d,'rejected')}>Từ chối</button><button className="approveButton" disabled={busy} onClick={()=>void approve(d,'approved')}>Duyệt</button></div></div>)}</div></article>
        </section>
        <section className="card tableCard"><div className="cardHeader"><div><h3>Công việc gần đây</h3><span className="cardHint">Theo dõi tiến trình của cả 3 dịch vụ</span></div><span className="countBadge">{trips.length}</span></div><div className="tableWrap">{trips.length===0?<div className="emptyState tableEmpty"><span className="emptyIcon">⇄</span><strong>Chưa có công việc</strong><span>Công việc sẽ xuất hiện khi khách đặt một trong ba dịch vụ FlashX.</span></div>:<table><thead><tr><th>Mã việc</th><th>Khách</th><th>Tài xế</th><th>Dịch vụ</th><th>Giá</th><th>Trạng thái</th></tr></thead><tbody>{trips.map(t=><tr key={t.id}><td className="tripId">{t.id}</td><td>{t.rider_id}</td><td>{t.driver_id||'—'}</td><td>{serviceLabel(t.service_type)}</td><td><strong>{money(t.final_fare_minor||t.estimated_fare_minor)}</strong></td><td><span className={`status ${t.incident_open?'cancelled':t.status==='completed'?'active':'pending'}`}>{t.incident_open?'Có sự cố':tripStatusLabel(t.status)}</span>{t.inspection_result&&<div className="cardHint">KQ: {inspectionResultLabel(t.inspection_result)}</div>}</td></tr>)}</tbody></table>}</div></section>
      </div>
    </main>
  </div>;
}
