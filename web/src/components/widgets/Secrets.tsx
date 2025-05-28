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
import { Accessor, createSignal, Index, Setter, Show } from 'solid-js'

type Secret = { name: string; value: string }

type SecretRowProps = { secret: Accessor<Secret>; index: number; setSecrets: Setter<Secret[]> }

const SecretRow = ({ secret, index, setSecrets }: SecretRowProps) => {
  const [inputs, setInputs] = createSignal<Secret>(secret())
  const [isEditing, setIsEditing] = createSignal(false)
  const [toggleVisible, setToggleVisible] = createSignal(false)

  const updateSecret = () => {
    setSecrets((secrets) => secrets.map((s, i) => (i === index ? inputs() : s)))
    setIsEditing(false)
  }

  const deleteSecret = () => setSecrets((secrets) => secrets.filter((_, i) => i !== index))

  return (
    <TableRow>
      <TableCell>
        <Show when={isEditing()} fallback={secret().name}>
          <TextField
            value={inputs().name}
            onChange={(name) => setInputs((inputs) => ({ ...inputs, name }))}
          >
            <TextFieldInput />
          </TextField>
        </Show>
      </TableCell>
      <TableCell class="flex">
        <TextField
          value={inputs().value}
          onChange={(value) => setInputs((inputs) => ({ ...inputs, value }))}
          readOnly={!isEditing()}
          class="flex-1"
        >
          <TextFieldInput type={toggleVisible() ? 'text' : 'password'} />
        </TextField>
        <Button onClick={() => setToggleVisible(!toggleVisible())}>Toggle</Button>
      </TableCell>
      <TableCell>
        <div class="flex">
          <Show
            when={isEditing()}
            fallback={<Button onClick={() => setIsEditing(true)}>Edit</Button>}
          >
            <Button onClick={updateSecret}>Save</Button>
          </Show>
          <Button onClick={deleteSecret}>Delete</Button>
        </div>
      </TableCell>
    </TableRow>
  )
}

const AddSecretRow = ({ setSecrets }: { setSecrets: Setter<Secret[]> }) => {
  const [toggleVisible, setToggleVisible] = createSignal(false)
  const [inputs, setInputs] = createSignal<Secret>({ name: '', value: '' })

  const addSecret = () => {
    setSecrets((secrets) => [...secrets, inputs()])
    setInputs({ name: '', value: '' })
  }

  return (
    <TableRow>
      <TableCell>
        <TextField
          value={inputs().name}
          onChange={(name) => setInputs((inputs) => ({ ...inputs, name }))}
        >
          <TextFieldInput placeholder="SECRET_NAME" />
        </TextField>
      </TableCell>
      <TableCell class="flex">
        <TextField
          value={inputs().value}
          onChange={(value) => setInputs((inputs) => ({ ...inputs, value }))}
          class="flex-1"
        >
          <TextFieldInput placeholder="Secret value" type={toggleVisible() ? 'text' : 'password'} />
        </TextField>
        <Button onClick={() => setToggleVisible(!toggleVisible())}>Toggle</Button>
      </TableCell>
      <TableCell>
        <Button disabled={!inputs().name || !inputs().value} onClick={addSecret}>
          Add
        </Button>
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
        <Index each={secrets()}>
          {(secret, index) => <SecretRow {...{ secret, index, setSecrets }} />}
        </Index>
        <AddSecretRow setSecrets={setSecrets} />
      </TableBody>
    </Table>
  )
}

export default Secrets
