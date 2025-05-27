import type { Component, ComponentProps } from 'solid-js'
import { splitProps } from 'solid-js'

import { cn } from '@/components/ui/utils'

const Table: Component<ComponentProps<'table'>> = (props) => {
  const [local, others] = splitProps(props, ['class'])
  return (
    <div class="relative w-full overflow-auto">
      <table class={cn('w-full caption-bottom text-sm', local.class)} {...others} />
    </div>
  )
}

const TableHeader: Component<ComponentProps<'thead'>> = (props) => {
  const [local, others] = splitProps(props, ['class'])
  return <thead class={cn('[&_tr]:border-b', local.class)} {...others} />
}

const TableBody: Component<ComponentProps<'tbody'>> = (props) => {
  const [local, others] = splitProps(props, ['class'])
  return <tbody class={cn('[&_tr:last-child]:border-0', local.class)} {...others} />
}

const TableFooter: Component<ComponentProps<'tfoot'>> = (props) => {
  const [local, others] = splitProps(props, ['class'])
  return (
    <tfoot class={cn('bg-primary text-primary-foreground font-medium', local.class)} {...others} />
  )
}

const TableRow: Component<ComponentProps<'tr'>> = (props) => {
  const [local, others] = splitProps(props, ['class'])
  return (
    <tr
      class={cn(
        'hover:bg-muted/50 data-[state=selected]:bg-muted border-b transition-colors',
        local.class,
      )}
      {...others}
    />
  )
}

const TableHead: Component<ComponentProps<'th'>> = (props) => {
  const [local, others] = splitProps(props, ['class'])
  return (
    <th
      class={cn(
        'text-muted-foreground h-10 px-2 text-left align-middle font-medium [&:has([role=checkbox])]:pr-0',
        local.class,
      )}
      {...others}
    />
  )
}

const TableCell: Component<ComponentProps<'td'>> = (props) => {
  const [local, others] = splitProps(props, ['class'])
  return (
    <td class={cn('p-2 align-middle [&:has([role=checkbox])]:pr-0', local.class)} {...others} />
  )
}

const TableCaption: Component<ComponentProps<'caption'>> = (props) => {
  const [local, others] = splitProps(props, ['class'])
  return <caption class={cn('text-muted-foreground mt-4 text-sm', local.class)} {...others} />
}

export { Table, TableBody, TableCaption, TableCell, TableFooter, TableHead, TableHeader, TableRow }
