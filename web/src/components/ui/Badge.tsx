import type { Component, ComponentProps } from 'solid-js'
import { splitProps } from 'solid-js'

import type { VariantProps } from 'class-variance-authority'
import { cva } from 'class-variance-authority'

import { cn } from '@/components/ui/utils'

const badgeVariants = cva(
  'inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-primary ',
        secondary: 'border-transparent bg-secondary text-secondary-foreground',
        outline: 'text-foreground',
        success: 'border-success-foreground bg-success text-success-foreground',
        warning: 'border-warning-foreground bg-warning text-warning-foreground',
        error: 'border-error-foreground bg-error text-error-foreground',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

type BadgeProps = ComponentProps<'div'> & VariantProps<typeof badgeVariants>

const Badge: Component<BadgeProps> = (props) => {
  const [local, others] = splitProps(props, ['class', 'variant'])

  return <div class={cn(badgeVariants({ variant: local.variant }), local.class)} {...others} />
}

export { Badge, badgeVariants }
export type { BadgeProps }
