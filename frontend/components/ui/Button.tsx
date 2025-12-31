import { ButtonHTMLAttributes, forwardRef } from 'react'
import { cn } from '@/lib/utils'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = 'primary', size = 'md', ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          'relative inline-flex items-center justify-center',
          'font-medium transition-all duration-200',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
          'disabled:pointer-events-none disabled:opacity-50',
          {
            'px-4 py-2 text-sm rounded-lg': size === 'sm',
            'px-6 py-3 text-base rounded-xl': size === 'md',
            'px-8 py-4 text-lg rounded-xl': size === 'lg',
          },
          {
            'bg-gradient-to-r from-primary to-primary/80 text-primary-foreground shadow-lg shadow-primary/20 hover:shadow-xl hover:shadow-primary/30 hover:scale-[1.02] active:scale-[0.98]':
              variant === 'primary',
            'bg-secondary/50 backdrop-blur-sm border border-border text-foreground hover:bg-secondary/70 hover:scale-[1.02] active:scale-[0.98]':
              variant === 'secondary',
            'bg-transparent text-foreground hover:bg-accent/50 hover:scale-[1.02] active:scale-[0.98]':
              variant === 'ghost',
            'bg-destructive/10 text-destructive border border-destructive/20 hover:bg-destructive/20 hover:scale-[1.02] active:scale-[0.98]':
              variant === 'danger',
          },
          className
        )}
        {...props}
      />
    )
  }
)
Button.displayName = 'Button'

