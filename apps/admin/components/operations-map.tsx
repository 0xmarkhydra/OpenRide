'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import type { Map as MapLibreMap, Marker as MapLibreMarker } from 'maplibre-gl';
import { AlertTriangle, LocateFixed, MapPin, Navigation } from 'lucide-react';

import { Badge } from './ui/badge';
import { Button } from './ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';

type DriverMapItem = {
  id: string;
  name: string;
  availability: string;
  lat: number;
  lng: number;
  capturedAt?: string;
};

type TripMapItem = {
  id: string;
  serviceLabel: string;
  statusLabel: string;
  incidentOpen: boolean;
  pickup: { lat: number; lng: number };
};

type Props = {
  drivers: DriverMapItem[];
  trips: TripMapItem[];
  onDriver: (id: string) => void;
  onTrip: (id: string) => void;
};

const MAP_STYLE = 'https://tiles.openfreemap.org/styles/liberty';
const THANH_HOA_CENTER: [number, number] = [105.776, 19.807];

export function OperationsMap({ drivers, trips, onDriver, onTrip }: Props) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<MapLibreMap | null>(null);
  const libraryRef = useRef<typeof import('maplibre-gl') | null>(null);
  const markersRef = useRef<MapLibreMarker[]>([]);
  const [ready, setReady] = useState(false);
  const [mapError, setMapError] = useState('');

  const points = useMemo(() => [
    ...drivers.map(item => ({ lng: item.lng, lat: item.lat })),
    ...trips.map(item => ({ lng: item.pickup.lng, lat: item.pickup.lat })),
  ].filter(item => Number.isFinite(item.lat) && Number.isFinite(item.lng)), [drivers, trips]);

  useEffect(() => {
    let disposed = false;
    if (!containerRef.current) return;

    void import('maplibre-gl').then(maplibre => {
      if (disposed || !containerRef.current) return;
      libraryRef.current = maplibre;
      try {
        const map = new maplibre.Map({
          container: containerRef.current,
          style: MAP_STYLE,
          center: THANH_HOA_CENTER,
          zoom: 11.2,
          attributionControl: { compact: true },
          cooperativeGestures: true,
        });
        map.addControl(new maplibre.NavigationControl({ showCompass: true, showZoom: true }), 'top-right');
        map.on('load', () => {
          if (!disposed) setReady(true);
        });
        mapRef.current = map;
      } catch {
        setMapError('Thiết bị hoặc trình duyệt không thể khởi tạo bản đồ WebGL.');
      }
    }).catch(() => setMapError('Không thể tải thư viện bản đồ.'));

    return () => {
      disposed = true;
      markersRef.current.forEach(marker => marker.remove());
      markersRef.current = [];
      mapRef.current?.remove();
      mapRef.current = null;
      libraryRef.current = null;
    };
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    const maplibre = libraryRef.current;
    if (!ready || !map || !maplibre) return;

    markersRef.current.forEach(marker => marker.remove());
    markersRef.current = [];

    for (const driver of drivers) {
      if (!Number.isFinite(driver.lat) || !Number.isFinite(driver.lng)) continue;
      const element = document.createElement('button');
      element.type = 'button';
      element.className = `fx-map-marker fx-map-marker-driver ${driver.availability === 'busy' ? 'is-busy' : ''}`;
      element.setAttribute('aria-label', `Mở tài xế ${driver.name}`);
      element.title = `${driver.name} · ${driver.availability === 'busy' ? 'Đang bận' : 'Online'}`;
      element.addEventListener('click', event => {
        event.stopPropagation();
        onDriver(driver.id);
      });
      const marker = new maplibre.Marker({ element, anchor: 'center' }).setLngLat([driver.lng, driver.lat]).addTo(map);
      markersRef.current.push(marker);
    }

    for (const trip of trips) {
      if (!Number.isFinite(trip.pickup.lat) || !Number.isFinite(trip.pickup.lng)) continue;
      const element = document.createElement('button');
      element.type = 'button';
      element.className = `fx-map-marker fx-map-marker-job ${trip.incidentOpen ? 'has-incident' : ''}`;
      element.setAttribute('aria-label', `Mở công việc ${trip.id}`);
      element.title = `${trip.serviceLabel} · ${trip.statusLabel}`;
      element.addEventListener('click', event => {
        event.stopPropagation();
        onTrip(trip.id);
      });
      const marker = new maplibre.Marker({ element, anchor: 'bottom' }).setLngLat([trip.pickup.lng, trip.pickup.lat]).addTo(map);
      markersRef.current.push(marker);
    }

    if (points.length === 1) {
      map.easeTo({ center: [points[0].lng, points[0].lat], zoom: 13.5, duration: 500 });
    } else if (points.length > 1) {
      const bounds = new maplibre.LngLatBounds();
      points.forEach(point => bounds.extend([point.lng, point.lat]));
      map.fitBounds(bounds, { padding: 70, maxZoom: 14, duration: 550 });
    } else {
      map.easeTo({ center: THANH_HOA_CENTER, zoom: 11.2, duration: 500 });
    }
  }, [drivers, trips, onDriver, onTrip, points, ready]);

  function focusOperations() {
    const map = mapRef.current;
    const maplibre = libraryRef.current;
    if (!map || !maplibre) return;
    if (points.length > 1) {
      const bounds = new maplibre.LngLatBounds();
      points.forEach(point => bounds.extend([point.lng, point.lat]));
      map.fitBounds(bounds, { padding: 70, maxZoom: 14, duration: 450 });
    } else if (points.length === 1) {
      map.easeTo({ center: [points[0].lng, points[0].lat], zoom: 13.5, duration: 450 });
    } else {
      map.easeTo({ center: THANH_HOA_CENTER, zoom: 11.2, duration: 450 });
    }
  }

  return <Card className='overflow-hidden rounded-[28px] shadow-none'>
    <CardHeader className='flex-row items-start justify-between gap-4'>
      <div>
        <div className='mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-[.13em] text-primary'><Navigation aria-hidden='true' className='size-4' /> Live operations map</div>
        <CardTitle>Bản đồ vận hành</CardTitle>
        <CardDescription>Vị trí tài xế online/busy và điểm nhận của công việc đang hoạt động tại Thanh Hóa.</CardDescription>
      </div>
      <Button size='icon' variant='outline' onClick={focusOperations} aria-label='Căn bản đồ theo dữ liệu vận hành' disabled={!ready}><LocateFixed aria-hidden='true' className='size-4' /></Button>
    </CardHeader>
    <CardContent className='p-0'>
      <div className='flex flex-wrap items-center gap-2 border-y border-border bg-surface-soft/55 px-5 py-3 text-xs text-muted'>
        <span className='inline-flex items-center gap-1.5'><span className='size-2.5 rounded-full bg-primary ring-4 ring-primary/10' /> {drivers.length} tài xế có vị trí</span>
        <span className='text-border'>•</span>
        <span className='inline-flex items-center gap-1.5'><MapPin aria-hidden='true' className='size-3.5 text-primary' /> {trips.length} công việc đang hoạt động</span>
        {trips.some(item => item.incidentOpen) ? <><span className='text-border'>•</span><Badge variant='destructive'><AlertTriangle aria-hidden='true' className='mr-1 size-3' />Có sự cố trên bản đồ</Badge></> : null}
      </div>
      <div className='relative h-[360px] bg-surface-soft md:h-[430px]'>
        <div ref={containerRef} className='absolute inset-0' aria-label='Bản đồ vận hành FlashX' />
        {!ready && !mapError ? <div className='pointer-events-none absolute inset-0 grid place-items-center bg-surface-soft/70 backdrop-blur-sm'><div className='rounded-2xl border border-border bg-surface/90 px-4 py-3 text-sm font-medium text-muted shadow-sm'>Đang tải bản đồ…</div></div> : null}
        {mapError ? <div className='absolute inset-0 grid place-items-center p-6'><div className='max-w-md rounded-2xl border border-warning/15 bg-warning-soft p-4 text-center text-sm leading-6 text-foreground'><strong>Bản đồ chưa khả dụng.</strong><br />{mapError} Danh sách công việc và tài xế vẫn hoạt động bình thường.</div></div> : null}
      </div>
      <div className='flex flex-wrap items-center justify-between gap-3 border-t border-border px-5 py-3 text-[11px] leading-5 text-muted'><span>Dữ liệu vị trí lấy từ FlashX; nền bản đồ từ OpenFreeMap/OpenStreetMap.</span><span>Click marker để mở chi tiết.</span></div>
    </CardContent>
  </Card>;
}
