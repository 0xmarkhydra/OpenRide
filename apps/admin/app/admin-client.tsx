'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ColumnDef } from '@tanstack/react-table';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { AlertTriangle, BadgeDollarSign, CarFront, ClipboardList, Gauge, LayoutDashboard, ListChecks, LogOut, RefreshCw, SearchCheck, Settings, ShieldCheck, UserRound, UsersRound } from 'lucide-react';
import { DataTable } from '../components/data-table';
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

class ApiHttpError extends Error { constructor(public status: number, message: string) { super(message); } }

async function api<T>(path: string, options: RequestInit = {}, token?: string): Promise<T> {
  const headers = new Headers(options.headers); headers.set('content-type', 'application/json');
  if (token) headers.set('authorization', `Bearer ${token}`);
  const response = await fetch(`/api/flashx${path}`, { ...options, headers, cache: 'no-store' });
  const payload = (await response.json()) as ApiError & { data?: T };
  if (!response.ok) throw new ApiHttpError(response.status, payload.error?.message || `HTTP ${response.status}`);
  return payload.data as T;
}

function money(value?: number) { return new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND', maximumFractionDigits: 0 }).format(value || 0); }
function serviceLabel(service: string) { return ({ car: 'Lái hộ ô tô', designated_driver_car: 'Lái hộ ô tô', bike: 'Lái hộ xe máy', designated_driver_bike: 'Lái hộ xe máy', vehicle_inspection_assist: 'Đăng kiểm hộ' } as Record<string, string>)[service] || service; }
function tripStatusLabel(status: string) { return ({ scheduled:'Đã hẹn lịch', searching:'Đang tìm tài xế', accepted:'Đã nhận việc', arriving:'Đang đến nhận xe', arrived:'Đã tới điểm nhận', arriving_for_pickup:'Đang đến nhận xe', arrived_for_pickup:'Đã tới điểm nhận', vehicle_received:'Đã nhận xe khách', in_progress:'Đang thực hiện', en_route_to_inspection:'Đang tới nơi đăng kiểm', arrived_at_inspection_center:'Đã tới nơi đăng kiểm', inspection_in_progress:'Đang đăng kiểm', inspection_completed:'Đã có kết quả đăng kiểm', returning_vehicle:'Đang trả xe', arrived_for_return:'Đã tới điểm trả', handover:'Đang bàn giao', completed:'Hoàn thành', cancelled:'Đã hủy' } as Record<string,string>)[status] || status; }
function statusVariant(status: string, incident?: boolean): 'success' | 'warning' | 'destructive' | 'secondary' { if (incident || status === 'cancelled') return 'destructive'; if (status === 'completed') return 'success'; if (status === 'searching' || status === 'scheduled') return 'warning'; return 'secondary'; }

const phoneSchema = z.object({ phone: z.string().trim().min(9, 'Nhập số điện thoại Admin') });
const otpSchema = z.object({ code: z.string().regex(/^\d{6}$/, 'OTP phải gồm 6 chữ số') });
type PhoneForm = z.infer<typeof phoneSchema>; type OTPForm = z.infer<typeof otpSchema>;

const navItems: { key: ViewKey; label: string; icon: typeof LayoutDashboard }[] = [
  { key:'overview', label:'Tổng quan', icon:LayoutDashboard }, { key:'trips', label:'Công việc', icon:ClipboardList }, { key:'drivers', label:'Tài xế', icon:UserRound }, { key:'customers', label:'Khách hàng', icon:UsersRound }, { key:'vehicles', label:'Xe khách', icon:CarFront }, { key:'pricing', label:'Bảng giá', icon:BadgeDollarSign }, { key:'incidents', label:'Sự cố', icon:AlertTriangle }, { key:'audit', label:'Nhật ký', icon:ListChecks }, { key:'settings', label:'Cài đặt', icon:Settings },
];

export default function AdminClient() {
  const [token, setToken] = useState<string | null>(null); const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [challenge, setChallenge] = useState(''); const [debugCode, setDebugCode] = useState('');
  const [metrics, setMetrics] = useState<Metrics>({}); const [drivers, setDrivers] = useState<Driver[]>([]); const [trips, setTrips] = useState<Trip[]>([]);
  const [busy, setBusy] = useState(false); const [error, setError] = useState(''); const [activeView, setActiveView] = useState<ViewKey>('overview');
  const phoneForm = useForm<PhoneForm>({ resolver: zodResolver(phoneSchema), defaultValues: { phone: process.env.NEXT_PUBLIC_ADMIN_PHONE || '' } });
  const otpForm = useForm<OTPForm>({ resolver: zodResolver(otpSchema), defaultValues: { code: '' } });

  useEffect(() => { setToken(localStorage.getItem('flashx_admin_access_token')); setRefreshToken(localStorage.getItem('flashx_admin_refresh_token')); }, []);
  const clearSession = useCallback(() => { localStorage.removeItem('flashx_admin_access_token'); localStorage.removeItem('flashx_admin_refresh_token'); setToken(null); setRefreshToken(null); }, []);
  const refreshSession = useCallback(async () => { const current = refreshToken || localStorage.getItem('flashx_admin_refresh_token'); if (!current) return null; try { const tokens = await api<TokenPair>('/v1/auth/refresh', { method:'POST', body:JSON.stringify({ refresh_token:current }) }); localStorage.setItem('flashx_admin_access_token', tokens.access_token); localStorage.setItem('flashx_admin_refresh_token', tokens.refresh_token); setToken(tokens.access_token); setRefreshToken(tokens.refresh_token); return tokens.access_token; } catch { clearSession(); return null; } }, [clearSession, refreshToken]);
  const authedApi = useCallback(async <T,>(path:string, options:RequestInit={}, access=token):Promise<T> => { if (!access) throw new ApiHttpError(401,'Phiên đăng nhập đã hết hạn'); try { return await api<T>(path, options, access); } catch (e) { if (!(e instanceof ApiHttpError) || e.status !== 401) throw e; const refreshed = await refreshSession(); if (!refreshed) throw e; return api<T>(path, options, refreshed); } }, [refreshSession, token]);
  const load = useCallback(async (access=token) => { if (!access) return; setBusy(true); try { const [dashboard, driverData, tripData] = await Promise.all([authedApi<Metrics>('/v1/admin/dashboard',{},access), authedApi<Driver[]>('/v1/admin/drivers',{},access), authedApi<Trip[]>('/v1/admin/trips',{},access)]); setMetrics(dashboard); setDrivers(driverData); setTrips(tripData); setError(''); } catch (e) { setError(e instanceof Error ? e.message : 'Không thể tải dữ liệu'); } finally { setBusy(false); } }, [authedApi, token]);
  useEffect(() => { if (!token) return; void load(token); const timer = window.setInterval(() => void load(token), 5000); return () => window.clearInterval(timer); }, [token, load]);

  const requestOtp = phoneForm.handleSubmit(async ({ phone }) => { setBusy(true); setError(''); try { const result = await api<{ challenge_id:string; debug_code?:string }>('/v1/auth/otp/request',{ method:'POST', body:JSON.stringify({ phone, role:'admin' }) }); setChallenge(result.challenge_id); setDebugCode(result.debug_code || ''); } catch (e) { setError(e instanceof Error ? e.message : 'Không thể gửi OTP'); } finally { setBusy(false); } });
  const verifyOtp = otpForm.handleSubmit(async ({ code }) => { setBusy(true); setError(''); try { const result = await api<{ tokens:TokenPair }>('/v1/auth/otp/verify',{ method:'POST', body:JSON.stringify({ challenge_id:challenge, role:'admin', code }) }); localStorage.setItem('flashx_admin_access_token', result.tokens.access_token); localStorage.setItem('flashx_admin_refresh_token', result.tokens.refresh_token); setToken(result.tokens.access_token); setRefreshToken(result.tokens.refresh_token); setChallenge(''); setDebugCode(''); otpForm.reset(); } catch (e) { setError(e instanceof Error ? e.message : 'OTP không hợp lệ'); } finally { setBusy(false); } });
  async function approve(driver:Driver,status:'approved'|'rejected') { if (!token) return; setBusy(true); try { await authedApi(`/v1/admin/drivers/${driver.id}/approval`,{ method:'POST', body:JSON.stringify({ status, reason:status==='rejected'?'Từ chối bởi vận hành':'' }) },token); await load(); } catch (e) { setError(e instanceof Error ? e.message : 'Không thể cập nhật tài xế'); } finally { setBusy(false); } }
  async function logout() { if (refreshToken) { try { await api('/v1/auth/logout',{ method:'POST', body:JSON.stringify({ refresh_token:refreshToken }) }); } catch {} } clearSession(); }

  const pendingDrivers = useMemo(() => drivers.filter(driver => driver.approval_status === 'pending'), [drivers]);
  const incidents = useMemo(() => trips.filter(trip => trip.incident_open), [trips]);
  const customers = useMemo<SimpleRow[]>(() => Array.from(new Set(trips.map(t => t.rider_id))).filter(Boolean).map(id => ({ id, subtitle:`${trips.filter(t => t.rider_id === id).length} công việc` })), [trips]);
  const vehicles = useMemo<SimpleRow[]>(() => Array.from(new Set(trips.map(t => t.customer_vehicle_id).filter(Boolean) as string[])).map(id => ({ id, subtitle:`${trips.filter(t => t.customer_vehicle_id === id).length} công việc` })), [trips]);

  const tripColumns = useMemo<ColumnDef<Trip>[]>(() => [
    { accessorKey:'id', header:'Mã việc', cell:({row}) => <span className='font-mono text-xs font-semibold'>{row.original.id}</span> },
    { accessorKey:'service_type', header:'Dịch vụ', cell:({row}) => serviceLabel(row.original.service_type) },
    { accessorKey:'rider_id', header:'Khách' }, { accessorKey:'driver_id', header:'Tài xế', cell:({row}) => row.original.driver_id || '—' },
    { id:'fare', header:'Giá', cell:({row}) => <strong>{money(row.original.final_fare_minor || row.original.estimated_fare_minor)}</strong> },
    { accessorKey:'status', header:'Trạng thái', cell:({row}) => <Badge variant={statusVariant(row.original.status,row.original.incident_open)}>{row.original.incident_open?'Có sự cố':tripStatusLabel(row.original.status)}</Badge> },
  ], []);
  const driverColumns = useMemo<ColumnDef<Driver>[]>(() => [
    { accessorKey:'full_name', header:'Tài xế', cell:({row}) => <div><div className='font-semibold text-slate-950'>{row.original.full_name || row.original.phone || row.original.id}</div><div className='text-xs text-slate-500'>{row.original.phone || row.original.id}</div></div> },
    { accessorKey:'service_type', header:'Dịch vụ', cell:({row}) => serviceLabel(row.original.service_type) },
    { accessorKey:'availability_status', header:'Trạng thái', cell:({row}) => <Badge variant={row.original.availability_status === 'online' ? 'success':'secondary'}>{row.original.availability_status || 'offline'}</Badge> },
    { accessorKey:'approval_status', header:'Duyệt', cell:({row}) => <Badge variant={row.original.approval_status === 'approved'?'success':row.original.approval_status === 'rejected'?'destructive':'warning'}>{row.original.approval_status}</Badge> },
    { id:'actions', header:'Thao tác', cell:({row}) => row.original.approval_status === 'pending' ? <div className='flex gap-2'><Button size='sm' variant='destructive' disabled={busy} onClick={() => void approve(row.original,'rejected')}>Từ chối</Button><Button size='sm' variant='accent' disabled={busy} onClick={() => void approve(row.original,'approved')}>Duyệt</Button></div> : null },
  ], [busy]);
  const simpleColumns = useMemo<ColumnDef<SimpleRow>[]>(() => [{ accessorKey:'id', header:'Mã', cell:({row}) => <span className='font-mono text-xs font-semibold'>{row.original.id}</span> }, { accessorKey:'subtitle', header:'Hoạt động' }], []);

  if (!token) return <main className='grid min-h-screen place-items-center bg-slate-950 p-5'><Card className='w-full max-w-md border-white/10 shadow-2xl'><CardHeader><div className='mb-4 flex size-11 items-center justify-center rounded-xl bg-yellow-400 font-black text-slate-950'>⚡</div><CardTitle className='text-2xl'>Đăng nhập quản trị</CardTitle><CardDescription>FlashX Operations · OTP dành cho Super Admin.</CardDescription></CardHeader><CardContent>{error && <div className='mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-700'>{error}</div>}{!challenge ? <form onSubmit={requestOtp} className='space-y-4'><div><label className='mb-1.5 block text-sm font-semibold'>Số điện thoại</label><Input {...phoneForm.register('phone')} disabled={busy} placeholder='0999999999' />{phoneForm.formState.errors.phone && <p className='mt-1 text-xs text-red-600'>{phoneForm.formState.errors.phone.message}</p>}</div><Button className='w-full' type='submit' disabled={busy}>{busy?'Đang xử lý…':'Gửi OTP'}</Button></form> : <form onSubmit={verifyOtp} className='space-y-4'><div><label className='mb-1.5 block text-sm font-semibold'>Mã OTP</label><Input {...otpForm.register('code')} inputMode='numeric' maxLength={6} placeholder='000000' />{otpForm.formState.errors.code && <p className='mt-1 text-xs text-red-600'>{otpForm.formState.errors.code.message}</p>}</div>{debugCode && <div className='rounded-lg bg-amber-50 p-3 text-sm text-amber-800'>DEV OTP: <strong>{debugCode}</strong></div>}<div className='flex gap-2'><Button type='button' variant='outline' onClick={() => { setChallenge(''); setDebugCode(''); }}>Quay lại</Button><Button className='flex-1' type='submit' disabled={busy}>Xác nhận OTP</Button></div></form>}</CardContent></Card></main>;

  const activeLabel = navItems.find(item => item.key === activeView)?.label || 'Tổng quan';
  const metricsCards = [
    { label:'Tổng công việc', value:metrics.trips_total || 0, foot:`${metrics.trips_completed || 0} hoàn thành`, icon:ClipboardList },
    { label:'Tài xế online', value:metrics.drivers_online || 0, foot:`${metrics.drivers_total || 0} tổng tài xế`, icon:UserRound },
    { label:'Đang vận hành', value:metrics.trips_active || 0, foot:`${metrics.trips_searching || 0} đang tìm`, icon:Gauge },
    { label:'Cần chú ý', value:metrics.trips_incident || 0, foot:`${metrics.drivers_pending || 0} tài xế chờ duyệt`, icon:AlertTriangle },
  ];

  function renderView() {
    if (activeView === 'trips') return <Section title='Công việc' description='Theo dõi toàn bộ vòng đời của ba dịch vụ.'><DataTable columns={tripColumns} data={trips} searchPlaceholder='Tìm mã việc, khách, tài xế...' /></Section>;
    if (activeView === 'drivers') return <Section title='Tài xế' description='Duyệt hồ sơ và theo dõi trạng thái cung ứng.'><DataTable columns={driverColumns} data={drivers} searchPlaceholder='Tìm tài xế...' /></Section>;
    if (activeView === 'customers') return <Section title='Khách hàng' description='Khách hàng phát sinh từ dữ liệu công việc hiện tại.'><DataTable columns={simpleColumns} data={customers} searchPlaceholder='Tìm khách hàng...' /></Section>;
    if (activeView === 'vehicles') return <Section title='Xe khách' description='Các xe đã được gắn với công việc FlashX.'><DataTable columns={simpleColumns} data={vehicles} searchPlaceholder='Tìm xe...' /></Section>;
    if (activeView === 'incidents') return <Section title='Sự cố' description='Các công việc được đánh dấu cần Operations xử lý.'><DataTable columns={tripColumns} data={incidents} searchPlaceholder='Tìm sự cố...' /></Section>;
    if (activeView === 'pricing') return <Section title='Bảng giá' description='Giá hiện do backend tính theo dịch vụ và hành trình.'><InfoState icon={BadgeDollarSign} title='Chưa mở API quản trị bảng giá' text='Màn hình đã sẵn sàng. Bước tiếp theo là đưa cấu hình pricing version và components lên Admin thay vì hard-code trên frontend.' /></Section>;
    if (activeView === 'audit') return <Section title='Nhật ký' description='Theo dõi các thao tác quản trị quan trọng.'><InfoState icon={SearchCheck} title='Backend đã ghi audit' text='API đọc audit log chưa expose cho Admin. Các thay đổi duyệt tài xế hiện đã được ghi audit ở backend.' /></Section>;
    if (activeView === 'settings') return <Section title='Cài đặt' description='Cấu hình môi trường vận hành FlashX.'><div className='grid gap-4 md:grid-cols-2'><InfoState icon={ShieldCheck} title='Môi trường dev' text='Admin xác thực bằng OTP development; dữ liệu lấy trực tiếp từ FlashX-BE.' /><InfoState icon={Settings} title='Cấu hình có kiểm soát' text='Các secret và provider config tiếp tục nằm trên Railway, không đưa xuống client.' /></div></Section>;
    return <div className='space-y-5'><div className='grid gap-4 sm:grid-cols-2 xl:grid-cols-4'>{metricsCards.map(({label,value,foot,icon:Icon}) => <Card key={label}><CardContent className='p-5'><div className='flex items-start justify-between'><div><p className='text-sm font-medium text-slate-500'>{label}</p><p className='mt-2 text-3xl font-bold tracking-tight'>{value}</p><p className='mt-1 text-xs text-slate-500'>{foot}</p></div><div className='rounded-xl bg-yellow-50 p-2.5 text-slate-950'><Icon className='size-5' /></div></div></CardContent></Card>)}</div><div className='grid gap-5 xl:grid-cols-[1.35fr_.65fr]'><Card><CardHeader><CardTitle>Công việc gần đây</CardTitle><CardDescription>Dữ liệu realtime từ backend.</CardDescription></CardHeader><CardContent><DataTable columns={tripColumns} data={trips.slice(0,8)} /></CardContent></Card><Card><CardHeader><div className='flex items-center justify-between'><div><CardTitle>Chờ duyệt tài xế</CardTitle><CardDescription>Hồ sơ cần Operations xử lý.</CardDescription></div><Badge variant='warning'>{pendingDrivers.length}</Badge></div></CardHeader><CardContent className='space-y-3'>{pendingDrivers.length ? pendingDrivers.slice(0,6).map(driver => <div key={driver.id} className='rounded-xl border border-slate-100 p-3'><div className='font-semibold'>{driver.full_name || driver.phone || driver.id}</div><div className='mt-1 text-xs text-slate-500'>{serviceLabel(driver.service_type)}</div><div className='mt-3 flex gap-2'><Button size='sm' variant='destructive' onClick={() => void approve(driver,'rejected')}>Từ chối</Button><Button size='sm' variant='accent' onClick={() => void approve(driver,'approved')}>Duyệt</Button></div></div>) : <InfoState icon={ShieldCheck} title='Đã xử lý hết' text='Không có tài xế nào đang chờ duyệt.' />}</CardContent></Card></div></div>;
  }

  return <div className='min-h-screen bg-slate-50 text-slate-950 lg:grid lg:grid-cols-[260px_1fr]'><aside className='hidden min-h-screen bg-slate-950 p-4 text-white lg:block'><div className='mb-7 flex items-center gap-3 px-2 py-2'><div className='grid size-10 place-items-center rounded-xl bg-yellow-400 font-black text-slate-950'>⚡</div><div><div className='font-bold'>FlashX</div><div className='text-xs text-slate-400'>Operations</div></div></div><nav className='space-y-1'>{navItems.map(({key,label,icon:Icon}) => <button key={key} onClick={() => setActiveView(key)} className={`flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left text-sm font-medium transition ${activeView===key?'bg-yellow-400 text-slate-950':'text-slate-400 hover:bg-white/5 hover:text-white'}`}><Icon className='size-4' />{label}</button>)}</nav><div className='mt-8 rounded-xl border border-white/10 bg-white/5 p-3 text-xs text-slate-400'>MVP Operations<br/>Dữ liệu trực tiếp từ FlashX API.</div></aside><main className='min-w-0'><header className='sticky top-0 z-20 flex h-16 items-center justify-between border-b border-slate-200 bg-white/90 px-4 backdrop-blur md:px-6'><div><div className='text-xs font-semibold uppercase tracking-wider text-slate-400'>FlashX Operations</div><h1 className='text-lg font-bold'>{activeLabel}</h1></div><div className='flex items-center gap-2'><Badge variant='success'>LIVE</Badge><Button size='icon' variant='ghost' onClick={() => void load()} disabled={busy} title='Làm mới'><RefreshCw className={`size-4 ${busy?'animate-spin':''}`} /></Button><Button size='icon' variant='ghost' onClick={() => void logout()} title='Đăng xuất'><LogOut className='size-4' /></Button></div></header><div className='mx-auto max-w-[1500px] p-4 md:p-6'>{error && <div className='mb-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700'>{error}</div>}{renderView()}</div></main></div>;
}

function Section({ title, description, children }: { title:string; description:string; children:React.ReactNode }) { return <Card><CardHeader><CardTitle className='text-xl'>{title}</CardTitle><CardDescription>{description}</CardDescription></CardHeader><CardContent>{children}</CardContent></Card>; }
function InfoState({ icon:Icon, title, text }: { icon:typeof ShieldCheck; title:string; text:string }) { return <div className='rounded-xl border border-dashed border-slate-200 bg-slate-50 p-6 text-center'><Icon className='mx-auto size-6 text-slate-400' /><div className='mt-3 font-semibold text-slate-900'>{title}</div><p className='mx-auto mt-1 max-w-md text-sm leading-6 text-slate-500'>{text}</p></div>; }
