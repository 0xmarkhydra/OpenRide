'use client';

import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ColumnDef } from '@tanstack/react-table';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import {
  AlertTriangle,
  BadgeDollarSign,
  Ban,
  CarFront,
  CheckCircle2,
  ChevronRight,
  ClipboardList,
  Clock3,
  Database,
  ExternalLink,
  FileCheck2,
  Gauge,
  LayoutDashboard,
  ListChecks,
  LogOut,
  MapPin,
  Menu,
  RefreshCw,
  SearchCheck,
  Settings,
  ShieldCheck,
  Sparkles,
  UserCheck,
  UserRound,
  UsersRound,
  Wifi,
  X,
  Zap,
  type LucideIcon,
} from 'lucide-react';
import { DataTable, dataTableFeatures } from '../components/data-table';
import { DetailPanel } from '../components/detail-panel';
import { Badge } from '../components/ui/badge';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Input } from '../components/ui/input';

type TokenPair = { access_token: string; refresh_token: string; expires_at: string };
type AdminUser = {
  id: string;
  phone: string;
  email: string;
  display_name: string;
  role: 'super_admin' | 'operations';
  status: 'active' | 'disabled';
  created_at: string;
  updated_at: string;
};
type Driver = {
  id: string;
  phone?: string;
  full_name: string;
  service_type: string;
  capabilities?: string[];
  approval_status: string;
  availability_status: string;
  location?: { lat: number; lng: number; captured_at?: string };
  created_at?: string;
};
type DriverCandidate = { driver: Driver; distance_to_pickup_m: number };
type Trip = {
  id: string;
  rider_id: string;
  driver_id?: string;
  customer_vehicle_id?: string;
  service_type: string;
  booking_mode?: string;
  scheduled_at?: string;
  status: string;
  pickup?: { lat: number; lng: number };
  destination?: { lat: number; lng: number };
  estimated_distance_m?: number;
  estimated_duration_s?: number;
  estimated_fare_minor: number;
  final_fare_minor: number;
  inspection_result?: string;
  incident_type?: string;
  incident_note?: string;
  incident_open?: boolean;
  created_at: string;
};
type Customer = { id: string; phone: string; full_name: string; status: string; created_at: string; updated_at: string };
type Vehicle = {
  id: string;
  owner_user_id: string;
  type: string;
  license_plate: string;
  brand: string;
  model: string;
  year?: number;
  color: string;
  transmission: string;
  seats?: number;
  notes?: string;
  status: string;
  created_at: string;
};
type PricingRule = {
  service_type: string;
  base_fare_minor: number;
  per_km_minor: number;
  service_minor: number;
  minimum_minor: number;
  currency: string;
  pricing_version: string;
};
type AuditEntry = {
  actor_type: string;
  actor_id: string;
  action: string;
  resource_type: string;
  resource_id: string;
  metadata?: Record<string, unknown>;
  created_at: string;
};
type SystemInfo = {
  app_env: string;
  persistence: string;
  allow_dev_identity: boolean;
  supported_services: string[];
  features: Record<string, boolean>;
};
type DriverDocument = {
  id: string;
  driver_id: string;
  document_type: string;
  filename: string;
  content_type: string;
  size_bytes: number;
  review_status: string;
  review_note?: string;
  created_at: string;
};
type DriverDocumentWithURL = {
  document: DriverDocument;
  view: { method: string; url: string; expires_at: string };
};
type Metrics = Record<string, number>;
type ApiError = { error?: { message?: string; code?: string } };
type ViewKey = 'overview' | 'trips' | 'drivers' | 'customers' | 'vehicles' | 'pricing' | 'incidents' | 'audit' | 'accounts' | 'settings';
type NavItem = { key: ViewKey; label: string; icon: LucideIcon; superOnly?: boolean };

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
function formatDate(value?: string) {
  if (!value) return '—';
  return new Intl.DateTimeFormat('vi-VN', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(value));
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
function tripStatusLabel(status: string) {
  return ({
    scheduled: 'Đã hẹn lịch', searching: 'Đang tìm tài xế', accepted: 'Đã nhận việc', arriving: 'Đang đến nhận xe', arrived: 'Đã tới điểm nhận',
    arriving_for_pickup: 'Đang đến nhận xe', arrived_for_pickup: 'Đã tới điểm nhận', vehicle_received: 'Đã nhận xe khách', in_progress: 'Đang thực hiện',
    en_route_to_inspection: 'Đang tới nơi đăng kiểm', arrived_at_inspection_center: 'Đã tới nơi đăng kiểm', inspection_in_progress: 'Đang đăng kiểm',
    inspection_completed: 'Đã có kết quả đăng kiểm', returning_vehicle: 'Đang trả xe', arrived_for_return: 'Đã tới điểm trả', handover: 'Đang bàn giao',
    completed: 'Hoàn thành', cancelled: 'Đã hủy',
  } as Record<string, string>)[status] || status;
}
function statusVariant(status: string, incident?: boolean): 'success' | 'warning' | 'destructive' | 'secondary' {
  if (incident || status === 'cancelled') return 'destructive';
  if (status === 'completed') return 'success';
  if (status === 'searching' || status === 'scheduled') return 'warning';
  return 'secondary';
}
function approvalLabel(status: string) {
  return ({ pending: 'Chờ duyệt', approved: 'Đã duyệt', rejected: 'Từ chối', suspended: 'Tạm khóa' } as Record<string, string>)[status] || status;
}
function adminRoleLabel(role: string) { return role === 'super_admin' ? 'Super Admin' : role === 'operations' ? 'Operations' : role; }
function approvalVariant(status: string): 'success' | 'warning' | 'destructive' | 'secondary' {
  if (status === 'approved') return 'success';
  if (status === 'rejected' || status === 'suspended') return 'destructive';
  return status === 'pending' ? 'warning' : 'secondary';
}
function documentLabel(value: string) {
  return ({
    identity_front: 'CCCD mặt trước', identity_back: 'CCCD mặt sau', driver_license: 'Giấy phép lái xe',
    vehicle_registration: 'Đăng ký xe', vehicle_insurance: 'Bảo hiểm xe', portrait: 'Ảnh chân dung',
  } as Record<string, string>)[value] || value;
}
function vehicleTypeLabel(value: string) { return value === 'motorbike' ? 'Xe máy' : value === 'car' ? 'Ô tô' : value; }
function km(value?: number) { return value ? `${(value / 1000).toLocaleString('vi-VN', { maximumFractionDigits: 1 })} km` : '—'; }
function duration(value?: number) { return value ? `${Math.max(1, Math.round(value / 60))} phút` : '—'; }
function coordinate(point?: { lat: number; lng: number }) { return point ? `${point.lat.toFixed(5)}, ${point.lng.toFixed(5)}` : '—'; }
function canManageDriver(trip: Trip) {
  return !trip.incident_open && ['searching', 'accepted', 'arriving', 'arrived', 'arriving_for_pickup', 'arrived_for_pickup'].includes(trip.status);
}

const phoneSchema = z.object({ phone: z.string().trim().min(9, 'Nhập số điện thoại Admin') });
const otpSchema = z.object({ code: z.string().regex(/^\d{6}$/, 'OTP phải gồm 6 chữ số') });
const pricingSchema = z.object({
  base_fare_minor: z.number().min(0, 'Không được âm').max(100_000_000, 'Giá trị quá lớn'),
  per_km_minor: z.number().min(0, 'Không được âm').max(100_000_000, 'Giá trị quá lớn'),
  service_minor: z.number().min(0, 'Không được âm').max(100_000_000, 'Giá trị quá lớn'),
  minimum_minor: z.number().min(0, 'Không được âm').max(100_000_000, 'Giá trị quá lớn'),
}).refine(values => Object.values(values).some(value => value > 0), { message: 'Ít nhất một thành phần giá phải lớn hơn 0', path: ['minimum_minor'] });
const adminAccountSchema = z.object({
  phone: z.string().trim().min(9, 'Nhập số điện thoại quản trị viên'),
  display_name: z.string().trim().min(2, 'Tên hiển thị tối thiểu 2 ký tự'),
  role: z.enum(['operations', 'super_admin']),
});
type PhoneForm = z.infer<typeof phoneSchema>;
type OTPForm = z.infer<typeof otpSchema>;
type PricingForm = z.infer<typeof pricingSchema>;
type AdminAccountForm = z.infer<typeof adminAccountSchema>;

const navItems: NavItem[] = [
  { key: 'overview', label: 'Tổng quan', icon: LayoutDashboard },
  { key: 'trips', label: 'Công việc', icon: ClipboardList },
  { key: 'drivers', label: 'Tài xế', icon: UserRound },
  { key: 'customers', label: 'Khách hàng', icon: UsersRound },
  { key: 'vehicles', label: 'Xe khách', icon: CarFront },
  { key: 'pricing', label: 'Bảng giá', icon: BadgeDollarSign },
  { key: 'incidents', label: 'Sự cố', icon: AlertTriangle },
  { key: 'audit', label: 'Nhật ký', icon: ListChecks },
  { key: 'accounts', label: 'Quản trị viên', icon: ShieldCheck, superOnly: true },
  { key: 'settings', label: 'Cài đặt', icon: Settings },
];

export default function AdminClient() {
  const [token, setToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  const [challenge, setChallenge] = useState('');
  const [debugCode, setDebugCode] = useState('');
  const [metrics, setMetrics] = useState<Metrics>({});
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [trips, setTrips] = useState<Trip[]>([]);
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [pricing, setPricing] = useState<PricingRule[]>([]);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [currentAdmin, setCurrentAdmin] = useState<AdminUser | null>(null);
  const [adminAccounts, setAdminAccounts] = useState<AdminUser[]>([]);
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [busy, setBusy] = useState(false);
  const [hasLoaded, setHasLoaded] = useState(false);
  const [error, setError] = useState('');
  const [activeView, setActiveView] = useState<ViewKey>('overview');
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [selectedTrip, setSelectedTrip] = useState<Trip | null>(null);
  const [selectedDriver, setSelectedDriver] = useState<Driver | null>(null);
  const [selectedCustomer, setSelectedCustomer] = useState<Customer | null>(null);
  const [selectedVehicle, setSelectedVehicle] = useState<Vehicle | null>(null);
  const [driverDocs, setDriverDocs] = useState<DriverDocumentWithURL[]>([]);
  const [driverDocsLoading, setDriverDocsLoading] = useState(false);
  const [driverCandidates, setDriverCandidates] = useState<DriverCandidate[]>([]);
  const [driverCandidatesLoading, setDriverCandidatesLoading] = useState(false);
  const [assignmentReason, setAssignmentReason] = useState('');
  const [driverReason, setDriverReason] = useState('');
  const [documentNote, setDocumentNote] = useState('');
  const [incidentResolutionNote, setIncidentResolutionNote] = useState('');

  const phoneForm = useForm<PhoneForm>({ resolver: zodResolver(phoneSchema), defaultValues: { phone: process.env.NEXT_PUBLIC_ADMIN_PHONE || '' } });
  const otpForm = useForm<OTPForm>({ resolver: zodResolver(otpSchema), defaultValues: { code: '' } });

  useEffect(() => {
    setToken(localStorage.getItem('flashx_admin_access_token'));
    setRefreshToken(localStorage.getItem('flashx_admin_refresh_token'));
  }, []);

  const clearSession = useCallback(() => {
    localStorage.removeItem('flashx_admin_access_token');
    localStorage.removeItem('flashx_admin_refresh_token');
    setToken(null);
    setRefreshToken(null);
    setCurrentAdmin(null);
    setAdminAccounts([]);
    setHasLoaded(false);
  }, []);

  const refreshSession = useCallback(async () => {
    const current = refreshToken || localStorage.getItem('flashx_admin_refresh_token');
    if (!current) return null;
    try {
      const tokens = await api<TokenPair>('/v1/auth/refresh', { method: 'POST', body: JSON.stringify({ refresh_token: current }) });
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
    } catch (requestError) {
      if (!(requestError instanceof ApiHttpError) || requestError.status !== 401) throw requestError;
      const refreshed = await refreshSession();
      if (!refreshed) throw requestError;
      return api<T>(path, options, refreshed);
    }
  }, [refreshSession, token]);

  const loadCore = useCallback(async (access = token) => {
    if (!access) return;
    setBusy(true);
    try {
      const [dashboard, driverData, tripData] = await Promise.all([
        authedApi<Metrics>('/v1/admin/dashboard', {}, access),
        authedApi<Driver[]>('/v1/admin/drivers', {}, access),
        authedApi<Trip[]>('/v1/admin/trips', {}, access),
      ]);
      setMetrics(dashboard);
      setDrivers(driverData);
      setTrips(tripData);
      setHasLoaded(true);
      setError('');
      setSelectedTrip(current => current ? tripData.find(item => item.id === current.id) || current : null);
      setSelectedDriver(current => current ? driverData.find(item => item.id === current.id) || current : null);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Không thể tải dữ liệu vận hành');
    } finally {
      setBusy(false);
    }
  }, [authedApi, token]);

  const loadReference = useCallback(async (access = token) => {
    if (!access) return;
    try {
      const [adminData, customerData, vehicleData, pricingData, auditData, systemData] = await Promise.all([
        authedApi<AdminUser>('/v1/admin/me', {}, access),
        authedApi<Customer[]>('/v1/admin/customers', {}, access),
        authedApi<Vehicle[]>('/v1/admin/vehicles', {}, access),
        authedApi<PricingRule[]>('/v1/admin/pricing', {}, access),
        authedApi<AuditEntry[]>('/v1/admin/audit?limit=150', {}, access),
        authedApi<SystemInfo>('/v1/admin/system', {}, access),
      ]);
      setCurrentAdmin(adminData);
      setCustomers(customerData);
      setVehicles(vehicleData);
      setPricing(pricingData);
      setAudit(auditData);
      setSystemInfo(systemData);
      if (adminData.role === 'super_admin') {
        setAdminAccounts(await authedApi<AdminUser[]>('/v1/admin/accounts?limit=200', {}, access));
      } else {
        setAdminAccounts([]);
        setActiveView(current => current === 'accounts' ? 'overview' : current);
      }
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Không thể tải dữ liệu quản trị');
    }
  }, [authedApi, token]);

  const loadAudit = useCallback(async () => {
    if (!token) return;
    const items = await authedApi<AuditEntry[]>('/v1/admin/audit?limit=150');
    setAudit(items);
  }, [authedApi, token]);

  const loadDriverCandidates = useCallback(async (trip: Trip | null) => {
    if (!trip || !canManageDriver(trip)) {
      setDriverCandidates([]);
      setDriverCandidatesLoading(false);
      return;
    }
    setDriverCandidatesLoading(true);
    try {
      setDriverCandidates(await authedApi<DriverCandidate[]>(`/v1/admin/trips/${trip.id}/driver-candidates?limit=15`));
    } catch (candidateError) {
      setDriverCandidates([]);
      if (candidateError instanceof ApiHttpError && candidateError.status === 409) return;
      setError(candidateError instanceof Error ? candidateError.message : 'Không thể tải tài xế phù hợp');
    } finally {
      setDriverCandidatesLoading(false);
    }
  }, [authedApi]);

  useEffect(() => {
    if (!token) return;
    void loadCore(token);
    void loadReference(token);
    const timer = window.setInterval(() => void loadCore(token), 5000);
    return () => window.clearInterval(timer);
  }, [token, loadCore, loadReference]);

  useEffect(() => {
    setAssignmentReason('');
    void loadDriverCandidates(selectedTrip);
  }, [selectedTrip?.id, selectedTrip?.status, selectedTrip?.driver_id, selectedTrip?.incident_open, loadDriverCandidates]);

  const requestOtp = phoneForm.handleSubmit(async ({ phone }) => {
    setBusy(true);
    setError('');
    try {
      const result = await api<{ challenge_id: string; debug_code?: string }>('/v1/auth/otp/request', { method: 'POST', body: JSON.stringify({ phone, role: 'admin' }) });
      setChallenge(result.challenge_id);
      setDebugCode(result.debug_code || '');
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Không thể gửi OTP');
    } finally {
      setBusy(false);
    }
  });

  const verifyOtp = otpForm.handleSubmit(async ({ code }) => {
    setBusy(true);
    setError('');
    try {
      const result = await api<{ tokens: TokenPair }>('/v1/auth/otp/verify', { method: 'POST', body: JSON.stringify({ challenge_id: challenge, role: 'admin', code }) });
      localStorage.setItem('flashx_admin_access_token', result.tokens.access_token);
      localStorage.setItem('flashx_admin_refresh_token', result.tokens.refresh_token);
      setToken(result.tokens.access_token);
      setRefreshToken(result.tokens.refresh_token);
      setChallenge('');
      setDebugCode('');
      otpForm.reset();
    } catch (verifyError) {
      setError(verifyError instanceof Error ? verifyError.message : 'OTP không hợp lệ');
    } finally {
      setBusy(false);
    }
  });

  async function setDriverApproval(driver: Driver, status: 'approved' | 'rejected' | 'suspended') {
    if (!token) return;
    const reason = status === 'approved' ? '' : driverReason.trim();
    if (status !== 'approved' && !reason) {
      setError('Nhập lý do trước khi từ chối hoặc tạm khóa tài xế.');
      return;
    }
    setBusy(true);
    setError('');
    try {
      const updated = await authedApi<Driver>(`/v1/admin/drivers/${driver.id}/approval`, { method: 'POST', body: JSON.stringify({ status, reason }) }, token);
      setDrivers(items => items.map(item => item.id === updated.id ? updated : item));
      setSelectedDriver(updated);
      setDriverReason('');
      await Promise.all([loadCore(), loadAudit()]);
    } catch (approvalError) {
      setError(approvalError instanceof Error ? approvalError.message : 'Không thể cập nhật tài xế');
    } finally {
      setBusy(false);
    }
  }

  async function openDriver(driver: Driver) {
    setSelectedDriver(driver);
    setDriverReason('');
    setDocumentNote('');
    setDriverDocsLoading(true);
    try {
      setDriverDocs(await authedApi<DriverDocumentWithURL[]>(`/v1/admin/drivers/${driver.id}/documents`));
    } catch (documentsError) {
      setDriverDocs([]);
      setError(documentsError instanceof Error ? documentsError.message : 'Không thể tải hồ sơ KYC');
    } finally {
      setDriverDocsLoading(false);
    }
  }

  async function reviewDocument(item: DriverDocumentWithURL, status: 'approved' | 'rejected') {
    if (!selectedDriver) return;
    setBusy(true);
    setError('');
    try {
      const updated = await authedApi<DriverDocument>(`/v1/admin/drivers/${selectedDriver.id}/documents/${item.document.id}/review`, {
        method: 'POST', body: JSON.stringify({ status, note: documentNote.trim() }),
      });
      setDriverDocs(items => items.map(current => current.document.id === updated.id ? { ...current, document: updated } : current));
      setDocumentNote('');
      await loadAudit();
    } catch (reviewError) {
      setError(reviewError instanceof Error ? reviewError.message : 'Không thể duyệt tài liệu');
    } finally {
      setBusy(false);
    }
  }

  async function resolveIncident(trip: Trip) {
    setBusy(true);
    setError('');
    try {
      const updated = await authedApi<Trip>(`/v1/admin/trips/${trip.id}/incident/resolve`, {
        method: 'POST', body: JSON.stringify({ note: incidentResolutionNote.trim() }),
      });
      setTrips(items => items.map(item => item.id === updated.id ? updated : item));
      setSelectedTrip(updated);
      setIncidentResolutionNote('');
      await Promise.all([loadCore(), loadAudit()]);
    } catch (incidentError) {
      setError(incidentError instanceof Error ? incidentError.message : 'Không thể đóng sự cố');
    } finally {
      setBusy(false);
    }
  }

  async function assignTripDriver(candidate: DriverCandidate) {
    if (!selectedTrip) return;
    const reason = assignmentReason.trim();
    if (selectedTrip.driver_id && !reason) {
      setError('Nhập lý do trước khi đổi tài xế cho công việc đang được nhận.');
      return;
    }
    setBusy(true);
    setError('');
    try {
      const updated = await authedApi<Trip>(`/v1/admin/trips/${selectedTrip.id}/assign-driver`, {
        method: 'POST', body: JSON.stringify({ driver_id: candidate.driver.id, reason }),
      });
      setTrips(items => items.map(item => item.id === updated.id ? updated : item));
      setSelectedTrip(updated);
      setAssignmentReason('');
      await Promise.all([loadCore(), loadAudit(), loadDriverCandidates(updated)]);
    } catch (assignError) {
      setError(assignError instanceof Error ? assignError.message : 'Không thể điều phối tài xế');
    } finally {
      setBusy(false);
    }
  }

  async function updatePricingRule(serviceType: string, values: PricingForm) {
    setBusy(true);
    setError('');
    try {
      const updated = await authedApi<PricingRule>(`/v1/admin/pricing/${encodeURIComponent(serviceType)}`, {
        method: 'PATCH', body: JSON.stringify(values),
      });
      setPricing(items => items.map(item => item.service_type === updated.service_type ? updated : item));
      await loadAudit();
    } catch (pricingError) {
      setError(pricingError instanceof Error ? pricingError.message : 'Không thể cập nhật bảng giá');
      throw pricingError;
    } finally {
      setBusy(false);
    }
  }

  async function createAdminAccount(values: AdminAccountForm) {
    setBusy(true);
    setError('');
    try {
      const created = await authedApi<AdminUser>('/v1/admin/accounts', { method: 'POST', body: JSON.stringify(values) });
      setAdminAccounts(items => [created, ...items]);
      await loadAudit();
    } catch (accountError) {
      setError(accountError instanceof Error ? accountError.message : 'Không thể tạo quản trị viên');
      throw accountError;
    } finally {
      setBusy(false);
    }
  }

  async function updateAdminAccount(account: AdminUser, values: { display_name: string; role: AdminUser['role']; status: AdminUser['status']; reason: string }) {
    setBusy(true);
    setError('');
    try {
      const updated = await authedApi<AdminUser>(`/v1/admin/accounts/${account.id}`, { method: 'PATCH', body: JSON.stringify(values) });
      setAdminAccounts(items => items.map(item => item.id === updated.id ? updated : item));
      if (currentAdmin?.id === updated.id) setCurrentAdmin(updated);
      await loadAudit();
    } catch (accountError) {
      setError(accountError instanceof Error ? accountError.message : 'Không thể cập nhật quản trị viên');
      throw accountError;
    } finally {
      setBusy(false);
    }
  }

  async function logout() {
    if (refreshToken) {
      try { await api('/v1/auth/logout', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) }); } catch {}
    }
    clearSession();
  }

  const pendingDrivers = useMemo(() => drivers.filter(driver => driver.approval_status === 'pending'), [drivers]);
  const incidents = useMemo(() => trips.filter(trip => trip.incident_open), [trips]);
  const customerMap = useMemo(() => new Map(customers.map(customer => [customer.id, customer])), [customers]);
  const driverMap = useMemo(() => new Map(drivers.map(driver => [driver.id, driver])), [drivers]);
  const vehicleMap = useMemo(() => new Map(vehicles.map(vehicle => [vehicle.id, vehicle])), [vehicles]);
  const visibleNavItems = useMemo(() => navItems.filter(item => !item.superOnly || currentAdmin?.role === 'super_admin'), [currentAdmin?.role]);

  const tripColumns = useMemo<Array<ColumnDef<typeof dataTableFeatures, Trip>>>(() => [
    { accessorKey: 'id', header: 'Mã việc', cell: ({ row }) => <span className='font-mono text-xs font-semibold text-foreground'>{row.original.id}</span> },
    { accessorKey: 'service_type', header: 'Dịch vụ', cell: ({ row }) => serviceLabel(row.original.service_type) },
    { accessorKey: 'rider_id', header: 'Khách', cell: ({ row }) => customerMap.get(row.original.rider_id)?.full_name || customerMap.get(row.original.rider_id)?.phone || row.original.rider_id },
    { accessorKey: 'driver_id', header: 'Tài xế', cell: ({ row }) => row.original.driver_id ? driverMap.get(row.original.driver_id)?.full_name || row.original.driver_id : 'Chưa ghép' },
    { id: 'fare', header: 'Giá', cell: ({ row }) => <strong className='font-semibold'>{money(row.original.final_fare_minor || row.original.estimated_fare_minor)}</strong> },
    { accessorKey: 'status', header: 'Trạng thái', cell: ({ row }) => <Badge variant={statusVariant(row.original.status, row.original.incident_open)}>{row.original.incident_open ? 'Có sự cố' : tripStatusLabel(row.original.status)}</Badge> },
  ], [customerMap, driverMap]);

  const driverColumns = useMemo<Array<ColumnDef<typeof dataTableFeatures, Driver>>>(() => [
    { accessorKey: 'full_name', header: 'Tài xế', cell: ({ row }) => <div><div className='font-semibold text-foreground'>{row.original.full_name || row.original.phone || row.original.id}</div><div className='text-xs text-muted'>{row.original.phone || row.original.id}</div></div> },
    { accessorKey: 'service_type', header: 'Dịch vụ', cell: ({ row }) => serviceLabel(row.original.service_type) },
    { accessorKey: 'availability_status', header: 'Cung ứng', cell: ({ row }) => <Badge variant={row.original.availability_status === 'online' ? 'success' : 'secondary'}>{row.original.availability_status === 'online' ? 'Online' : row.original.availability_status === 'busy' ? 'Đang bận' : 'Offline'}</Badge> },
    { accessorKey: 'approval_status', header: 'Hồ sơ', cell: ({ row }) => <Badge variant={approvalVariant(row.original.approval_status)}>{approvalLabel(row.original.approval_status)}</Badge> },
  ], []);

  const customerColumns = useMemo<Array<ColumnDef<typeof dataTableFeatures, Customer>>>(() => [
    { accessorKey: 'full_name', header: 'Khách hàng', cell: ({ row }) => <div><div className='font-semibold text-foreground'>{row.original.full_name || 'Chưa cập nhật tên'}</div><div className='text-xs text-muted'>{row.original.phone}</div></div> },
    { accessorKey: 'status', header: 'Trạng thái', cell: ({ row }) => <Badge variant={row.original.status === 'active' ? 'success' : 'secondary'}>{row.original.status === 'active' ? 'Đang hoạt động' : row.original.status}</Badge> },
    { id: 'vehicles', header: 'Xe', cell: ({ row }) => vehicles.filter(vehicle => vehicle.owner_user_id === row.original.id).length },
    { id: 'jobs', header: 'Công việc', cell: ({ row }) => trips.filter(trip => trip.rider_id === row.original.id).length },
    { accessorKey: 'created_at', header: 'Tham gia', cell: ({ row }) => formatDate(row.original.created_at) },
  ], [trips, vehicles]);

  const vehicleColumns = useMemo<Array<ColumnDef<typeof dataTableFeatures, Vehicle>>>(() => [
    { accessorKey: 'license_plate', header: 'Biển số', cell: ({ row }) => <div><div className='font-semibold text-foreground'>{row.original.license_plate}</div><div className='text-xs text-muted'>{[row.original.brand, row.original.model].filter(Boolean).join(' ') || vehicleTypeLabel(row.original.type)}</div></div> },
    { accessorKey: 'type', header: 'Loại xe', cell: ({ row }) => vehicleTypeLabel(row.original.type) },
    { accessorKey: 'owner_user_id', header: 'Chủ xe', cell: ({ row }) => customerMap.get(row.original.owner_user_id)?.full_name || customerMap.get(row.original.owner_user_id)?.phone || row.original.owner_user_id },
    { id: 'jobs', header: 'Công việc', cell: ({ row }) => trips.filter(trip => trip.customer_vehicle_id === row.original.id).length },
    { accessorKey: 'status', header: 'Trạng thái', cell: ({ row }) => <Badge variant={row.original.status === 'active' ? 'success' : 'secondary'}>{row.original.status === 'active' ? 'Đang dùng' : row.original.status}</Badge> },
  ], [customerMap, trips]);

  const auditColumns = useMemo<Array<ColumnDef<typeof dataTableFeatures, AuditEntry>>>(() => [
    { accessorKey: 'created_at', header: 'Thời gian', cell: ({ row }) => formatDate(row.original.created_at) },
    { accessorKey: 'action', header: 'Hành động', cell: ({ row }) => <span className='font-medium text-foreground'>{auditActionLabel(row.original.action)}</span> },
    { accessorKey: 'resource_type', header: 'Đối tượng', cell: ({ row }) => `${row.original.resource_type}${row.original.resource_id ? ` · ${row.original.resource_id}` : ''}` },
    { accessorKey: 'actor_id', header: 'Người thao tác', cell: ({ row }) => <span className='font-mono text-xs'>{row.original.actor_id}</span> },
  ], []);

  if (!token) {
    return <LoginScreen challenge={challenge} debugCode={debugCode} busy={busy} error={error} phoneForm={phoneForm} otpForm={otpForm} requestOtp={requestOtp} verifyOtp={verifyOtp} resetChallenge={() => { setChallenge(''); setDebugCode(''); }} />;
  }

  const activeLabel = visibleNavItems.find(item => item.key === activeView)?.label || 'Tổng quan';
  const selectView = (view: ViewKey) => { setActiveView(view); setMobileNavOpen(false); };
  const initialLoading = busy && !hasLoaded;

  function renderView() {
    if (activeView === 'trips') {
      return <Section title='Công việc' description='Theo dõi vòng đời, khách, xe, tài xế và các điểm cần Operations can thiệp.'>
        <DataTable columns={tripColumns} data={trips} loading={initialLoading} onRowClick={setSelectedTrip} searchPlaceholder='Tìm mã việc, khách, tài xế...' mobileRow={trip => <TripMobileRow trip={trip} customer={customerMap.get(trip.rider_id)} driver={trip.driver_id ? driverMap.get(trip.driver_id) : undefined} />} />
      </Section>;
    }
    if (activeView === 'drivers') {
      return <Section title='Tài xế' description='Duyệt KYC, kiểm tra năng lực và quản lý trạng thái cung ứng.'>
        <div className='mb-4 grid gap-3 sm:grid-cols-3'><MiniMetric label='Chờ duyệt' value={pendingDrivers.length} tone='warning' /><MiniMetric label='Đã duyệt' value={drivers.filter(item => item.approval_status === 'approved').length} /><MiniMetric label='Online' value={drivers.filter(item => item.availability_status === 'online').length} /></div>
        <DataTable columns={driverColumns} data={drivers} loading={initialLoading} onRowClick={openDriver} searchPlaceholder='Tìm tên, số điện thoại, năng lực...' mobileRow={driver => <DriverMobileRow driver={driver} />} />
      </Section>;
    }
    if (activeView === 'customers') {
      return <Section title='Khách hàng' description='Danh sách khách thật từ hệ thống, kèm số xe và lịch sử sử dụng dịch vụ.'>
        <DataTable columns={customerColumns} data={customers} loading={initialLoading} onRowClick={setSelectedCustomer} searchPlaceholder='Tìm tên hoặc số điện thoại...' mobileRow={customer => <CustomerMobileRow customer={customer} jobs={trips.filter(trip => trip.rider_id === customer.id).length} vehicles={vehicles.filter(vehicle => vehicle.owner_user_id === customer.id).length} />} />
      </Section>;
    }
    if (activeView === 'vehicles') {
      return <Section title='Xe khách' description='Phương tiện thuộc khách đang được dùng cho ba dịch vụ FlashX.'>
        <DataTable columns={vehicleColumns} data={vehicles} loading={initialLoading} onRowClick={setSelectedVehicle} searchPlaceholder='Tìm biển số, hãng xe, chủ xe...' mobileRow={vehicle => <VehicleMobileRow vehicle={vehicle} owner={customerMap.get(vehicle.owner_user_id)} jobs={trips.filter(trip => trip.customer_vehicle_id === vehicle.id).length} />} />
      </Section>;
    }
    if (activeView === 'incidents') {
      return <Section title='Sự cố' description='Hàng đợi ưu tiên cao. Công việc có sự cố bị chặn tiến trình cho tới khi Operations xử lý.'>
        {incidents.length ? <div className='mb-4 rounded-2xl border border-danger/10 bg-danger-soft/60 p-4 text-sm leading-6 text-danger'><strong>{incidents.length} sự cố đang mở.</strong> Mở từng công việc để xem loại sự cố, ghi chú và đóng sự cố sau khi đã xử lý thực tế.</div> : null}
        <DataTable columns={tripColumns} data={incidents} loading={initialLoading} onRowClick={setSelectedTrip} searchPlaceholder='Tìm sự cố...' mobileRow={trip => <TripMobileRow trip={trip} customer={customerMap.get(trip.rider_id)} driver={trip.driver_id ? driverMap.get(trip.driver_id) : undefined} />} />
      </Section>;
    }
    if (activeView === 'pricing') {
      const canEditPricing = currentAdmin?.role === 'super_admin';
      return <Section title='Bảng giá' description={canEditPricing ? 'Quản lý rule giá đang dùng thật cho ba dịch vụ MVP. Mỗi lần lưu tạo một version mới và được ghi audit.' : 'Theo dõi rule giá hiện hành. Tài khoản Operations chỉ có quyền xem.'}>
        <div className='mb-4 rounded-2xl border border-primary/10 bg-primary-soft/55 p-4 text-sm leading-6 text-muted'><strong className='text-primary'>{canEditPricing ? 'An toàn tài chính:' : 'Quyền truy cập:'}</strong> {canEditPricing ? 'thay đổi chỉ áp dụng cho báo giá/công việc tạo sau khi lưu. Công việc đã tạo giữ nguyên fare snapshot trước đó.' : 'chỉ Super Admin mới được tạo version giá mới; quyền này cũng được backend kiểm tra, không chỉ ẩn ở giao diện.'}</div>
        <div className='grid gap-4 xl:grid-cols-3'>{pricing.map(rule => canEditPricing ? <PricingEditor key={`${rule.service_type}:${rule.pricing_version}`} rule={rule} busy={busy} onSave={values => updatePricingRule(rule.service_type, values)} /> : <PricingSummary key={`${rule.service_type}:${rule.pricing_version}`} rule={rule} />)}</div>
      </Section>;
    }
    if (activeView === 'audit') {
      return <Section title='Nhật ký' description='Lịch sử thao tác quản trị được đọc trực tiếp từ audit_logs.'>
        <DataTable columns={auditColumns} data={audit} loading={initialLoading} searchPlaceholder='Tìm hành động, đối tượng, admin...' mobileRow={entry => <AuditMobileRow entry={entry} />} />
      </Section>;
    }
    if (activeView === 'accounts' && currentAdmin?.role === 'super_admin') {
      return <AdminAccountsView currentAdmin={currentAdmin} accounts={adminAccounts} busy={busy} onCreate={createAdminAccount} onUpdate={updateAdminAccount} />;
    }
    if (activeView === 'settings') {
      return <Section title='Cài đặt hệ thống' description='Thông tin vận hành an toàn được backend công bố cho Admin; secret không được đưa xuống trình duyệt.'>
        {systemInfo ? <div className='grid gap-4 md:grid-cols-2 xl:grid-cols-3'><SystemCard icon={Database} title='Persistence' value={systemInfo.persistence} detail={`Môi trường: ${systemInfo.app_env}`} /><SystemCard icon={Wifi} title='Realtime' value={systemInfo.features.realtime ? 'Sẵn sàng' : 'Không khả dụng'} detail='Trạng thái công việc và kết nối realtime.' /><SystemCard icon={FileCheck2} title='KYC documents' value={systemInfo.features.driver_documents ? 'Sẵn sàng' : 'Không khả dụng'} detail='Upload trực tiếp object storage + review Admin.' /><SystemCard icon={BadgeDollarSign} title='Payments' value={systemInfo.features.payments ? 'Sẵn sàng' : 'Không khả dụng'} detail='Payment service hiện được backend khởi tạo.' /><SystemCard icon={SearchCheck} title='Ratings' value={systemInfo.features.ratings ? 'Sẵn sàng' : 'Không khả dụng'} detail='Đánh giá sau công việc.' /><SystemCard icon={ShieldCheck} title='Dev identity' value={systemInfo.allow_dev_identity ? 'Đang bật' : 'Đã tắt'} detail='Chỉ dùng cho môi trường phát triển/demo khi được cấu hình.' /></div> : <InfoState icon={Settings} title='Đang tải cấu hình' text='Thông tin hệ thống chưa được đồng bộ.' />}
      </Section>;
    }

    return <Overview
      metrics={metrics}
      initialLoading={initialLoading}
      incidents={incidents}
      pendingDrivers={pendingDrivers}
      trips={trips}
      customerMap={customerMap}
      driverMap={driverMap}
      onSelectView={selectView}
      onTrip={setSelectedTrip}
      onDriver={openDriver}
    />;
  }

  const selectedTripCustomer = selectedTrip ? customerMap.get(selectedTrip.rider_id) : undefined;
  const selectedTripDriver = selectedTrip?.driver_id ? driverMap.get(selectedTrip.driver_id) : undefined;
  const selectedTripVehicle = selectedTrip?.customer_vehicle_id ? vehicleMap.get(selectedTrip.customer_vehicle_id) : undefined;

  return <div className='min-h-screen lg:grid lg:grid-cols-[264px_1fr]'>
    <aside className='hidden border-r border-border bg-surface/80 p-4 backdrop-blur-xl lg:flex lg:min-h-screen lg:flex-col'><Brand /><nav className='mt-7 flex flex-col gap-1'>{visibleNavItems.map(({ key, label, icon: Icon }) => <button key={key} onClick={() => selectView(key)} className={`flex min-h-11 items-center gap-3 rounded-2xl px-3.5 text-left text-sm font-medium transition-colors ${activeView === key ? 'bg-primary-soft text-primary' : 'text-muted hover:bg-surface-soft hover:text-foreground'}`}><Icon aria-hidden='true' className='size-[18px]' /><span>{label}</span>{activeView === key ? <span className='ml-auto size-1.5 rounded-full bg-primary' /> : null}</button>)}</nav><div className='mt-auto rounded-2xl border border-primary/10 bg-primary-soft/55 p-4'><div className='flex items-center gap-2 text-sm font-semibold text-primary'><ShieldCheck aria-hidden='true' className='size-4' /> {currentAdmin?.display_name || 'FlashX Admin'}</div><p className='mt-2 text-xs leading-5 text-muted'>{currentAdmin ? `${adminRoleLabel(currentAdmin.role)} · ${currentAdmin.phone}` : 'Đang đồng bộ quyền quản trị…'}</p></div></aside>

    {mobileNavOpen ? <div className='fixed inset-0 z-50 bg-foreground/20 p-3 backdrop-blur-sm lg:hidden' onClick={() => setMobileNavOpen(false)}><div role='dialog' aria-modal='true' aria-label='Điều hướng quản trị' className='fx-glass ml-auto flex h-full w-[min(88vw,340px)] flex-col rounded-[28px] p-4' onClick={event => event.stopPropagation()}><div className='flex items-center justify-between'><Brand /><Button size='icon' variant='ghost' onClick={() => setMobileNavOpen(false)} aria-label='Đóng menu'><X aria-hidden='true' className='size-5' /></Button></div><nav className='mt-6 flex flex-col gap-1'>{visibleNavItems.map(({ key, label, icon: Icon }) => <button key={key} onClick={() => selectView(key)} className={`flex min-h-12 items-center gap-3 rounded-2xl px-4 text-left text-sm font-semibold ${activeView === key ? 'bg-primary text-primary-foreground' : 'text-foreground hover:bg-surface-soft'}`}><Icon aria-hidden='true' className='size-5' />{label}</button>)}</nav></div></div> : null}

    <main className='min-w-0'><header className='sticky top-0 z-30 border-b border-border/80 bg-surface/80 backdrop-blur-xl'><div className='flex h-16 items-center justify-between px-4 md:px-6'><div className='flex items-center gap-3'><Button size='icon' variant='ghost' className='lg:hidden' onClick={() => setMobileNavOpen(true)} aria-label='Mở menu'><Menu aria-hidden='true' className='size-5' /></Button><div><div className='text-[11px] font-semibold uppercase tracking-[.13em] text-primary'>FlashX Operations</div><h1 className='text-lg font-semibold tracking-tight text-foreground'>{activeLabel}</h1></div></div><div className='flex items-center gap-2'><Badge variant={error ? 'destructive' : busy ? 'secondary' : 'success'}><span className={`mr-1.5 size-1.5 rounded-full ${error ? 'bg-danger' : busy ? 'bg-muted-soft' : 'bg-success'}`} />{error ? 'Cần kiểm tra' : busy ? 'Đang đồng bộ' : 'Đã đồng bộ'}</Badge><Button size='icon' variant='ghost' onClick={() => { void loadCore(); void loadReference(); }} disabled={busy} aria-label='Làm mới'><RefreshCw aria-hidden='true' className={`size-4 ${busy ? 'animate-spin motion-reduce:animate-none' : ''}`} /></Button><Button size='icon' variant='ghost' onClick={() => void logout()} aria-label='Đăng xuất'><LogOut aria-hidden='true' className='size-4' /></Button></div></div></header><div className='mx-auto max-w-[1500px] p-4 md:p-6 lg:p-7'>{error ? <div role='alert' className='mb-4 flex items-start justify-between gap-3 rounded-2xl border border-danger/15 bg-danger-soft p-4 text-sm text-danger'><span>{error}</span><button type='button' className='font-semibold' onClick={() => setError('')}>Đóng</button></div> : null}{renderView()}</div></main>

    <DetailPanel open={Boolean(selectedTrip)} title={selectedTrip ? `Công việc ${selectedTrip.id}` : 'Chi tiết công việc'} description={selectedTrip ? `${serviceLabel(selectedTrip.service_type)} · ${tripStatusLabel(selectedTrip.status)}` : undefined} onClose={() => setSelectedTrip(null)}>
      {selectedTrip ? <TripDetail trip={selectedTrip} customer={selectedTripCustomer} driver={selectedTripDriver} vehicle={selectedTripVehicle} busy={busy} resolutionNote={incidentResolutionNote} setResolutionNote={setIncidentResolutionNote} onResolve={() => void resolveIncident(selectedTrip)} candidates={driverCandidates} candidatesLoading={driverCandidatesLoading} assignmentReason={assignmentReason} setAssignmentReason={setAssignmentReason} onAssign={candidate => void assignTripDriver(candidate)} /> : null}
    </DetailPanel>

    <DetailPanel open={Boolean(selectedDriver)} title={selectedDriver ? selectedDriver.full_name || selectedDriver.phone || selectedDriver.id : 'Chi tiết tài xế'} description={selectedDriver ? `${serviceLabel(selectedDriver.service_type)} · ${approvalLabel(selectedDriver.approval_status)}` : undefined} onClose={() => { setSelectedDriver(null); setDriverDocs([]); }}>
      {selectedDriver ? <DriverDetail driver={selectedDriver} documents={driverDocs} loading={driverDocsLoading} busy={busy} reason={driverReason} setReason={setDriverReason} documentNote={documentNote} setDocumentNote={setDocumentNote} onApproval={status => void setDriverApproval(selectedDriver, status)} onReview={(item, status) => void reviewDocument(item, status)} /> : null}
    </DetailPanel>

    <DetailPanel open={Boolean(selectedCustomer)} title={selectedCustomer?.full_name || selectedCustomer?.phone || 'Chi tiết khách hàng'} description={selectedCustomer?.phone} onClose={() => setSelectedCustomer(null)}>
      {selectedCustomer ? <CustomerDetail customer={selectedCustomer} trips={trips.filter(trip => trip.rider_id === selectedCustomer.id)} vehicles={vehicles.filter(vehicle => vehicle.owner_user_id === selectedCustomer.id)} onTrip={trip => { setSelectedCustomer(null); setSelectedTrip(trip); }} /> : null}
    </DetailPanel>

    <DetailPanel open={Boolean(selectedVehicle)} title={selectedVehicle?.license_plate || 'Chi tiết xe khách'} description={selectedVehicle ? `${vehicleTypeLabel(selectedVehicle.type)} · ${[selectedVehicle.brand, selectedVehicle.model].filter(Boolean).join(' ')}` : undefined} onClose={() => setSelectedVehicle(null)}>
      {selectedVehicle ? <VehicleDetail vehicle={selectedVehicle} owner={customerMap.get(selectedVehicle.owner_user_id)} trips={trips.filter(trip => trip.customer_vehicle_id === selectedVehicle.id)} onTrip={trip => { setSelectedVehicle(null); setSelectedTrip(trip); }} /> : null}
    </DetailPanel>
  </div>;
}

function Overview({ metrics, initialLoading, incidents, pendingDrivers, trips, customerMap, driverMap, onSelectView, onTrip, onDriver }: {
  metrics: Metrics;
  initialLoading: boolean;
  incidents: Trip[];
  pendingDrivers: Driver[];
  trips: Trip[];
  customerMap: Map<string, Customer>;
  driverMap: Map<string, Driver>;
  onSelectView: (view: ViewKey) => void;
  onTrip: (trip: Trip) => void;
  onDriver: (driver: Driver) => void;
}) {
  const metricsCards = [
    { label: 'Tổng công việc', value: metrics.trips_total || 0, foot: `${metrics.trips_completed || 0} hoàn thành`, icon: ClipboardList },
    { label: 'Tài xế online', value: metrics.drivers_online || 0, foot: `${metrics.drivers_total || 0} tổng tài xế`, icon: UserRound },
    { label: 'Đang vận hành', value: metrics.trips_active || 0, foot: `${metrics.trips_searching || 0} đang tìm`, icon: Gauge },
    { label: 'Cần chú ý', value: metrics.trips_incident || 0, foot: `${metrics.drivers_pending || 0} tài xế chờ duyệt`, icon: AlertTriangle },
  ];
  return <div className='flex flex-col gap-5 fx-enter'>
    <section className='fx-flowline overflow-hidden rounded-[30px] border border-primary/10 bg-gradient-to-br from-surface via-surface to-primary-soft p-5 shadow-sm md:p-7'><div className='flex flex-col justify-between gap-6 lg:flex-row lg:items-end'><div className='max-w-2xl'><div className='mb-3 flex items-center gap-2 text-xs font-semibold uppercase tracking-[.14em] text-primary'><Sparkles aria-hidden='true' className='size-4' /> Operations pulse</div><h2 className='text-3xl font-bold tracking-[-.045em] text-foreground md:text-4xl'>Việc cần xử lý nằm trước, số liệu nằm sau.</h2><p className='mt-3 max-w-xl text-sm leading-6 text-muted md:text-base'>Hiện có <strong className='text-foreground'>{incidents.length}</strong> sự cố mở, <strong className='text-foreground'>{pendingDrivers.length}</strong> hồ sơ tài xế chờ duyệt và <strong className='text-foreground'>{metrics.trips_active || 0}</strong> công việc đang vận hành.</p></div><div className='flex items-center gap-3'><div className='fx-pulse-dot size-2 rounded-full bg-primary' /><div><div className='text-xs font-semibold text-primary'>Dữ liệu vận hành</div><div className='text-xs text-muted'>Tự đồng bộ mỗi 5 giây</div></div></div></div></section>
    <div className='grid gap-4 sm:grid-cols-2 xl:grid-cols-4'>{metricsCards.map(({ label, value, foot, icon: Icon }) => <Card key={label} className='transition-shadow duration-200 hover:shadow-md'><CardContent className='p-5'><div className='flex items-start justify-between gap-4'><div><p className='text-sm font-medium text-muted'>{label}</p>{initialLoading ? <div className='mt-3 h-9 w-20 animate-pulse rounded-lg bg-surface-soft motion-reduce:animate-none' /> : <p className='mt-2 text-3xl font-bold tracking-[-.04em] text-foreground'>{value}</p>}<p className='mt-1 text-xs text-muted'>{initialLoading ? 'Đang đồng bộ…' : foot}</p></div><div className='grid size-11 place-items-center rounded-2xl bg-primary-soft text-primary'><Icon aria-hidden='true' className='size-5' /></div></div></CardContent></Card>)}</div>
    <div className='grid gap-5 xl:grid-cols-[1.2fr_.8fr]'>
      <Card><CardHeader className='flex-row items-start justify-between'><div><CardTitle>Ưu tiên xử lý</CardTitle><CardDescription>Sự cố trước, hồ sơ tài xế sau.</CardDescription></div><Button size='sm' variant='ghost' onClick={() => onSelectView(incidents.length ? 'incidents' : 'drivers')}>Mở hàng đợi <ChevronRight aria-hidden='true' className='size-4' /></Button></CardHeader><CardContent className='grid gap-3'>{incidents.slice(0, 3).map(trip => <button key={trip.id} type='button' onClick={() => onTrip(trip)} className='flex min-h-16 items-center justify-between gap-3 rounded-2xl border border-danger/10 bg-danger-soft/60 p-4 text-left transition-colors hover:bg-danger-soft'><div><div className='font-semibold text-foreground'>{serviceLabel(trip.service_type)} · {trip.id}</div><div className='mt-1 text-xs text-muted'>{trip.incident_type || 'Sự cố cần kiểm tra'} · {customerMap.get(trip.rider_id)?.full_name || customerMap.get(trip.rider_id)?.phone || trip.rider_id}</div></div><ChevronRight aria-hidden='true' className='size-4 text-danger' /></button>)}{!incidents.length ? pendingDrivers.slice(0, 3).map(driver => <button key={driver.id} type='button' onClick={() => onDriver(driver)} className='flex min-h-16 items-center justify-between gap-3 rounded-2xl border border-warning/10 bg-warning-soft/50 p-4 text-left'><div><div className='font-semibold text-foreground'>{driver.full_name || driver.phone}</div><div className='mt-1 text-xs text-muted'>{serviceLabel(driver.service_type)} · Chờ duyệt KYC</div></div><ChevronRight aria-hidden='true' className='size-4 text-warning' /></button>) : null}{!incidents.length && !pendingDrivers.length ? <InfoState icon={CheckCircle2} title='Không có việc khẩn cấp' text='Hiện không có sự cố mở hoặc hồ sơ tài xế chờ duyệt.' /> : null}</CardContent></Card>
      <Card><CardHeader className='flex-row items-start justify-between'><div><CardTitle>Công việc gần đây</CardTitle><CardDescription>Mở nhanh chi tiết vòng đời.</CardDescription></div><Button size='sm' variant='ghost' onClick={() => onSelectView('trips')}>Tất cả <ChevronRight aria-hidden='true' className='size-4' /></Button></CardHeader><CardContent className='grid gap-2'>{trips.slice(0, 6).map(trip => <button key={trip.id} type='button' onClick={() => onTrip(trip)} className='flex min-h-14 items-center justify-between gap-3 rounded-2xl px-3 py-2 text-left transition-colors hover:bg-surface-soft'><div className='min-w-0'><div className='truncate text-sm font-semibold text-foreground'>{serviceLabel(trip.service_type)} · {customerMap.get(trip.rider_id)?.full_name || customerMap.get(trip.rider_id)?.phone || trip.id}</div><div className='mt-1 text-xs text-muted'>{trip.driver_id ? driverMap.get(trip.driver_id)?.full_name || trip.driver_id : 'Chưa ghép tài xế'} · {formatDate(trip.created_at)}</div></div><Badge variant={statusVariant(trip.status, trip.incident_open)}>{trip.incident_open ? 'Sự cố' : tripStatusLabel(trip.status)}</Badge></button>)}</CardContent></Card>
    </div>
  </div>;
}

function TripDetail({ trip, customer, driver, vehicle, busy, resolutionNote, setResolutionNote, onResolve, candidates, candidatesLoading, assignmentReason, setAssignmentReason, onAssign }: {
  trip: Trip;
  customer?: Customer;
  driver?: Driver;
  vehicle?: Vehicle;
  busy: boolean;
  resolutionNote: string;
  setResolutionNote: (value: string) => void;
  onResolve: () => void;
  candidates: DriverCandidate[];
  candidatesLoading: boolean;
  assignmentReason: string;
  setAssignmentReason: (value: string) => void;
  onAssign: (candidate: DriverCandidate) => void;
}) {
  const steps = tripSteps(trip);
  const manageable = canManageDriver(trip);
  return <div className='grid gap-5'>
    {trip.incident_open ? <div className='rounded-2xl border border-danger/15 bg-danger-soft p-4'><div className='flex items-start gap-3'><AlertTriangle aria-hidden='true' className='mt-0.5 size-5 text-danger' /><div><div className='font-semibold text-danger'>{trip.incident_type || 'Sự cố đang mở'}</div><p className='mt-1 text-sm leading-6 text-foreground'>{trip.incident_note || 'Tài xế chưa để lại ghi chú chi tiết.'}</p></div></div><label className='mt-4 block text-sm font-semibold text-foreground'>Ghi chú xử lý<Input className='mt-1.5' value={resolutionNote} onChange={event => setResolutionNote(event.target.value)} placeholder='Ví dụ: Đã gọi xác nhận với khách và tài xế' /></label><Button className='mt-3 w-full sm:w-auto' disabled={busy} onClick={onResolve}>Đóng sự cố</Button></div> : null}
    <Card className='rounded-2xl shadow-none'><CardHeader><div className='flex items-start justify-between gap-3'><div><CardTitle className='text-base'>Điều phối tài xế</CardTitle><CardDescription>{manageable ? 'Chỉ hiển thị tài xế đã duyệt, online, đúng năng lực và có vị trí mới gần điểm nhận.' : 'Điều phối thủ công đã khóa ở trạng thái hiện tại để bảo vệ chuỗi bàn giao xe.'}</CardDescription></div><Badge variant={manageable ? 'success' : 'secondary'}>{manageable ? 'Có thể điều phối' : 'Đã khóa'}</Badge></div></CardHeader><CardContent>{manageable ? <div className='grid gap-3'>{trip.driver_id ? <label className='block text-sm font-semibold text-foreground'>Lý do đổi tài xế<Input className='mt-1.5' value={assignmentReason} onChange={event => setAssignmentReason(event.target.value)} placeholder='Bắt buộc khi thay tài xế đang nhận việc' /></label> : null}{candidatesLoading ? Array.from({ length: 3 }, (_, index) => <div key={index} className='h-16 animate-pulse rounded-2xl bg-surface-soft motion-reduce:animate-none' />) : candidates.length ? candidates.map(candidate => <div key={candidate.driver.id} className='flex flex-col gap-3 rounded-2xl border border-border p-3 sm:flex-row sm:items-center sm:justify-between'><div className='min-w-0'><div className='font-semibold text-foreground'>{candidate.driver.full_name || candidate.driver.phone || candidate.driver.id}</div><div className='mt-1 text-xs text-muted'>{candidate.driver.phone || candidate.driver.id} · {Math.max(0.1, candidate.distance_to_pickup_m / 1000).toLocaleString('vi-VN', { maximumFractionDigits: 1 })} km tới điểm nhận</div></div><Button size='sm' disabled={busy || Boolean(trip.driver_id && !assignmentReason.trim())} onClick={() => onAssign(candidate)}>{trip.driver_id ? 'Đổi tài xế' : 'Gán tài xế'}</Button></div>) : <InfoState icon={UserRound} title='Chưa có tài xế phù hợp' text='Không có tài xế online đủ năng lực với vị trí mới trong bán kính điều phối hiện tại.' />}</div> : <div className='rounded-2xl bg-surface-soft p-4 text-sm leading-6 text-muted'>{trip.status === 'vehicle_received' || !['scheduled', 'searching', 'completed', 'cancelled'].includes(trip.status) ? 'Từ lúc tài xế xác nhận đã nhận xe, FlashX không cho phép đổi tài xế theo luồng thông thường. Nếu bắt buộc phải chuyển người giữ xe, Operations phải xử lý theo quy trình sự cố/bàn giao có ghi nhận.' : 'Công việc này không ở trạng thái cho phép điều phối thủ công.'}</div>}</CardContent></Card>
    <div className='grid gap-3 sm:grid-cols-2'><InfoCard label='Khách hàng' value={customer?.full_name || customer?.phone || trip.rider_id} detail={customer?.phone} icon={UsersRound} /><InfoCard label='Tài xế' value={driver?.full_name || (trip.driver_id ? trip.driver_id : 'Chưa ghép')} detail={driver?.phone} icon={UserRound} /><InfoCard label='Xe khách' value={vehicle?.license_plate || trip.customer_vehicle_id || 'Chưa gắn xe'} detail={vehicle ? [vehicle.brand, vehicle.model].filter(Boolean).join(' ') : undefined} icon={CarFront} /><InfoCard label='Giá hiện tại' value={money(trip.final_fare_minor || trip.estimated_fare_minor)} detail={trip.booking_mode === 'scheduled' ? `Hẹn ${formatDate(trip.scheduled_at)}` : 'Đặt ngay'} icon={BadgeDollarSign} /></div>
    <Card className='rounded-2xl shadow-none'><CardHeader><CardTitle className='text-base'>FlashX Flowline</CardTitle><CardDescription>Vòng đời thực tế của {serviceLabel(trip.service_type).toLocaleLowerCase('vi')}.</CardDescription></CardHeader><CardContent><div className='grid gap-0'>{steps.map((step, index) => <div key={step.status} className='grid grid-cols-[24px_1fr] gap-3'><div className='relative flex justify-center'>{index < steps.length - 1 ? <span className={`absolute top-5 h-full w-px ${step.done ? 'bg-primary' : 'bg-border'}`} /> : null}<span className={`relative mt-1 size-3 rounded-full border-2 ${step.active ? 'fx-pulse-dot border-primary bg-surface' : step.done ? 'border-primary bg-primary' : 'border-border bg-surface'}`} /></div><div className='pb-5'><div className={`text-sm font-semibold ${step.active ? 'text-primary' : step.done ? 'text-foreground' : 'text-muted'}`}>{step.label}</div>{step.active ? <div className='mt-1 text-xs text-muted'>Trạng thái hiện tại</div> : null}</div></div>)}</div></CardContent></Card>
    <div className='grid gap-3 sm:grid-cols-2'><InfoCard label='Điểm nhận' value={coordinate(trip.pickup)} detail='Tọa độ backend hiện có' icon={MapPin} /><InfoCard label='Điểm đến / trả xe' value={coordinate(trip.destination)} detail='Tọa độ backend hiện có' icon={MapPin} /><InfoCard label='Quãng đường dự kiến' value={km(trip.estimated_distance_m)} icon={Gauge} /><InfoCard label='Thời gian dự kiến' value={duration(trip.estimated_duration_s)} icon={Clock3} /></div>
    {trip.service_type === 'vehicle_inspection_assist' ? <InfoCard label='Kết quả đăng kiểm' value={trip.inspection_result || 'Chưa có kết quả'} detail='Kết quả đăng kiểm độc lập với trạng thái hoàn thành dịch vụ.' icon={FileCheck2} /> : null}
  </div>;
}

function DriverDetail({ driver, documents, loading, busy, reason, setReason, documentNote, setDocumentNote, onApproval, onReview }: {
  driver: Driver;
  documents: DriverDocumentWithURL[];
  loading: boolean;
  busy: boolean;
  reason: string;
  setReason: (value: string) => void;
  documentNote: string;
  setDocumentNote: (value: string) => void;
  onApproval: (status: 'approved' | 'rejected' | 'suspended') => void;
  onReview: (item: DriverDocumentWithURL, status: 'approved' | 'rejected') => void;
}) {
  return <div className='grid gap-5'>
    <div className='grid gap-3 sm:grid-cols-2'><InfoCard label='Số điện thoại' value={driver.phone || '—'} icon={UserRound} /><InfoCard label='Cung ứng' value={driver.availability_status === 'online' ? 'Online' : driver.availability_status === 'busy' ? 'Đang bận' : 'Offline'} detail={driver.location ? `Vị trí cuối: ${driver.location.lat.toFixed(5)}, ${driver.location.lng.toFixed(5)}` : 'Chưa có vị trí gần nhất'} icon={Wifi} /><InfoCard label='Năng lực' value={(driver.capabilities || []).map(serviceLabel).join(', ') || serviceLabel(driver.service_type)} icon={ShieldCheck} /><InfoCard label='Trạng thái hồ sơ' value={approvalLabel(driver.approval_status)} icon={FileCheck2} /></div>
    <Card className='rounded-2xl shadow-none'><CardHeader><CardTitle className='text-base'>Quyết định tài xế</CardTitle><CardDescription>Duyệt, từ chối hoặc tạm khóa đều được ghi audit.</CardDescription></CardHeader><CardContent><label className='block text-sm font-semibold text-foreground'>Lý do khi từ chối / tạm khóa<Input className='mt-1.5' value={reason} onChange={event => setReason(event.target.value)} placeholder='Nhập lý do vận hành' /></label><div className='mt-3 flex flex-wrap gap-2'>{driver.approval_status !== 'approved' ? <Button disabled={busy} onClick={() => onApproval('approved')}><UserCheck aria-hidden='true' className='size-4' /> Duyệt tài xế</Button> : null}{driver.approval_status !== 'rejected' ? <Button variant='destructive' disabled={busy} onClick={() => onApproval('rejected')}><Ban aria-hidden='true' className='size-4' /> Từ chối</Button> : null}{driver.approval_status === 'approved' ? <Button variant='outline' disabled={busy} onClick={() => onApproval('suspended')}><Ban aria-hidden='true' className='size-4' /> Tạm khóa</Button> : null}</div></CardContent></Card>
    <Card className='rounded-2xl shadow-none'><CardHeader><CardTitle className='text-base'>Hồ sơ KYC</CardTitle><CardDescription>Mỗi tài liệu có signed URL riêng; file không đi xuyên qua backend.</CardDescription></CardHeader><CardContent className='grid gap-3'>{loading ? Array.from({ length: 3 }, (_, index) => <div key={index} className='h-28 animate-pulse rounded-2xl bg-surface-soft motion-reduce:animate-none' />) : documents.length ? documents.map(item => <div key={item.document.id} className='rounded-2xl border border-border p-4'><div className='flex flex-col justify-between gap-3 sm:flex-row sm:items-start'><div><div className='font-semibold text-foreground'>{documentLabel(item.document.document_type)}</div><div className='mt-1 text-xs text-muted'>{item.document.filename} · {Math.max(1, Math.round(item.document.size_bytes / 1024))} KB</div></div><Badge variant={approvalVariant(item.document.review_status)}>{approvalLabel(item.document.review_status)}</Badge></div>{item.document.review_note ? <p className='mt-3 rounded-xl bg-surface-soft p-3 text-sm text-muted'>Ghi chú: {item.document.review_note}</p> : null}<div className='mt-3 flex flex-wrap gap-2'><a href={item.view.url} target='_blank' rel='noreferrer' className='inline-flex min-h-10 items-center gap-2 rounded-xl border border-border px-3 text-sm font-semibold text-foreground hover:bg-surface-soft'>Xem tài liệu <ExternalLink aria-hidden='true' className='size-4' /></a><Button size='sm' variant='outline' disabled={busy} onClick={() => onReview(item, 'rejected')}>Từ chối file</Button><Button size='sm' disabled={busy} onClick={() => onReview(item, 'approved')}>Duyệt file</Button></div></div>) : <InfoState icon={FileCheck2} title='Chưa có tài liệu' text='Tài xế chưa hoàn tất upload hồ sơ KYC.' />}<label className='block text-sm font-semibold text-foreground'>Ghi chú cho lần review tiếp theo<Input className='mt-1.5' value={documentNote} onChange={event => setDocumentNote(event.target.value)} placeholder='Không bắt buộc khi duyệt' /></label></CardContent></Card>
  </div>;
}

function CustomerDetail({ customer, trips, vehicles, onTrip }: { customer: Customer; trips: Trip[]; vehicles: Vehicle[]; onTrip: (trip: Trip) => void }) {
  return <div className='grid gap-5'><div className='grid gap-3 sm:grid-cols-2'><InfoCard label='Số điện thoại' value={customer.phone} icon={UserRound} /><InfoCard label='Trạng thái' value={customer.status === 'active' ? 'Đang hoạt động' : customer.status} icon={ShieldCheck} /><InfoCard label='Số xe' value={String(vehicles.length)} icon={CarFront} /><InfoCard label='Tổng công việc' value={String(trips.length)} icon={ClipboardList} /></div><Card className='rounded-2xl shadow-none'><CardHeader><CardTitle className='text-base'>Xe đã khai báo</CardTitle></CardHeader><CardContent className='grid gap-2'>{vehicles.length ? vehicles.map(vehicle => <div key={vehicle.id} className='rounded-2xl bg-surface-soft p-3'><div className='font-semibold text-foreground'>{vehicle.license_plate}</div><div className='mt-1 text-xs text-muted'>{vehicleTypeLabel(vehicle.type)} · {[vehicle.brand, vehicle.model].filter(Boolean).join(' ') || 'Chưa cập nhật hãng/model'}</div></div>) : <p className='text-sm text-muted'>Khách chưa khai báo xe.</p>}</CardContent></Card><Card className='rounded-2xl shadow-none'><CardHeader><CardTitle className='text-base'>Công việc gần đây</CardTitle></CardHeader><CardContent className='grid gap-2'>{trips.length ? trips.slice(0, 10).map(trip => <button key={trip.id} type='button' onClick={() => onTrip(trip)} className='flex items-center justify-between gap-3 rounded-2xl p-3 text-left hover:bg-surface-soft'><div><div className='font-semibold text-foreground'>{serviceLabel(trip.service_type)}</div><div className='mt-1 text-xs text-muted'>{formatDate(trip.created_at)}</div></div><Badge variant={statusVariant(trip.status, trip.incident_open)}>{trip.incident_open ? 'Sự cố' : tripStatusLabel(trip.status)}</Badge></button>) : <p className='text-sm text-muted'>Chưa có công việc.</p>}</CardContent></Card></div>;
}

function VehicleDetail({ vehicle, owner, trips, onTrip }: { vehicle: Vehicle; owner?: Customer; trips: Trip[]; onTrip: (trip: Trip) => void }) {
  return <div className='grid gap-5'><div className='grid gap-3 sm:grid-cols-2'><InfoCard label='Chủ xe' value={owner?.full_name || owner?.phone || vehicle.owner_user_id} detail={owner?.phone} icon={UsersRound} /><InfoCard label='Loại xe' value={vehicleTypeLabel(vehicle.type)} icon={CarFront} /><InfoCard label='Thông tin xe' value={[vehicle.brand, vehicle.model, vehicle.year].filter(Boolean).join(' ') || 'Chưa cập nhật'} detail={vehicle.color || undefined} icon={CarFront} /><InfoCard label='Hộp số' value={vehicle.transmission === 'automatic' ? 'Tự động' : vehicle.transmission === 'manual' ? 'Số sàn' : 'Không áp dụng'} icon={Settings} /></div>{vehicle.notes ? <div className='rounded-2xl bg-surface-soft p-4 text-sm leading-6 text-muted'><strong className='text-foreground'>Ghi chú xe:</strong> {vehicle.notes}</div> : null}<Card className='rounded-2xl shadow-none'><CardHeader><CardTitle className='text-base'>Lịch sử công việc</CardTitle></CardHeader><CardContent className='grid gap-2'>{trips.length ? trips.map(trip => <button key={trip.id} type='button' onClick={() => onTrip(trip)} className='flex items-center justify-between gap-3 rounded-2xl p-3 text-left hover:bg-surface-soft'><div><div className='font-semibold text-foreground'>{serviceLabel(trip.service_type)}</div><div className='mt-1 text-xs text-muted'>{formatDate(trip.created_at)}</div></div><Badge variant={statusVariant(trip.status, trip.incident_open)}>{trip.incident_open ? 'Sự cố' : tripStatusLabel(trip.status)}</Badge></button>) : <p className='text-sm text-muted'>Xe chưa phát sinh công việc.</p>}</CardContent></Card></div>;
}

function TripMobileRow({ trip, customer, driver }: { trip: Trip; customer?: Customer; driver?: Driver }) {
  return <div className='rounded-2xl border border-border bg-surface p-4 shadow-sm'><div className='flex items-start justify-between gap-3'><div><div className='font-semibold text-foreground'>{serviceLabel(trip.service_type)}</div><div className='mt-1 font-mono text-[11px] text-muted'>{trip.id}</div></div><Badge variant={statusVariant(trip.status, trip.incident_open)}>{trip.incident_open ? 'Sự cố' : tripStatusLabel(trip.status)}</Badge></div><div className='mt-4 grid gap-1 text-sm text-muted'><span>Khách: <strong className='font-medium text-foreground'>{customer?.full_name || customer?.phone || trip.rider_id}</strong></span><span>Tài xế: <strong className='font-medium text-foreground'>{driver?.full_name || (trip.driver_id ? trip.driver_id : 'Chưa ghép')}</strong></span><span>Giá: <strong className='font-medium text-foreground'>{money(trip.final_fare_minor || trip.estimated_fare_minor)}</strong></span></div></div>;
}
function DriverMobileRow({ driver }: { driver: Driver }) { return <div className='rounded-2xl border border-border bg-surface p-4 shadow-sm'><div className='flex items-start justify-between gap-3'><div><div className='font-semibold text-foreground'>{driver.full_name || driver.phone || driver.id}</div><div className='mt-1 text-xs text-muted'>{driver.phone}</div></div><Badge variant={approvalVariant(driver.approval_status)}>{approvalLabel(driver.approval_status)}</Badge></div><div className='mt-3 flex items-center justify-between text-sm text-muted'><span>{serviceLabel(driver.service_type)}</span><span>{driver.availability_status === 'online' ? 'Online' : driver.availability_status === 'busy' ? 'Đang bận' : 'Offline'}</span></div></div>; }
function CustomerMobileRow({ customer, jobs, vehicles }: { customer: Customer; jobs: number; vehicles: number }) { return <div className='rounded-2xl border border-border bg-surface p-4 shadow-sm'><div className='font-semibold text-foreground'>{customer.full_name || 'Chưa cập nhật tên'}</div><div className='mt-1 text-sm text-muted'>{customer.phone}</div><div className='mt-3 flex gap-4 text-xs text-muted'><span>{vehicles} xe</span><span>{jobs} công việc</span></div></div>; }
function VehicleMobileRow({ vehicle, owner, jobs }: { vehicle: Vehicle; owner?: Customer; jobs: number }) { return <div className='rounded-2xl border border-border bg-surface p-4 shadow-sm'><div className='flex items-start justify-between gap-3'><div><div className='font-semibold text-foreground'>{vehicle.license_plate}</div><div className='mt-1 text-xs text-muted'>{[vehicle.brand, vehicle.model].filter(Boolean).join(' ') || vehicleTypeLabel(vehicle.type)}</div></div><Badge variant={vehicle.status === 'active' ? 'success' : 'secondary'}>{vehicle.status === 'active' ? 'Đang dùng' : vehicle.status}</Badge></div><div className='mt-3 text-sm text-muted'>{owner?.full_name || owner?.phone || vehicle.owner_user_id} · {jobs} công việc</div></div>; }
function AuditMobileRow({ entry }: { entry: AuditEntry }) { return <div className='rounded-2xl border border-border bg-surface p-4 shadow-sm'><div className='font-semibold text-foreground'>{auditActionLabel(entry.action)}</div><div className='mt-1 text-xs text-muted'>{formatDate(entry.created_at)}</div><div className='mt-3 font-mono text-xs text-muted'>{entry.resource_type} · {entry.resource_id || '—'}</div></div>; }

function Section({ title, description, children }: { title: string; description: string; children: ReactNode }) { return <Card className='fx-enter'><CardHeader><CardTitle className='text-xl'>{title}</CardTitle><CardDescription>{description}</CardDescription></CardHeader><CardContent>{children}</CardContent></Card>; }
function InfoState({ icon: Icon, title, text }: { icon: LucideIcon; title: string; text: string }) { return <div className='rounded-3xl border border-dashed border-border bg-surface-soft/60 p-7 text-center'><div className='mx-auto grid size-12 place-items-center rounded-2xl bg-primary-soft text-primary'><Icon aria-hidden='true' className='size-5' /></div><div className='mt-4 font-semibold text-foreground'>{title}</div><p className='mx-auto mt-1 max-w-md text-sm leading-6 text-muted'>{text}</p></div>; }
function InfoCard({ label, value, detail, icon: Icon }: { label: string; value: string; detail?: string; icon: LucideIcon }) { return <div className='rounded-2xl border border-border bg-surface p-4'><div className='flex items-start gap-3'><div className='grid size-10 shrink-0 place-items-center rounded-2xl bg-primary-soft text-primary'><Icon aria-hidden='true' className='size-4' /></div><div className='min-w-0'><div className='text-xs font-semibold text-muted'>{label}</div><div className='mt-1 break-words font-semibold text-foreground'>{value}</div>{detail ? <div className='mt-1 text-xs leading-5 text-muted'>{detail}</div> : null}</div></div></div>; }
function MiniMetric({ label, value, tone }: { label: string; value: number; tone?: 'warning' }) { return <div className={`rounded-2xl border p-4 ${tone === 'warning' && value ? 'border-warning/15 bg-warning-soft/40' : 'border-border bg-surface-soft/45'}`}><div className='text-xs font-semibold text-muted'>{label}</div><div className='mt-1 text-2xl font-bold tracking-tight text-foreground'>{value}</div></div>; }
function AdminAccountsView({ currentAdmin, accounts, busy, onCreate, onUpdate }: { currentAdmin: AdminUser; accounts: AdminUser[]; busy: boolean; onCreate: (values: AdminAccountForm) => Promise<void>; onUpdate: (account: AdminUser, values: { display_name: string; role: AdminUser['role']; status: AdminUser['status']; reason: string }) => Promise<void> }) {
  const form = useForm<AdminAccountForm>({ resolver: zodResolver(adminAccountSchema), defaultValues: { phone: '', display_name: '', role: 'operations' } });
  const submit = form.handleSubmit(async values => { try { await onCreate(values); form.reset({ phone: '', display_name: '', role: 'operations' }); } catch {} });
  const active = accounts.filter(item => item.status === 'active').length;
  const superAdmins = accounts.filter(item => item.status === 'active' && item.role === 'super_admin').length;
  return <Section title='Quản trị viên' description='Tài khoản Admin dùng số điện thoại + OTP. Quyền nhạy cảm được backend kiểm tra theo role.'>
    <div className='grid gap-3 sm:grid-cols-3'><MiniMetric label='Đang hoạt động' value={active} /><MiniMetric label='Super Admin' value={superAdmins} /><MiniMetric label='Operations' value={accounts.filter(item => item.role === 'operations').length} /></div>
    <div className='mt-5 grid gap-5 xl:grid-cols-[380px_1fr]'>
      <Card className='h-fit rounded-3xl shadow-none'><CardHeader><CardTitle className='text-base'>Thêm quản trị viên</CardTitle><CardDescription>Người mới đăng nhập bằng OTP gửi tới số điện thoại này.</CardDescription></CardHeader><CardContent><form className='grid gap-3' onSubmit={submit}><label className='text-sm font-semibold text-foreground'>Tên hiển thị<Input className='mt-1.5' {...form.register('display_name')} placeholder='Ví dụ: Vận hành Thanh Hóa' /><span className='mt-1 block text-xs font-normal text-danger'>{form.formState.errors.display_name?.message || ''}</span></label><label className='text-sm font-semibold text-foreground'>Số điện thoại<Input className='mt-1.5' {...form.register('phone')} inputMode='tel' autoComplete='tel' placeholder='0912345678' /><span className='mt-1 block text-xs font-normal text-danger'>{form.formState.errors.phone?.message || ''}</span></label><label className='text-sm font-semibold text-foreground'>Vai trò<select className='mt-1.5 min-h-11 w-full rounded-xl border border-border bg-surface px-3 text-sm text-foreground outline-none transition focus:border-primary focus:ring-2 focus:ring-primary/15' {...form.register('role')}><option value='operations'>Operations</option><option value='super_admin'>Super Admin</option></select></label><Button className='mt-2 w-full' type='submit' disabled={busy || form.formState.isSubmitting}><UserCheck aria-hidden='true' className='size-4' />{form.formState.isSubmitting ? 'Đang tạo…' : 'Thêm quản trị viên'}</Button></form><div className='mt-4 rounded-2xl bg-surface-soft p-3 text-xs leading-5 text-muted'><strong className='text-foreground'>Operations</strong> xử lý công việc, tài xế, KYC và sự cố. <strong className='text-foreground'>Super Admin</strong> có thêm quyền bảng giá và quản trị tài khoản.</div></CardContent></Card>
      <div className='grid gap-3'>{accounts.map(account => <AdminAccountCard key={account.id} account={account} current={account.id === currentAdmin.id} busy={busy} onUpdate={onUpdate} />)}{!accounts.length ? <InfoState icon={ShieldCheck} title='Chưa có dữ liệu tài khoản' text='Danh sách quản trị viên chưa được đồng bộ.' /> : null}</div>
    </div>
  </Section>;
}
function AdminAccountCard({ account, current, busy, onUpdate }: { account: AdminUser; current: boolean; busy: boolean; onUpdate: (account: AdminUser, values: { display_name: string; role: AdminUser['role']; status: AdminUser['status']; reason: string }) => Promise<void> }) {
  const [reason, setReason] = useState('');
  const apply = async (role: AdminUser['role'], status: AdminUser['status']) => { if (!reason.trim()) return; try { await onUpdate(account, { display_name: account.display_name, role, status, reason: reason.trim() }); setReason(''); } catch {} };
  return <Card className='rounded-3xl shadow-none'><CardContent className='p-5'><div className='flex flex-col justify-between gap-4 md:flex-row md:items-start'><div className='min-w-0'><div className='flex flex-wrap items-center gap-2'><div className='font-semibold text-foreground'>{account.display_name || account.phone}</div>{current ? <Badge variant='success'>Phiên hiện tại</Badge> : null}<Badge variant={account.status === 'active' ? 'success' : 'secondary'}>{account.status === 'active' ? 'Đang hoạt động' : 'Đã vô hiệu hóa'}</Badge></div><div className='mt-1 text-sm text-muted'>{account.phone} · {adminRoleLabel(account.role)}</div><div className='mt-1 text-xs text-muted'>Tạo {formatDate(account.created_at)}</div></div><div className='flex shrink-0 flex-wrap gap-2'>{account.role === 'operations' ? <Button size='sm' variant='outline' disabled={busy || !reason.trim()} onClick={() => void apply('super_admin', account.status)}>Nâng Super Admin</Button> : <Button size='sm' variant='outline' disabled={busy || !reason.trim()} onClick={() => void apply('operations', account.status)}>Chuyển Operations</Button>}{account.status === 'active' ? <Button size='sm' variant='destructive' disabled={busy || !reason.trim()} onClick={() => void apply(account.role, 'disabled')}>Vô hiệu hóa</Button> : <Button size='sm' disabled={busy || !reason.trim()} onClick={() => void apply(account.role, 'active')}>Kích hoạt</Button>}</div></div><label className='mt-4 block text-sm font-semibold text-foreground'>Lý do thay đổi<Input className='mt-1.5' value={reason} onChange={event => setReason(event.target.value)} placeholder='Bắt buộc trước khi đổi role hoặc trạng thái' /></label>{account.role === 'super_admin' && account.status === 'active' ? <p className='mt-2 text-xs leading-5 text-muted'>Backend luôn bắt buộc còn ít nhất một Super Admin đang hoạt động.</p> : null}</CardContent></Card>;
}
function PricingSummary({ rule }: { rule: PricingRule }) { return <Card className='rounded-3xl shadow-none'><CardHeader><div className='flex items-start justify-between gap-3'><div><CardTitle className='text-base'>{serviceLabel(rule.service_type)}</CardTitle><CardDescription className='mt-1 break-all text-xs'>Version: {rule.pricing_version}</CardDescription></div><Badge variant='outline'>{rule.currency}</Badge></div></CardHeader><CardContent className='grid gap-3'><PriceMetric label='Giá mở cửa' value={money(rule.base_fare_minor)} /><PriceMetric label='Giá mỗi km' value={`${money(rule.per_km_minor)} / km`} /><PriceMetric label='Phí dịch vụ' value={money(rule.service_minor)} /><PriceMetric label='Giá tối thiểu' value={money(rule.minimum_minor)} /></CardContent></Card>; }
function PriceMetric({ label, value }: { label: string; value: string }) { return <div className='flex items-center justify-between gap-3 border-b border-border pb-3 last:border-0 last:pb-0'><span className='text-sm text-muted'>{label}</span><strong className='text-sm font-semibold text-foreground'>{value}</strong></div>; }
function PricingEditor({ rule, busy, onSave }: { rule: PricingRule; busy: boolean; onSave: (values: PricingForm) => Promise<void> }) {
  const form = useForm<PricingForm>({
    resolver: zodResolver(pricingSchema),
    defaultValues: { base_fare_minor: rule.base_fare_minor, per_km_minor: rule.per_km_minor, service_minor: rule.service_minor, minimum_minor: rule.minimum_minor },
  });
  const values = form.watch();
  const submit = form.handleSubmit(async data => { try { await onSave(data); } catch {} });
  return <Card className='rounded-3xl shadow-none'><CardHeader><div className='flex items-start justify-between gap-3'><div><CardTitle className='text-base'>{serviceLabel(rule.service_type)}</CardTitle><CardDescription className='mt-1 break-all text-xs'>Version: {rule.pricing_version}</CardDescription></div><Badge variant='outline'>{rule.currency}</Badge></div></CardHeader><CardContent><form onSubmit={submit} className='grid gap-3'>
    <label className='text-sm font-semibold text-foreground'>Giá mở cửa<Input className='mt-1.5' type='number' min={0} max={100000000} step={1000} inputMode='numeric' {...form.register('base_fare_minor', { valueAsNumber: true })} /><span className='mt-1 block text-xs font-normal text-muted'>{money(Number(values.base_fare_minor) || 0)}</span><span className='text-xs font-normal text-danger'>{form.formState.errors.base_fare_minor?.message || ''}</span></label>
    <label className='text-sm font-semibold text-foreground'>Giá mỗi km<Input className='mt-1.5' type='number' min={0} max={100000000} step={1000} inputMode='numeric' {...form.register('per_km_minor', { valueAsNumber: true })} /><span className='mt-1 block text-xs font-normal text-muted'>{money(Number(values.per_km_minor) || 0)} / km</span><span className='text-xs font-normal text-danger'>{form.formState.errors.per_km_minor?.message || ''}</span></label>
    <label className='text-sm font-semibold text-foreground'>Phí dịch vụ<Input className='mt-1.5' type='number' min={0} max={100000000} step={1000} inputMode='numeric' {...form.register('service_minor', { valueAsNumber: true })} /><span className='mt-1 block text-xs font-normal text-muted'>{money(Number(values.service_minor) || 0)}</span><span className='text-xs font-normal text-danger'>{form.formState.errors.service_minor?.message || ''}</span></label>
    <label className='text-sm font-semibold text-foreground'>Giá tối thiểu<Input className='mt-1.5' type='number' min={0} max={100000000} step={1000} inputMode='numeric' {...form.register('minimum_minor', { valueAsNumber: true })} /><span className='mt-1 block text-xs font-normal text-muted'>{money(Number(values.minimum_minor) || 0)}</span><span className='text-xs font-normal text-danger'>{form.formState.errors.minimum_minor?.message || ''}</span></label>
    <Button className='mt-2 w-full' type='submit' disabled={busy || form.formState.isSubmitting || !form.formState.isDirty}><BadgeDollarSign aria-hidden='true' className='size-4' />{form.formState.isSubmitting ? 'Đang lưu…' : 'Lưu version giá mới'}</Button>
  </form></CardContent></Card>;
}
function SystemCard({ icon: Icon, title, value, detail }: { icon: LucideIcon; title: string; value: string; detail: string }) { return <div className='rounded-2xl border border-border bg-surface p-5'><div className='grid size-10 place-items-center rounded-2xl bg-primary-soft text-primary'><Icon aria-hidden='true' className='size-5' /></div><div className='mt-4 text-sm font-semibold text-muted'>{title}</div><div className='mt-1 text-xl font-bold tracking-tight text-foreground'>{value}</div><p className='mt-2 text-sm leading-6 text-muted'>{detail}</p></div>; }

function Brand() { return <div className='flex items-center gap-3 px-2'><div className='grid size-10 place-items-center rounded-2xl bg-primary text-primary-foreground shadow-sm'><Zap aria-hidden='true' className='size-5 fill-current' /></div><div><div className='text-base font-bold tracking-tight text-foreground'>FlashX</div><div className='text-[11px] font-medium text-muted'>Operations</div></div></div>; }

function LoginScreen({ challenge, debugCode, busy, error, phoneForm, otpForm, requestOtp, verifyOtp, resetChallenge }: any) {
  return <main className='relative grid min-h-screen place-items-center overflow-hidden bg-background p-5'><div className='absolute left-[-10%] top-[-20%] size-[38rem] rounded-full bg-primary-tint/50 blur-3xl' /><div className='absolute bottom-[-20%] right-[-12%] size-[34rem] rounded-full bg-primary-soft blur-3xl' /><Card className='fx-glass fx-enter relative z-10 w-full max-w-[430px] rounded-[30px]'><CardHeader className='pb-4'><Brand /><div className='pt-5'><CardTitle className='text-2xl font-bold tracking-[-.035em]'>Đăng nhập quản trị</CardTitle><CardDescription>Trung tâm vận hành FlashX · Xác thực bằng OTP.</CardDescription></div></CardHeader><CardContent className='flex flex-col gap-4'>{error ? <div role='alert' className='rounded-2xl bg-danger-soft p-3.5 text-sm text-danger'>{error}</div> : null}{!challenge ? <form onSubmit={requestOtp} className='flex flex-col gap-4'><label className='flex flex-col gap-1.5 text-sm font-semibold text-foreground'>Số điện thoại<Input {...phoneForm.register('phone')} disabled={busy} placeholder='0999999999' autoComplete='tel' /><span className='min-h-4 text-xs font-normal text-danger'>{phoneForm.formState.errors.phone?.message || ''}</span></label><Button size='lg' className='w-full' type='submit' disabled={busy}>{busy ? 'Đang xử lý…' : 'Tiếp tục'}</Button></form> : <form onSubmit={verifyOtp} className='flex flex-col gap-4'><label className='flex flex-col gap-1.5 text-sm font-semibold text-foreground'>Mã OTP<Input {...otpForm.register('code')} disabled={busy} inputMode='numeric' autoComplete='one-time-code' maxLength={6} placeholder='000000' /><span className='min-h-4 text-xs font-normal text-danger'>{otpForm.formState.errors.code?.message || ''}</span></label>{debugCode ? <div className='rounded-2xl border border-primary/10 bg-primary-soft p-3.5 text-sm text-primary'>DEV OTP: <strong className='font-mono tracking-wider'>{debugCode}</strong></div> : null}<div className='flex gap-2'><Button type='button' variant='outline' onClick={resetChallenge}>Quay lại</Button><Button className='flex-1' type='submit' disabled={busy}>Xác nhận OTP</Button></div></form>}<p className='text-center text-xs leading-5 text-muted'>Production không dựa vào DEV OTP; mã debug chỉ xuất hiện khi backend development trả về.</p></CardContent></Card></main>;
}

function tripSteps(trip: Trip) {
  const inspection = trip.service_type === 'vehicle_inspection_assist';
  const statuses = inspection
    ? ['searching', 'accepted', 'arriving_for_pickup', 'arrived_for_pickup', 'vehicle_received', 'en_route_to_inspection', 'arrived_at_inspection_center', 'inspection_in_progress', 'inspection_completed', 'returning_vehicle', 'arrived_for_return', 'handover', 'completed']
    : ['searching', 'accepted', 'arriving', 'arrived', 'vehicle_received', 'in_progress', 'handover', 'completed'];
  if (trip.status === 'scheduled') statuses.unshift('scheduled');
  const current = statuses.indexOf(trip.status);
  return statuses.map((status, index) => ({ status, label: tripStatusLabel(status), active: status === trip.status, done: trip.status === 'completed' || (current >= 0 && index < current) }));
}

function auditActionLabel(action: string) {
  return ({
    'driver.approval_changed': 'Đổi trạng thái tài xế',
    'driver.document_reviewed': 'Duyệt tài liệu KYC',
    'trip.incident_resolved': 'Đóng sự cố công việc',
    'trip.driver_assigned': 'Gán tài xế thủ công',
    'trip.driver_reassigned': 'Đổi tài xế thủ công',
    'pricing.rule_updated': 'Cập nhật bảng giá',
    'admin.account_created': 'Thêm quản trị viên',
    'admin.account_updated': 'Cập nhật quản trị viên',
  } as Record<string, string>)[action] || action;
}
