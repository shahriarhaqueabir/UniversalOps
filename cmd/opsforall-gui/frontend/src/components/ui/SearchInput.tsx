import { forwardRef } from 'react'
import { Search } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface SearchInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'> {
  /** Visual size variant */
  size?: 'sm' | 'md' | 'lg'
  /** Icon to show inside the input (defaults to Search) */
  icon?: React.ReactNode
  /** Optional label shown above the input */
  label?: string
}

const sizeStyles = {
  sm: {
    wrapper: 'h-9',
    iconSize: 14,
    iconLeft: 'left-3',
    padding: 'pl-9 pr-3',
    text: 'text-xs',
    radius: 'rounded-lg',
  },
  md: {
    wrapper: 'h-10',
    iconSize: 16,
    iconLeft: 'left-3',
    padding: 'pl-9 pr-4',
    text: 'text-sm',
    radius: 'rounded-xl',
  },
  lg: {
    wrapper: 'h-12',
    iconSize: 18,
    iconLeft: 'left-4',
    padding: 'pl-11 pr-5',
    text: 'text-sm',
    radius: 'rounded-xl',
  },
} as const

/**
 * SearchInput — standardized search field with left-aligned icon,
 * consistent spacing, and focus-accent transition.
 *
 * The icon is always *outside* the text padding area so it never
 * overlaps user-typed content.
 */
export const SearchInput = forwardRef<HTMLInputElement, SearchInputProps>(
  ({ size = 'md', icon, label, className, ...props }, ref) => {
    const s = sizeStyles[size]

    return (
      <div className={cn('flex flex-col gap-1.5', label && 'w-full')}>
        {label && (
          <span className="text-[10px] font-bold uppercase tracking-[0.2em] text-[var(--color-text-faint)]">
            {label}
          </span>
        )}
        <div className={cn('relative group flex-1', s.wrapper)}>
          {/* Icon — positioned inside padding, never overlaps text */}
          <span className={cn(
            'absolute top-1/2 -translate-y-1/2 text-[var(--color-text-faint)]',
            'group-focus-within:text-[var(--color-accent)] transition-colors duration-200 pointer-events-none',
            s.iconLeft,
          )}>
            {icon || <Search size={s.iconSize} />}
          </span>

          <input
            ref={ref}
            className={cn(
              'w-full h-full bg-[var(--color-panel)] border border-[var(--color-border)]',
              s.radius, s.padding, s.text,
              'text-[var(--color-text)] placeholder-[var(--color-text-faint)]',
              'font-medium',
              'focus:outline-none focus:border-[var(--color-accent)] focus:ring-1 focus:ring-[var(--color-accent)]/20',
              'transition-all duration-200',
              'placeholder:font-normal',
              props.disabled && 'opacity-50 cursor-not-allowed',
              className,
            )}
            {...props}
          />
        </div>
      </div>
    )
  },
)

SearchInput.displayName = 'SearchInput'
