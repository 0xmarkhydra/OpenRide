'use client';

import { useMemo, useState } from 'react';
import { tableFeatures, useTable } from '@tanstack/react-table';
import { Search } from 'lucide-react';
import { Input } from './ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';

const features = tableFeatures({});

export function DataTable<TData extends Record<string, unknown>>({ columns, data, searchPlaceholder = 'Tìm kiếm...' }: { columns: any[]; data: TData[]; searchPlaceholder?: string }) {
  const [query, setQuery] = useState('');
  const filteredData = useMemo(() => {
    const normalized = query.trim().toLocaleLowerCase('vi');
    if (!normalized) return data;
    return data.filter(row => JSON.stringify(row).toLocaleLowerCase('vi').includes(normalized));
  }, [data, query]);

  const table = useTable({ features, data: filteredData, columns: columns as any });

  return <div className='space-y-3'>
    <div className='relative max-w-sm'>
      <Search className='pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400' />
      <Input className='pl-9' value={query} onChange={event => setQuery(event.target.value)} placeholder={searchPlaceholder} />
    </div>
    <div className='overflow-hidden rounded-xl border border-slate-200 bg-white'>
      <Table>
        <TableHeader>{table.getHeaderGroups().map(group => <TableRow key={group.id}>{group.headers.map(header => <TableHead key={header.id}>{header.isPlaceholder ? null : <table.FlexRender header={header} />}</TableHead>)}</TableRow>)}</TableHeader>
        <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map(row => <TableRow key={row.id}>{row.getAllCells().map(cell => <TableCell key={cell.id}><table.FlexRender cell={cell} /></TableCell>)}</TableRow>) : <TableRow><TableCell className='h-28 text-center text-slate-500' colSpan={columns.length}>Không có dữ liệu phù hợp.</TableCell></TableRow>}</TableBody>
      </Table>
    </div>
    <div className='text-xs text-slate-500'>{filteredData.length} bản ghi</div>
  </div>;
}
