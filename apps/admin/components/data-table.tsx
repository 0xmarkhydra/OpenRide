'use client';

import { useEffect, useMemo, useState, type KeyboardEvent, type ReactNode } from 'react';
import { tableFeatures, useTable, type ColumnDef } from '@tanstack/react-table';
import { ChevronLeft, ChevronRight, Search } from 'lucide-react';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './ui/table';

export const dataTableFeatures = tableFeatures({});

type DataTableProps<TData extends Record<string, unknown>> = {
  columns: Array<ColumnDef<typeof dataTableFeatures, TData>>;
  data: TData[];
  searchPlaceholder?: string;
  loading?: boolean;
  onRowClick?: (row: TData) => void;
  mobileRow?: (row: TData) => ReactNode;
  searchText?: (row: TData) => string;
  defaultPageSize?: number;
  pageSizeOptions?: number[];
};

export function DataTable<TData extends Record<string, unknown>>({
  columns,
  data,
  searchPlaceholder = 'Tìm kiếm...',
  loading = false,
  onRowClick,
  mobileRow,
  searchText,
  defaultPageSize = 20,
  pageSizeOptions = [20, 50, 100],
}: DataTableProps<TData>) {
  const [query, setQuery] = useState('');
  const [pageIndex, setPageIndex] = useState(0);
  const [pageSize, setPageSize] = useState(defaultPageSize);

  const filteredData = useMemo(() => {
    const normalized = query.trim().toLocaleLowerCase('vi');
    if (!normalized) return data;
    return data.filter(row => (searchText ? searchText(row) : JSON.stringify(row)).toLocaleLowerCase('vi').includes(normalized));
  }, [data, query, searchText]);

  const pageCount = Math.max(1, Math.ceil(filteredData.length / pageSize));
  const safePageIndex = Math.min(pageIndex, pageCount - 1);
  const pageData = useMemo(() => {
    const start = safePageIndex * pageSize;
    return filteredData.slice(start, start + pageSize);
  }, [filteredData, pageSize, safePageIndex]);
  const table = useTable({ features: dataTableFeatures, data: pageData, columns });
  const noRowsText = query.trim() ? 'Không có kết quả phù hợp.' : 'Chưa có dữ liệu.';
  const startRecord = filteredData.length ? safePageIndex * pageSize + 1 : 0;
  const endRecord = Math.min((safePageIndex + 1) * pageSize, filteredData.length);

  useEffect(() => {
    setPageIndex(0);
  }, [query, pageSize]);

  useEffect(() => {
    if (pageIndex > pageCount - 1) setPageIndex(Math.max(0, pageCount - 1));
  }, [pageCount, pageIndex]);

  function activateRow(event: KeyboardEvent<HTMLTableRowElement>, row: TData) {
    if (!onRowClick || (event.key !== 'Enter' && event.key !== ' ')) return;
    event.preventDefault();
    onRowClick(row);
  }

  return (
    <div className='flex flex-col gap-3'>
      <div className='flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
        <div className='relative w-full max-w-sm'>
          <Search aria-hidden='true' className='pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-soft' />
          <Input
            aria-label={searchPlaceholder}
            className='pl-9'
            value={query}
            onChange={event => setQuery(event.target.value)}
            placeholder={searchPlaceholder}
          />
        </div>
        <label className='flex items-center gap-2 text-xs font-medium text-muted'>
          Hiển thị
          <select
            aria-label='Số bản ghi mỗi trang'
            className='min-h-10 rounded-xl border border-border bg-surface px-3 text-sm font-semibold text-foreground outline-none focus:border-primary focus:ring-2 focus:ring-primary/15'
            value={pageSize}
            onChange={event => setPageSize(Number(event.target.value))}
          >
            {pageSizeOptions.map(option => <option key={option} value={option}>{option}</option>)}
          </select>
        </label>
      </div>

      {mobileRow ? (
        <div className='grid gap-3 md:hidden'>
          {loading ? Array.from({ length: 4 }, (_, index) => (
            <div key={index} className='h-28 animate-pulse rounded-2xl border border-border bg-surface-soft motion-reduce:animate-none' />
          )) : pageData.length ? pageData.map((row, index) => (
            onRowClick ? (
              <button key={`${safePageIndex}-${index}`} type='button' onClick={() => onRowClick(row)} className='rounded-2xl text-left focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-primary/15'>
                {mobileRow(row)}
              </button>
            ) : <div key={`${safePageIndex}-${index}`}>{mobileRow(row)}</div>
          )) : <div className='rounded-2xl border border-dashed border-border bg-surface-soft/60 p-7 text-center text-sm text-muted'>{noRowsText}</div>}
        </div>
      ) : null}

      <div className={mobileRow ? 'hidden overflow-hidden rounded-2xl border border-border bg-surface md:block' : 'overflow-hidden rounded-2xl border border-border bg-surface'}>
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map(group => (
              <TableRow key={group.id}>
                {group.headers.map(header => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder ? null : <table.FlexRender header={header} />}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: Math.min(pageSize, 5) }, (_, rowIndex) => (
                <TableRow key={`loading-${rowIndex}`}>
                  {columns.map((_, columnIndex) => (
                    <TableCell key={`loading-${rowIndex}-${columnIndex}`}>
                      <div className='h-4 w-full max-w-36 animate-pulse rounded-md bg-surface-soft motion-reduce:animate-none' />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : table.getRowModel().rows.length ? (
              table.getRowModel().rows.map(row => (
                <TableRow
                  key={row.id}
                  className={onRowClick ? 'cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20' : undefined}
                  onClick={onRowClick ? () => onRowClick(row.original) : undefined}
                  onKeyDown={event => activateRow(event, row.original)}
                  tabIndex={onRowClick ? 0 : undefined}
                  aria-label={onRowClick ? 'Mở chi tiết' : undefined}
                >
                  {row.getAllCells().map(cell => (
                    <TableCell key={cell.id}>
                      <table.FlexRender cell={cell} />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell className='h-28 text-center text-muted' colSpan={columns.length}>
                  {noRowsText}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className='flex flex-col gap-2 text-xs text-muted sm:flex-row sm:items-center sm:justify-between' aria-live='polite'>
        <span>{loading ? 'Đang tải dữ liệu…' : filteredData.length ? `${startRecord}–${endRecord} / ${filteredData.length} bản ghi` : '0 bản ghi'}</span>
        <div className='flex items-center gap-2'>
          <span>Trang {safePageIndex + 1}/{pageCount}</span>
          <Button size='icon' variant='outline' className='size-9' disabled={loading || safePageIndex === 0} onClick={() => setPageIndex(index => Math.max(0, index - 1))} aria-label='Trang trước'><ChevronLeft aria-hidden='true' className='size-4' /></Button>
          <Button size='icon' variant='outline' className='size-9' disabled={loading || safePageIndex >= pageCount - 1} onClick={() => setPageIndex(index => Math.min(pageCount - 1, index + 1))} aria-label='Trang sau'><ChevronRight aria-hidden='true' className='size-4' /></Button>
        </div>
      </div>
    </div>
  );
}
