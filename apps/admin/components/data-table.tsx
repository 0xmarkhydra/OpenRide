'use client';

import { useState } from 'react';
import { flexRender, getCoreRowModel, getFilteredRowModel, useReactTable, type ColumnDef } from '@tanstack/react-table';
import { Search } from 'lucide-react';
import { Input } from './ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';

export function DataTable<TData, TValue>({ columns, data, searchPlaceholder = 'Tìm kiếm...' }: { columns: ColumnDef<TData, TValue>[]; data: TData[]; searchPlaceholder?: string }) {
  const [globalFilter, setGlobalFilter] = useState('');
  const table = useReactTable({
    data,
    columns,
    state: { globalFilter },
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  return <div className='space-y-3'>
    <div className='relative max-w-sm'>
      <Search className='pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400' />
      <Input className='pl-9' value={globalFilter} onChange={event => setGlobalFilter(event.target.value)} placeholder={searchPlaceholder} />
    </div>
    <div className='overflow-hidden rounded-xl border border-slate-200 bg-white'>
      <Table>
        <TableHeader>{table.getHeaderGroups().map(group => <TableRow key={group.id}>{group.headers.map(header => <TableHead key={header.id}>{header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}</TableHead>)}</TableRow>)}</TableHeader>
        <TableBody>{table.getRowModel().rows.length ? table.getRowModel().rows.map(row => <TableRow key={row.id}>{row.getVisibleCells().map(cell => <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>)}</TableRow>) : <TableRow><TableCell className='h-28 text-center text-slate-500' colSpan={columns.length}>Không có dữ liệu phù hợp.</TableCell></TableRow>}</TableBody>
      </Table>
    </div>
    <div className='text-xs text-slate-500'>{table.getFilteredRowModel().rows.length} bản ghi</div>
  </div>;
}
