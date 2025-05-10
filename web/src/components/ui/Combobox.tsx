import type { JSX, ValidComponent } from 'solid-js'
import { Show, splitProps } from 'solid-js'

import * as ComboboxPrimitive from '@kobalte/core/combobox'
import type { PolymorphicProps } from '@kobalte/core/polymorphic'

import { cn } from '@/components/ui/utils'

const Combobox = ComboboxPrimitive.Root
const ComboboxItemLabel = ComboboxPrimitive.ItemLabel
const ComboboxHiddenSelect = ComboboxPrimitive.HiddenSelect

type ComboboxItemProps<T extends ValidComponent = 'li'> = ComboboxPrimitive.ComboboxItemProps<T> & {
  class?: string | undefined
}

const ComboboxItem = <T extends ValidComponent = 'li'>(
  props: PolymorphicProps<T, ComboboxItemProps<T>>,
) => {
  const [local, others] = splitProps(props as ComboboxItemProps, ['class'])
  return (
    <ComboboxPrimitive.Item
      class={cn(
        'data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground relative flex cursor-default select-none items-center justify-between rounded-sm px-2 py-1.5 text-sm outline-none data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
        local.class,
      )}
      {...others}
    />
  )
}

type ComboboxItemIndicatorProps<T extends ValidComponent = 'div'> =
  ComboboxPrimitive.ComboboxItemIndicatorProps<T> & {
    children?: JSX.Element
  }

const ComboboxItemIndicator = <T extends ValidComponent = 'div'>(
  props: PolymorphicProps<T, ComboboxItemIndicatorProps<T>>,
) => {
  const [local, others] = splitProps(props as ComboboxItemIndicatorProps, ['children'])
  return (
    <ComboboxPrimitive.ItemIndicator {...others}>
      <Show
        when={local.children}
        fallback={
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="size-4"
          >
            <path d="M5 12l5 5l10 -10" />
          </svg>
        }
      >
        {(children) => children()}
      </Show>
    </ComboboxPrimitive.ItemIndicator>
  )
}

type ComboboxSectionProps<T extends ValidComponent = 'li'> =
  ComboboxPrimitive.ComboboxSectionProps<T> & { class?: string | undefined }

const ComboboxSection = <T extends ValidComponent = 'li'>(
  props: PolymorphicProps<T, ComboboxSectionProps<T>>,
) => {
  const [local, others] = splitProps(props as ComboboxSectionProps, ['class'])
  return (
    <ComboboxPrimitive.Section
      class={cn(
        'text-muted-foreground overflow-hidden p-1 px-2 py-1.5 text-xs font-medium',
        local.class,
      )}
      {...others}
    />
  )
}

type ComboboxControlProps<
  U,
  T extends ValidComponent = 'div',
> = ComboboxPrimitive.ComboboxControlProps<U, T> & {
  class?: string | undefined
}

const ComboboxControl = <T, U extends ValidComponent = 'div'>(
  props: PolymorphicProps<U, ComboboxControlProps<T>>,
) => {
  const [local, others] = splitProps(props as ComboboxControlProps<T>, ['class'])
  return (
    <ComboboxPrimitive.Control
      class={cn('flex h-10 items-center rounded-md border px-3', local.class)}
      {...others}
    />
  )
}

type ComboboxInputProps<T extends ValidComponent = 'input'> =
  ComboboxPrimitive.ComboboxInputProps<T> & { class?: string | undefined }

const ComboboxInput = <T extends ValidComponent = 'input'>(
  props: PolymorphicProps<T, ComboboxInputProps<T>>,
) => {
  const [local, others] = splitProps(props as ComboboxInputProps, ['class'])
  return (
    <ComboboxPrimitive.Input
      class={cn(
        'placeholder:text-muted-foreground flex size-full rounded-md bg-transparent py-3 text-sm outline-none disabled:cursor-not-allowed disabled:opacity-50',
        local.class,
      )}
      {...others}
    />
  )
}

type ComboboxTriggerProps<T extends ValidComponent = 'button'> =
  ComboboxPrimitive.ComboboxTriggerProps<T> & {
    class?: string | undefined
    children?: JSX.Element
  }

const ComboboxTrigger = <T extends ValidComponent = 'button'>(
  props: PolymorphicProps<T, ComboboxTriggerProps<T>>,
) => {
  const [local, others] = splitProps(props as ComboboxTriggerProps, ['class', 'children'])
  return (
    <ComboboxPrimitive.Trigger class={cn('size-4 opacity-50', local.class)} {...others}>
      <ComboboxPrimitive.Icon>
        <Show
          when={local.children}
          fallback={
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="size-4"
            >
              <path d="M8 9l4 -4l4 4" />
              <path d="M16 15l-4 4l-4 -4" />
            </svg>
          }
        >
          {(children) => children()}
        </Show>
      </ComboboxPrimitive.Icon>
    </ComboboxPrimitive.Trigger>
  )
}

type ComboboxContentProps<T extends ValidComponent = 'div'> =
  ComboboxPrimitive.ComboboxContentProps<T> & { class?: string | undefined }

const ComboboxContent = <T extends ValidComponent = 'div'>(
  props: PolymorphicProps<T, ComboboxContentProps<T>>,
) => {
  const [local, others] = splitProps(props as ComboboxContentProps, ['class'])
  return (
    <ComboboxPrimitive.Portal>
      <ComboboxPrimitive.Content
        class={cn(
          'bg-popover text-popover-foreground animate-in fade-in-80 relative z-50 min-w-32 overflow-hidden rounded-md border shadow-md',
          local.class,
        )}
        {...others}
      >
        <ComboboxPrimitive.Listbox class="m-0 p-1" />
      </ComboboxPrimitive.Content>
    </ComboboxPrimitive.Portal>
  )
}

export {
  Combobox,
  ComboboxContent,
  ComboboxControl,
  ComboboxHiddenSelect,
  ComboboxInput,
  ComboboxItem,
  ComboboxItemIndicator,
  ComboboxItemLabel,
  ComboboxSection,
  ComboboxTrigger,
}
