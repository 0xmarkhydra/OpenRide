'use client';

import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ColumnDef } from '@tanstack/react-table';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import {
  AlertTriangle, BadgeDollarSign, CarFront, ChevronRight, ClipboardList, Gauge, LayoutDashboard,
  ListChecks, LogOut, Menu, RefreshCw, SearchCheck, Settings, ShieldCheck, Sparkles, UserRound,
  UsersRound, X, Zap, type LucideIcon,
} from 'lucide-react';
import { DataTable, dataTableFeatures } from '../components/data-table';
import { Badge } from '../components/ui/badge';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Input } from '../components/ui/input';

type TokenPair = { access_token: string; refresh_token: string; expires_at: string };
type Driver = { id: string; phone?: string; full_name: string; service_type: string; capabilities?: string[]; approval_status: string; availability_status: string };
type Trip = { id: string; rider_id: string; driver_id?: string; customer_vehicle_id?: string; service_type: string; booking_mode?: string; scheduled_at?: string; status: string; estimated_fare_minor: number; final_fare_minor: number; inspection_result?: string; incident_open?: boolean; created_at: string };
type Metrics = Record<string, number>;
type ApiError = { error?: { message?: string; code?: string } };
type ViewKey = 'overview' | 'trips' | 'drivers' | 'customers' | 'vehicles' | 'pricing' | 'incidents' | 'audit' | 'settings';
type SimpleRow = { id: string; subtitle: string };
type NavItem = { key: ViewKey; label: string; icon: LucideIcon };

class ApiHttpError extends Error { constructor(public status: number, message: string) { super(message); } }

async function api<T>(path: string, options: RequestInit = {}, token?: string): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set('content-type', 'application/json');
  if (token) headers.set('authorization', `Bearer ${token}`);
  const response = await fetch(`/api/flashx${path}`, { ...options, headers, cache: 'no-store' });
  const payload = (await response.json()) as ApiError & { data?: T };
  if (!response.ok) throw new ApiHttpError(response.status, payload.error?.message || `HTTP ${response.status}`);
  return payload.data as T;
}

function money(value?: number) { return new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND', maximumFractionDigits: 0 }).format(value || 0); }
function serviceLabel(service: string) { return ({ car:'Lái hộ ô tô', designated_driver_car:'Lái hộ ô tô', bike:'Lái hộ xe máy', designated_driver_bike:'Lái hộ xe máy', vehicle_inspection_assist:'Đăng kiểm hộ' } as Record<string,string>)[service] || service; }
function tripStatusLabel(status: string) { return ({ scheduled:'Đã hẹn lịch', searching:'Đang tìm tài xế', accepted:'Đã nhận việc', arriving:'Đang đến nhận xe', arrived:'Đã tới điểm nhận', arriving_for_pickup:'Đang đến nhận xe', arrived_for_pickup:'Đã tới điểm nhận', vehicle_received:'Đã nhận xe khách', in_progress:'Đang thực hiện', en_route_to_inspection:'Đang tới nơi đăng kiểm', arrived_at_inspection_center:'Đã tới nơi đăng kiểm', inspection_in_progress:'Đang đăng kiểm', inspection_completed:'Đã có kết quả đăng kiểm', returning_vehicle:'Đang trả xe', arrived_for_return:'Đã tới điểm trả', handover:'Đang bàn giao', completed:'Hoàn thành', cancelled:'Đã hủy' } as Record<string,string>)[status] || status; }
function statusVariant(status: string, incident?: boolean): 'success'|'warning'|'destructive'|'secondary' { if (incident || status === 'cancelled') return 'destructive'; if (status === 'completed') return 'success'; if (status === 'searching' || status === 'scheduled') return 'warning'; return 'secondary'; }

const phoneSchema = z.object({ phone: z.string().trim().min(9, 'Nhập số điện thoại Admin') });
const otpSchema = z.object({ code: z.string().regex(/^\d{6}$/, 'OTP phải gồm 6 chữ số') });
type PhoneForm = z.infer<typeof phoneSchema>; type OTPForm = z.infer<typeof otpSchema>;

const navItems: NavItem[] = [
  { key:'overview', label:'Tổng quan', icon:LayoutDashboard }, { key:'trips', label:'Công việc', icon:ClipboardList },
  { key:'drivers', label:'Tài xế', icon:UserRound }, { key:'customers', label:'Khách hàng', icon:UsersRound },
  { key:'vehicles', label:'Xe khách', icon:CarFront }, { key:'pricing', label:'Bảng giá', icon:BadgeDollarSign },
  { key:'incidents', label:'Sự cố', icon:AlertTriangle }, { key:'audit', label:'Nhật ký', icon:ListChecks },
  { key:'settings', label:'Cài đặt', icon:Settings },
];

export default function AdminClient() {
  const [token, setToken] = useState<string | null>(null); const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [challenge, setChallenge] = useState(''); const [debugCode, setDebugCode] = useState('');
  const [metrics, setMetrics] = useState<Metrics>({}); const [drivers, setDrivers] = useState<Driver[]>([]); const [trips, setTrips] = useState<Trip[]>([]);
  const [busy, setBusy] = useState(false); const [hasLoaded, setHasLoaded] = useState(false); const [error, setError] = useState(''); const [activeView, setActiveView] = useState<ViewKey>('overview'); const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const phoneForm = useForm<PhoneForm>({ resolver:zodResolver(phoneSchema), defaultValues:{ phone:process.env.NEXT_PUBLIC_ADMIN_PHONE || '' } });
  const otpForm = useForm<OTPForm>({ resolver:zodResolver(otpSchema), defaultValues:{ code:'' } });

  useEffect(() => { setToken(localStorage.getItem('flashx_admin_access_token')); setRefreshToken(localStorage.getItem('flashx_admin_refresh_token')); }, []);
  const clearSession = useCallback(() => { localStorage.removeItem('flashx_admin_access_token'); localStorage.removeItem('flashx_admin_refresh_token'); setToken(null); setRefreshToken(null); setHasLoaded(false); }, []);
  const refreshSession = useCallback(async () => { const current = refreshToken || localStorage.getItem('flashx_admin_refresh_token'); if (!current) return null; try { const tokens = await api<TokenPair>('/v1/auth/refresh',{ method:'POST', body:JSON.stringify({ refresh_token:current }) }); localStorage.setItem('flashx_admin_access_token', tokens.access_token); localStorage.setItem('flashx_admin_refresh_token', tokens.refresh_token); setToken(tokens.access_token); setRefreshToken(tokens.refresh_token); return tokens.access_token; } catch { clearSession(); return null; } }, [clearSession, refreshToken]);
  const authedApi = useCallback(async <T,>(path:string, options:RequestInit={}, access=token):Promise<T> => { if (!access) throw new ApiHttpError(401,'Phiên đăng nhập đã hết hạn'); try { return await api<T>(path,options,access); } catch (requestError) { if (!(requestError instanceof ApiHttpError) || requestError.status !== 401) throw requestError; const refreshed = await refreshSession(); if (!refreshed) throw requestError; return api<T>(path,options,refreshed); } }, [refreshSession, token]);
  const load = useCallback(async (access=token) => { if (!access) return; setBusy(true); try { const [dashboard,driverData,tripData] = await Promise.all([authedApi<Metrics>('/v1/admin/dashboard',{},access), authedApi<Driver[]>('/v1/admin/drivers',{},access), authedApi<Trip[]>('/v1/admin/trips',{},access)]); setMetrics(dashboard); setDrivers(driverData); setTrips(tripData); setHasLoaded(true); setError(''); } catch (loadError) { setError(loadError instanceof Error ? loadError.message : 'Không thể tải dữ liệu'); } finally { setBusy(false); } }, [authedApi, token]);
  useEffect(() => { if (!token) return; void load(token); const timer = window.setInterval(() => void load(token),5000); return () => window.clearInterval(timer); }, [token,load]);

  const requestOtp = phoneForm.handleSubmit(async ({ phone }) => { setBusy(true); setError(''); try { const result = await api<{challenge_id:string;debug_code?:string}>('/v1/auth/otp/request',{ method:'POST', body:JSON.stringify({ phone, role:'admin' }) }); setChallenge(result.challenge_id); setDebugCode(result.debug_code || ''); } catch (requestError) { setError(requestError instanceof Error ? requestError.message : 'Không thể gửi OTP'); } finally { setBusy(false); } });
  const verifyOtp = otpForm.handleSubmit(async ({ code }) => { setBusy(true); setError(''); try { const result = await api<{tokens:TokenPair}>('/v1/auth/otp/verify',{ method:'POST', body:JSON.stringify({ challenge_id:challenge, role:'admin', code }) }); localStorage.setItem('flashx_admin_access_token', result.tokens.access_token); localStorage.setItem('flashx_admin_refresh_token', result.tokens.refresh_token); setToken(result.tokens.access_token); setRefreshToken(result.tokens.refresh_token); setChallenge(''); setDebugCode(''); otpForm.reset(); } catch (verifyError) { setError(verifyError instanceof Error ? verifyError.message : 'OTP không hợp lệ'); } finally { setBusy(false); } });
  async function approve(driver:Driver,status:'approved'|'rejected') { if (!token) return; setBusy(true); setError(''); try { await authedApi(`/v1/admin/drivers/${driver.id}/approval`,{ method:'POST', body:JSON.stringify({ status, reason:status==='rejected'?'Từ chối bởi vận hành':'' }) },token); await load(); } catch (approveError) { setError(approveError instanceof Error ? approveError.message : 'Không thể cập nhật tài xế'); } finally { setBusy(false); } }
  async function logout() { if (refreshToken) { try { await api('/v1/auth/logout',{ method:'POST', body:JSON.stringify({ refresh_token:refreshToken }) }); } catch {} } clearSession(); }

  const pendingDrivers = useMemo(() => drivers.filter(driver => driver.approval_status === 'pending'),[drivers]);
  const incidents = useMemo(() => trips.filter(trip => trip.incident_open),[trips]);
  const customers = useMemo<SimpleRow[]>(() => Array.from(new Set(trips.map(trip => trip.rider_id))).filter(Boolean).map(id => ({ id, subtitle:`${trips.filter(trip => trip.rider_id === id).length} công việc` })),[trips]);
  const vehicles = useMemo<SimpleRow[]>(() => Array.from(new Set(trips.map(trip => trip.customer_vehicle_id).filter(Boolean) as string[])).map(id => ({ id, subtitle:`${trips.filter(trip => trip.customer_vehicle_id === id).length} công việc` })),[trips]);

  const tripColumns = useMemo<Array<ColumnDef<typeof dataTableFeatures,Trip>>>(() => [
    { accessorKey:'id', header:'Mã việc', cell:({row}) => <span className='font-mono text-xs font-semibold text-foreground'>{row.original.id}</span> },
    { accessorKey:'service_type', header:'Dịch vụ', cell:({row}) => serviceLabel(row.original.service_type) }, { accessorKey:'rider_id', header:'Khách' },
    { accessorKey:'driver_id', header:'Tài xế', cell:({row}) => row.original.driver_id || '—' },
    { id:'fare', header:'Giá', cell:({row}) => <strong className='font-semibold'>{money(row.original.final_fare_minor || row.original.estimated_fare_minor)}</strong> },
    { accessorKey:'status', header:'Trạng thái', cell:({row}) => <Badge variant={statusVariant(row.original.status,row.original.incident_open)}>{row.original.incident_open?'Có sự cố':tripStatusLabel(row.original.status)}</Badge> },
  ],[]);
  const driverColumns = useMemo<Array<ColumnDef<typeof dataTableFeatures,Driver>>>(() => [
    { accessorKey:'full_name', header:'Tài xế', cell:({row}) => <div><div className='font-semibold text-foreground'>{row.original.full_name || row.original.phone || row.original.id}</div><div className='text-xs text-muted'>{row.original.phone || row.original.id}</div></div> },
    { accessorKey:'service_type', header:'Dịch vụ', cell:({row}) => serviceLabel(row.original.service_type) },
    { accessorKey:'availability_status', header:'Trạng thái', cell:({row}) => <Badge variant={row.original.availability_status === 'online'?'success':'secondary'}>{row.original.availability_status || 'offline'}</Badge> },
    { accessorKey:'approval_status', header:'Duyệt', cell:({row}) => <Badge variant={row.original.approval_status === 'approved'?'success':row.original.approval_status === 'rejected'?'destructive':'warning'}>{row.original.approval_status}</Badge> },
    { id:'actions', header:'Thao tác', cell:({row}) => row.original.approval_status === 'pending' ? <div className='flex gap-2'><Button size='sm' variant='destructive' disabled={busy} onClick={() => void approve(row.original,'rejected')}>Từ chối</Button><Button size='sm' disabled={busy} onClick={() => void approve(row.original,'approved')}>Duyệt</Button></div> : null },
  ],[busy]);
  const simpleColumns = useMemo<Array<ColumnDef<typeof dataTableFeatures,SimpleRow>>>(() => [ { accessorKey:'id', header:'Mã', cell:({row}) => <span className='font-mono text-xs font-semibold'>{row.original.id}</span> }, { accessorKey:'subtitle', header:'Hoạt động' } ],[]);

  if (!token) return <LoginScreen challenge={challenge} debugCode={debugCode} busy={busy} error={error} phoneForm={phoneForm} otpForm={otpForm} requestOtp={requestOtp} verifyOtp={verifyOtp} resetChallenge={() => { setChallenge(''); setDebugCode(''); }} />;

  const activeLabel = navItems.find(item => item.key === activeView)?.label || 'Tổng quan';
  const selectView = (view:ViewKey) => { setActiveView(view); setMobileNavOpen(false); };
  const metricsCards = [
    { label:'Tổng công việc', value:metrics.trips_total || 0, foot:`${metrics.trips_completed || 0} hoàn thành`, icon:ClipboardList },
    { label:'Tài xế online', value:metrics.drivers_online || 0, foot:`${metrics.drivers_total || 0} tổng tài xế`, icon:UserRound },
    { label:'Đang vận hành', value:metrics.trips_active || 0, foot:`${metrics.trips_searching || 0} đang tìm`, icon:Gauge },
    { label:'Cần chú ý', value:metrics.trips_incident || 0, foot:`${metrics.drivers_pending || 0} tài xế chờ duyệt`, icon:AlertTriangle },
  ];
  const initialLoading = busy && !hasLoaded;

  function renderView() {
    if (activeView === 'trips') return <Section title='Công việc' description='Theo dõi toàn bộ vòng đời của ba dịch vụ.'><DataTable columns={tripColumns} data={trips} loading={initialLoading} searchPlaceholder='Tìm mã việc, khách, tài xế...' /></Section>;
    if (activeView === 'drivers') return <Section title='Tài xế' description='Duyệt hồ sơ và theo dõi trạng thái cung ứng.'><DataTable columns={driverColumns} data={drivers} loading={initialLoading} searchPlaceholder='Tìm tài xế...' /></Section>;
    if (activeView === 'customers') return <Section title='Khách hàng' description='Khách hàng phát sinh từ dữ liệu công việc hiện tại.'><DataTable columns={simpleColumns} data={customers} loading={initialLoading} searchPlaceholder='Tìm khách hàng...' /></Section>;
    if (activeView === 'vehicles') return <Section title='Xe khách' description='Các xe đã được gắn với công việc FlashX.'><DataTable columns={simpleColumns} data={vehicles} loading={initialLoading} searchPlaceholder='Tìm xe...' /></Section>;
    if (activeView === 'incidents') return <Section title='Sự cố' description='Các công việc cần Operations xử lý.'><DataTable columns={tripColumns} data={incidents} loading={initialLoading} searchPlaceholder='Tìm sự cố...' /></Section>;
    if (activeView === 'pricing') return <Section title='Bảng giá' description='Giá hiện do backend tính theo dịch vụ và hành trình.'><InfoState icon={BadgeDollarSign} title='Chuẩn bị mở quản trị bảng giá' text='Design đã sẵn sàng cho pricing version, thành phần giá và lịch sử hiệu lực.' /></Section>;
    if (activeView === 'audit') return <Section title='Nhật ký' description='Theo dõi các thao tác quản trị quan trọng.'><InfoState icon={SearchCheck} title='Audit đã được ghi ở backend' text='Màn đọc audit sẽ kết nối khi API danh sách nhật ký được expose.' /></Section>;
    if (activeView === 'settings') return <Section title='Cài đặt' description='Cấu hình môi trường vận hành FlashX.'><div className='grid gap-4 md:grid-cols-2'><InfoState icon={ShieldCheck} title='Môi trường dev' text='Admin xác thực bằng OTP development; dữ liệu lấy trực tiếp từ FlashX-BE.' /><InfoState icon={Settings} title='Cấu hình có kiểm soát' text='Secret và provider config tiếp tục nằm trên Railway, không đưa xuống client.' /></div></Section>;

    return <div className='flex flex-col gap-5 fx-enter'>
      <section className='fx-flowline overflow-hidden rounded-[30px] border border-primary/10 bg-gradient-to-br from-surface via-surface to-primary-soft p-5 shadow-sm md:p-7'>
        <div className='flex flex-col justify-between gap-6 lg:flex-row lg:items-end'>
          <div className='max-w-2xl'><div className='mb-3 flex items-center gap-2 text-xs font-semibold uppercase tracking-[.14em] text-primary'><Sparkles aria-hidden='true' className='size-4' /> Operations pulse</div><h2 className='text-3xl font-bold tracking-[-.045em] text-foreground md:text-4xl'>Mọi thứ cần xử lý, trong một nhịp nhìn.</h2><p className='mt-3 max-w-xl text-sm leading-6 text-muted md:text-base'>FlashX đang có <strong className='text-foreground'>{metrics.trips_active || 0}</strong> công việc hoạt động, <strong className='text-foreground'>{metrics.trips_searching || 0}</strong> đang tìm tài xế và <strong className='text-foreground'>{pendingDrivers.length}</strong> hồ sơ chờ duyệt.</p></div>
          <div className='flex items-center gap-3'><div className='fx-pulse-dot size-2 rounded-full bg-primary' /><div><div className='text-xs font-semibold text-primary'>Hệ thống đang đồng bộ</div><div className='text-xs text-muted'>Làm mới mỗi 5 giây</div></div></div>
        </div>
      </section>

      <div className='grid gap-4 sm:grid-cols-2 xl:grid-cols-4'>{metricsCards.map(({label,value,foot,icon:Icon}) => <Card key={label} className='transition-shadow duration-200 hover:shadow-md'><CardContent className='p-5 md:p-5'><div className='flex items-start justify-between gap-4'><div><p className='text-sm font-medium text-muted'>{label}</p>{initialLoading ? <div className='mt-3 h-9 w-20 animate-pulse rounded-lg bg-surface-soft motion-reduce:animate-none' /> : <p className='mt-2 text-3xl font-bold tracking-[-.04em] text-foreground'>{value}</p>}<p className='mt-1 text-xs text-muted'>{initialLoading ? 'Đang đồng bộ…' : foot}</p></div><div className='grid size-11 place-items-center rounded-2xl bg-primary-soft text-primary'><Icon aria-hidden='true' className='size-5' /></div></div></CardContent></Card>)}</div>

      <div className='grid gap-5 xl:grid-cols-[1.4fr_.6fr]'>
        <Card><CardHeader className='flex-row items-start justify-between'><div><CardTitle>Công việc gần đây</CardTitle><CardDescription>Dữ liệu trực tiếp từ backend.</CardDescription></div><Button size='sm' variant='ghost' onClick={() => selectView('trips')}>Xem tất cả <ChevronRight aria-hidden='true' className='size-4' /></Button></CardHeader><CardContent><DataTable columns={tripColumns} data={trips.slice(0,8)} loading={initialLoading} /></CardContent></Card>
        <Card className='overflow-hidden'><CardHeader><div className='flex items-center justify-between'><div><CardTitle>Cần xử lý</CardTitle><CardDescription>Sự cố trước, hồ sơ chờ duyệt sau.</CardDescription></div><Badge variant={incidents.length || pendingDrivers.length ? 'warning':'success'}>{incidents.length + pendingDrivers.length}</Badge></div></CardHeader><CardContent className='flex flex-col gap-3'>{incidents.length ? <button type='button' onClick={() => selectView('incidents')} className='flex min-h-14 items-center justify-between gap-3 rounded-2xl border border-danger/10 bg-danger-soft/60 p-4 text-left transition-colors hover:bg-danger-soft'><div className='flex items-center gap-3'><div className='grid size-10 place-items-center rounded-2xl bg-danger-soft text-danger'><AlertTriangle aria-hidden='true' className='size-5' /></div><div><div className='font-semibold text-foreground'>{incidents.length} sự cố đang mở</div><div className='mt-1 text-xs text-muted'>Ưu tiên kiểm tra công việc bất thường.</div></div></div><ChevronRight aria-hidden='true' className='size-4 text-danger' /></button> : null}{pendingDrivers.slice(0,3).map(driver => <div key={driver.id} className='rounded-2xl border border-border bg-surface-soft/60 p-4'><div className='flex items-start justify-between gap-3'><div><div className='font-semibold text-foreground'>{driver.full_name || driver.phone || driver.id}</div><div className='mt-1 text-xs text-muted'>{serviceLabel(driver.service_type)}</div></div><Badge variant='warning'>Chờ duyệt</Badge></div><div className='mt-4 flex gap-2'><Button size='sm' variant='destructive' disabled={busy} onClick={() => void approve(driver,'rejected')}>Từ chối</Button><Button size='sm' disabled={busy} onClick={() => void approve(driver,'approved')}>Duyệt</Button></div></div>)}{!incidents.length && !pendingDrivers.length ? <InfoState icon={ShieldCheck} title='Đã xử lý hết' text='Không có sự cố hoặc tài xế nào đang chờ duyệt.' /> : null}</CardContent></Card>
      </div>
    </div>;
  }

  return <div className='min-h-screen lg:grid lg:grid-cols-[264px_1fr]'>
    <aside className='hidden border-r border-border bg-surface/80 p-4 backdrop-blur-xl lg:flex lg:min-h-screen lg:flex-col'><Brand /><nav className='mt-7 flex flex-col gap-1'>{navItems.map(({key,label,icon:Icon}) => <button key={key} onClick={() => selectView(key)} className={`flex min-h-11 items-center gap-3 rounded-2xl px-3.5 text-left text-sm font-medium transition ${activeView===key?'bg-primary-soft text-primary':'text-muted hover:bg-surface-soft hover:text-foreground'}`}><Icon aria-hidden='true' className='size-[18px]' /><span>{label}</span>{activeView===key ? <span className='ml-auto size-1.5 rounded-full bg-primary' /> : null}</button>)}</nav><div className='mt-auto rounded-2xl border border-primary/10 bg-primary-soft/55 p-4'><div className='flex items-center gap-2 text-sm font-semibold text-primary'><ShieldCheck aria-hidden='true' className='size-4' /> FlashX Operations</div><p className='mt-2 text-xs leading-5 text-muted'>Ba dịch vụ MVP đang dùng chung một trung tâm vận hành.</p></div></aside>

    {mobileNavOpen ? <div className='fixed inset-0 z-50 bg-foreground/20 p-3 backdrop-blur-sm lg:hidden' onClick={() => setMobileNavOpen(false)}><div role='dialog' aria-modal='true' aria-label='Điều hướng quản trị' className='fx-glass ml-auto flex h-full w-[min(88vw,340px)] flex-col rounded-[28px] p-4' onClick={event => event.stopPropagation()}><div className='flex items-center justify-between'><Brand /><Button size='icon' variant='ghost' onClick={() => setMobileNavOpen(false)} aria-label='Đóng menu'><X aria-hidden='true' className='size-5' /></Button></div><nav className='mt-6 flex flex-col gap-1'>{navItems.map(({key,label,icon:Icon}) => <button key={key} onClick={() => selectView(key)} className={`flex min-h-12 items-center gap-3 rounded-2xl px-4 text-left text-sm font-semibold ${activeView===key?'bg-primary text-primary-foreground':'text-foreground hover:bg-surface-soft'}`}><Icon aria-hidden='true' className='size-5' />{label}</button>)}</nav></div></div> : null}

    <main className='min-w-0'><header className='sticky top-0 z-30 border-b border-border/80 bg-surface/80 backdrop-blur-xl'><div className='flex h-16 items-center justify-between px-4 md:px-6'><div className='flex items-center gap-3'><Button size='icon' variant='ghost' className='lg:hidden' onClick={() => setMobileNavOpen(true)} aria-label='Mở menu'><Menu aria-hidden='true' className='size-5' /></Button><div><div className='text-[11px] font-semibold uppercase tracking-[.13em] text-primary'>FlashX Operations</div><h1 className='text-lg font-semibold tracking-tight text-foreground'>{activeLabel}</h1></div></div><div className='flex items-center gap-2'><Badge variant={error ? 'destructive' : busy ? 'secondary' : 'success'}><span className={`mr-1.5 size-1.5 rounded-full ${error ? 'bg-danger' : busy ? 'bg-muted-soft' : 'bg-success'}`} />{error ? 'Cần kiểm tra' : busy ? 'Đang đồng bộ' : 'Đã đồng bộ'}</Badge><Button size='icon' variant='ghost' onClick={() => void load()} disabled={busy} aria-label='Làm mới'><RefreshCw aria-hidden='true' className={`size-4 ${busy?'animate-spin motion-reduce:animate-none':''}`} /></Button><Button size='icon' variant='ghost' onClick={() => void logout()} aria-label='Đăng xuất'><LogOut aria-hidden='true' className='size-4' /></Button></div></div></header><div className='mx-auto max-w-[1500px] p-4 md:p-6 lg:p-7'>{error ? <div role='alert' className='mb-4 rounded-2xl border border-danger/15 bg-danger-soft p-4 text-sm text-danger'>{error}</div> : null}{renderView()}</div></main>
  </div>;
}

function Brand() { return <div className='flex items-center gap-3 px-2'><div className='grid size-10 place-items-center rounded-2xl bg-primary text-primary-foreground shadow-sm'><Zap aria-hidden='true' className='size-5 fill-current' /></div><div><div className='text-base font-bold tracking-tight text-foreground'>FlashX</div><div className='text-[11px] font-medium text-muted'>Operations</div></div></div>; }

function LoginScreen({ challenge, debugCode, busy, error, phoneForm, otpForm, requestOtp, verifyOtp, resetChallenge }: any) {
  return <main className='relative grid min-h-screen place-items-center overflow-hidden bg-background p-5'><div className='absolute left-[-10%] top-[-20%] size-[38rem] rounded-full bg-primary-tint/50 blur-3xl' /><div className='absolute bottom-[-20%] right-[-12%] size-[34rem] rounded-full bg-primary-soft blur-3xl' /><Card className='fx-glass fx-enter relative z-10 w-full max-w-[430px] rounded-[30px]'><CardHeader className='pb-4'><Brand /><div className='pt-5'><CardTitle className='text-2xl font-bold tracking-[-.035em]'>Đăng nhập quản trị</CardTitle><CardDescription>Trung tâm vận hành FlashX · Xác thực bằng OTP.</CardDescription></div></CardHeader><CardContent className='flex flex-col gap-4'>{error ? <div role='alert' className='rounded-2xl bg-danger-soft p-3.5 text-sm text-danger'>{error}</div> : null}{!challenge ? <form onSubmit={requestOtp} className='flex flex-col gap-4'><label className='flex flex-col gap-1.5 text-sm font-semibold text-foreground'>Số điện thoại<Input {...phoneForm.register('phone')} disabled={busy} placeholder='0999999999' autoComplete='tel' /><span className='min-h-4 text-xs font-normal text-danger'>{phoneForm.formState.errors.phone?.message || ''}</span></label><Button size='lg' className='w-full' type='submit' disabled={busy}>{busy?'Đang xử lý…':'Tiếp tục'}</Button></form> : <form onSubmit={verifyOtp} className='flex flex-col gap-4'><label className='flex flex-col gap-1.5 text-sm font-semibold text-foreground'>Mã OTP<Input {...otpForm.register('code')} disabled={busy} inputMode='numeric' autoComplete='one-time-code' maxLength={6} placeholder='000000' /><span className='min-h-4 text-xs font-normal text-danger'>{otpForm.formState.errors.code?.message || ''}</span></label>{debugCode ? <div className='rounded-2xl border border-primary/10 bg-primary-soft p-3.5 text-sm text-primary'>DEV OTP: <strong className='font-mono tracking-wider'>{debugCode}</strong></div> : null}<div className='flex gap-2'><Button type='button' variant='outline' onClick={resetChallenge}>Quay lại</Button><Button className='flex-1' type='submit' disabled={busy}>Xác nhận OTP</Button></div></form>}<p className='text-center text-xs leading-5 text-muted'>Phiên đăng nhập được bảo vệ bằng access token ngắn hạn và refresh token.</p></CardContent></Card></main>;
}

function Section({ title, description, children }:{ title:string; description:string; children:ReactNode }) { return <Card className='fx-enter'><CardHeader><CardTitle className='text-xl'>{title}</CardTitle><CardDescription>{description}</CardDescription></CardHeader><CardContent>{children}</CardContent></Card>; }
function InfoState({ icon:Icon, title, text }:{ icon:LucideIcon; title:string; text:string }) { return <div className='rounded-3xl border border-dashed border-border bg-surface-soft/60 p-7 text-center'><div className='mx-auto grid size-12 place-items-center rounded-2xl bg-primary-soft text-primary'><Icon aria-hidden='true' className='size-5' /></div><div className='mt-4 font-semibold text-foreground'>{title}</div><p className='mx-auto mt-1 max-w-md text-sm leading-6 text-muted'>{text}</p></div>; }
