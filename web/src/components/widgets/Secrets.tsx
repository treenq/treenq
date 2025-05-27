import { Button } from '@/components/ui/Button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/Table'
import { TextField, TextFieldInput } from '@/components/ui/TextField'
import { createSignal, For, Setter } from 'solid-js'

type Secret = { name: string; value: string }

type SecretRowProps = Secret & { index: number; setSecrets: Setter<Secret[]> }

const SecretRow = ({ name, value, index, setSecrets }: SecretRowProps) => {
  const [readOnly, setReadOnly] = createSignal(true)

  const updateSecret = (value: string) =>
    setSecrets((secrets) => secrets.map((s, i) => (i === index ? { ...s, value } : s)))

  return (
    <TableRow>
      <TableCell>{name}</TableCell>
      <TableCell>
        <TextField onChange={updateSecret} readOnly={readOnly()}>
          <TextFieldInput value={value} type="password" />
        </TextField>
      </TableCell>
      <TableCell>
        <Button onClick={() => setReadOnly(false)}>Edit</Button>
        <Button>Delete</Button>
      </TableCell>
    </TableRow>
  )
}

const Secrets = () => {
  const [secrets, setSecrets] = createSignal<Secret[]>([
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
          {(secret, index) => <SecretRow {...secret} index={index()} setSecrets={setSecrets} />}
        </For>
      </TableBody>
    </Table>
  )
}

export default Secrets
