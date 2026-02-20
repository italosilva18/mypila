import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { PaginationInfo } from '../types';

interface PaginationProps {
  pagination: PaginationInfo;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  pageSizeOptions?: number[];
  showPageSizeSelector?: boolean;
  disabled?: boolean;
}

export function Pagination({
  pagination,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = [10, 25, 50, 100],
  showPageSizeSelector = true,
  disabled = false
}: PaginationProps) {
  const { page, limit, total, totalPages } = pagination;

  const startItem = total === 0 ? 0 : (page - 1) * limit + 1;
  const endItem = Math.min(page * limit, total);

  const getVisiblePages = (): (number | 'ellipsis')[] => {
    const delta = 2;
    const range: (number | 'ellipsis')[] = [];
    const rangeWithDots: (number | 'ellipsis')[] = [];
    let l: number | undefined;

    for (let i = 1; i <= totalPages; i++) {
      if (i === 1 || i === totalPages || (i >= page - delta && i <= page + delta)) {
        range.push(i);
      }
    }

    for (const i of range) {
      if (l !== undefined) {
        if (typeof i === 'number' && i - l === 2) {
          rangeWithDots.push(l + 1);
        } else if (typeof i === 'number' && i - l !== 1) {
          rangeWithDots.push('ellipsis');
        }
      }
      rangeWithDots.push(i);
      l = typeof i === 'number' ? i : l;
    }

    return rangeWithDots;
  };

  if (totalPages <= 1 && !showPageSizeSelector) {
    return null;
  }

  return (
    <div className="flex flex-col sm:flex-row items-center justify-between gap-3 sm:gap-4 px-2 py-3">
      {/* Info and page size selector */}
      <div className="flex flex-col sm:flex-row items-center gap-2 sm:gap-4 text-xs sm:text-sm text-stone-500">
        <span>
          {startItem}-{endItem} de {total}
        </span>
        {showPageSizeSelector && onPageSizeChange && (
          <div className="flex items-center gap-2">
            <span className="hidden sm:inline">Itens:</span>
            <select
              value={limit}
              onChange={(e) => onPageSizeChange(Number(e.target.value))}
              disabled={disabled}
              className="bg-stone-100 border border-stone-200 rounded-lg px-2 py-1.5 text-stone-700 text-xs sm:text-sm focus:outline-none focus:ring-2 focus:ring-stone-400 min-h-[36px] sm:min-h-[32px]"
            >
              {pageSizeOptions.map((size) => (
                <option key={size} value={size}>
                  {size}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>

      {/* Pagination controls */}
      {totalPages > 1 && (
        <div className="flex items-center gap-0.5 sm:gap-1">
          {/* First page */}
          <button
            onClick={() => onPageChange(1)}
            disabled={disabled || page === 1}
            className="p-2.5 sm:p-2 rounded-lg hover:bg-stone-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors text-stone-600 min-w-[44px] min-h-[44px] sm:min-w-[36px] sm:min-h-[36px] flex items-center justify-center"
            title="Primeira pagina"
          >
            <ChevronsLeft className="w-5 h-5 sm:w-4 sm:h-4" />
          </button>

          {/* Previous page */}
          <button
            onClick={() => onPageChange(page - 1)}
            disabled={disabled || page === 1}
            className="p-2.5 sm:p-2 rounded-lg hover:bg-stone-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors text-stone-600 min-w-[44px] min-h-[44px] sm:min-w-[36px] sm:min-h-[36px] flex items-center justify-center"
            title="Pagina anterior"
          >
            <ChevronLeft className="w-5 h-5 sm:w-4 sm:h-4" />
          </button>

          {/* Page numbers */}
          <div className="flex items-center gap-0.5 sm:gap-1 mx-1 sm:mx-2">
            {getVisiblePages().map((pageNum, index) => (
              pageNum === 'ellipsis' ? (
                <span key={`ellipsis-${index}`} className="px-1 sm:px-2 text-stone-400 text-sm">
                  ...
                </span>
              ) : (
                <button
                  key={pageNum}
                  onClick={() => onPageChange(pageNum)}
                  disabled={disabled || pageNum === page}
                  className={`min-w-[40px] min-h-[40px] sm:min-w-[36px] sm:min-h-[36px] px-2 rounded-lg text-xs sm:text-sm font-medium transition-colors flex items-center justify-center ${
                    pageNum === page
                      ? 'bg-stone-800 text-white'
                      : 'hover:bg-stone-100 text-stone-600'
                  } disabled:cursor-default`}
                >
                  {pageNum}
                </button>
              )
            ))}
          </div>

          {/* Next page */}
          <button
            onClick={() => onPageChange(page + 1)}
            disabled={disabled || page === totalPages}
            className="p-2.5 sm:p-2 rounded-lg hover:bg-stone-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors text-stone-600 min-w-[44px] min-h-[44px] sm:min-w-[36px] sm:min-h-[36px] flex items-center justify-center"
            title="Proxima pagina"
          >
            <ChevronRight className="w-5 h-5 sm:w-4 sm:h-4" />
          </button>

          {/* Last page */}
          <button
            onClick={() => onPageChange(totalPages)}
            disabled={disabled || page === totalPages}
            className="p-2.5 sm:p-2 rounded-lg hover:bg-stone-100 disabled:opacity-40 disabled:cursor-not-allowed transition-colors text-stone-600 min-w-[44px] min-h-[44px] sm:min-w-[36px] sm:min-h-[36px] flex items-center justify-center"
            title="Ultima pagina"
          >
            <ChevronsRight className="w-5 h-5 sm:w-4 sm:h-4" />
          </button>
        </div>
      )}
    </div>
  );
}
