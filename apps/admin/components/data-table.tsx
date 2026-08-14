'use client';

import { useMemo, useState, type KeyboardEvent, type ReactNode } from 'react';
import { tableFeatures, useTable, type ColumnDef } from '@tanstack/react-table';
import { Search } from 'lucide-react';
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
};

export function DataTable<TData extends Record<string, unknown>>({
  columns,
  data,
  searchPlaceholder = 'Tìm kiếm...',
  loading = false,
  onRowClick,
  mobileRow,
}: DataTableProps<TData>) {
  const [query, setQuery] = useState('');
  const filteredData = useMemo(() => {
    const normalized = query.trim().toLocaleLowerCase('vi');
    if (!normalized) return data;
    return data.filter(row => JSON.stringify(row).toLocaleLowerCase('vi').includes(normalized));
  }, [data, query]);

  const table = useTable({ features: dataTableFeatures, data: filteredData, columns });
  const noRowsText = query.trim() ? 'Không có kết quả phù hợp.' : 'Chưa có dữ liệu.';

  function activateRow(event: KeyboardEvent<HTMLTableRowElement>, row: TData) {
    if (!onRowClick || (event.key !== 'Enter' && event.key !== ' ')) return;
    event.preventDefault();
    onRowClick(row);
  }

  return (
    <div className='flex flex-col gap-3'>
      <div className='relative max-w-sm'>
        <Search aria-hidden='true' className='pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-soft' />
        <Input
          aria-label={searchPlaceholder}
          className='pl-9'
          value={query}
          onChange={event => setQuery(event.target.value)}
          placeholder={searchPlaceholder}
        />
      </div>

      {mobileRow ? (
        <div className='grid gap-3 md:hidden'>
          {loading ? Array.from({ length: 4 }, (_, index) => (
            <div key={index} className='h-28 animate-pulse rounded-2xl border border-border bg-surface-soft motion-reduce:animate-none' />
          )) : filteredData.length ? filteredData.map((row, index) => (
            onRowClick ? (
              <button key={index} type='button' onClick={() => onRowClick(row)} className='rounded-2xl text-left focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-primary/15'>
                {mobileRow(row)}
              </button>
            ) : <div key={index}>{mobileRow(row)}</div>
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
              Array.from({ length: 5 }, (_, rowIndex) => (
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

      <div className='text-xs text-muted' aria-live='polite'>
        {loading ? 'Đang tải dữ liệu…' : `${filteredData.length} bản ghi`}
      </div>
    </div>
  );
}
