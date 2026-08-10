import { NextRequest, NextResponse } from 'next/server';

export const dynamic = 'force-dynamic';

async function proxy(request: NextRequest, context: { params: Promise<{ path: string[] }> }) {
  const { path } = await context.params;
  const base = (process.env.FLASHX_API_BASE_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');
  const url = new URL(`${base}/${path.join('/')}`);
  request.nextUrl.searchParams.forEach((value, key) => url.searchParams.append(key, value));

  const headers = new Headers();
  const auth = request.headers.get('authorization');
  const contentType = request.headers.get('content-type');
  if (auth) headers.set('authorization', auth);
  if (contentType) headers.set('content-type', contentType);
  headers.set('accept', 'application/json');

  const init: RequestInit = { method: request.method, headers, cache: 'no-store' };
  if (!['GET', 'HEAD'].includes(request.method)) {
    const body = await request.text();
    if (body) init.body = body;
  }

  try {
    const response = await fetch(url, init);
    const body = await response.text();
    return new NextResponse(body, {
      status: response.status,
      headers: { 'content-type': response.headers.get('content-type') || 'application/json; charset=utf-8' },
    });
  } catch {
    return NextResponse.json({ error: { code: 'BACKEND_UNAVAILABLE', message: 'Không thể kết nối FlashX API.' } }, { status: 503 });
  }
}

export const GET = proxy;
export const POST = proxy;
export const PATCH = proxy;
