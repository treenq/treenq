import { Button } from '@/components/ui/Button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table'
import { createSignal, For } from 'solid-js'

type Secret = {
  name: string
  value: string
}

const Secrets = () => {
  const [secrets] = createSignal<Secret[]>([
    { name: 'API_KEY', value: '123abc' },
    { name: 'DB_PASSWORD', value: 'password123' },
    { name: 'SECRET_TOKEN', value: 'token456' },
  ])

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead class="w-sm">Name</TableHead>
          <TableHead class="w-3xl">Value</TableHead>
          <TableHead class="w-3xs">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <For each={secrets()}>
          {({ name, value }) => (
            <TableRow>
              <TableCell>{name}</TableCell>
              <TableCell>{value}</TableCell>
              <TableCell>
                <Button>Edit</Button>
                <Button>Delete</Button>
              </TableCell>
            </TableRow>
          )}
        </For>
      </TableBody>
    </Table>
  )
}

export default Secrets
