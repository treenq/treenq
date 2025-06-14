import { A as SolidA, type AnchorProps } from '@solidjs/router'
import type { VariantProps } from 'class-variance-authority'
import { splitProps, type JSX } from 'solid-js'

import { cva } from 'class-variance-authority'

import { cn } from '@/components/ui/utils'

const AVariants = cva('', {
  variants: {
    variant: {
      primary: 'hover:text-primary',
      secondary: 'hover:text-secondary',
      light: 'hover:text-secondary-foreground',
    },
  },
  defaultVariants: {
    variant: 'primary',
  },
})

type AProps = AnchorProps &
  VariantProps<typeof AVariants> & { class?: string | undefined; children?: JSX.Element }

function A(props: AProps) {
  const [local, others] = splitProps(props, ['variant', 'class'])
  return <SolidA class={cn(AVariants({ variant: local.variant }), local.class)} {...others} />
}

export { A, AVariants }
export type { AProps }
