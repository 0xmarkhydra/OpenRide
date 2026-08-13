import { readFile } from 'node:fs/promises';
import { join } from 'node:path';

export const dynamic = 'force-dynamic';

export async function GET() {
  const lock = await readFile(join(process.cwd(), 'package-lock.json'), 'utf8');
  return new Response(lock, { headers: { 'content-type': 'application/json; charset=utf-8', 'cache-control': 'no-store' } });
}
