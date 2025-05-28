import { Eye, EyeOff } from '@/components/icons'
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
import { Accessor, createSignal, Index, JSX, Setter, Show } from 'solid-js'

type Secret = { name: string; value: string }

type SecretRowProps = { secret: Accessor<Secret>; index: number; setSecrets: Setter<Secret[]> }

type SecretTableRowProps = {
  isEditing: Accessor<boolean>
  inputs: Accessor<Secret>
  setInputs: Setter<Secret>
  secret?: Accessor<Secret>
  children: JSX.Element
  type: 'add' | 'edit'
}

const SecretTableRow = ({
  isEditing,
  inputs,
  setInputs,
  secret,
  children,
  type,
}: SecretTableRowProps) => {
  const [toggleVisible, setToggleVisible] = createSignal(false)

  return (
    <TableRow>
      <TableCell>
        <Show when={isEditing()} fallback={secret && secret().name}>
          <TextField
            value={inputs().name}
            onChange={(name) => setInputs((inputs) => ({ ...inputs, name }))}
          >
            <TextFieldInput placeholder="SECRET_NAME" />
          </TextField>
        </Show>
      </TableCell>
      <TableCell class="flex space-x-2">
        <TextField
          value={inputs().value}
          onChange={(value) => setInputs((inputs) => ({ ...inputs, value }))}
          readOnly={!isEditing()}
          class="flex-1"
        >
          <TextFieldInput
            placeholder="Secret value"
            type={toggleVisible() || isEditing() ? 'text' : 'password'}
          />
        </TextField>
        <Show when={type === 'edit'}>
          <Button onClick={() => setToggleVisible(!toggleVisible())}>
            <Show when={toggleVisible()} fallback={<Eye />}>
              <EyeOff />
            </Show>
          </Button>
        </Show>
      </TableCell>
      <TableCell>{children}</TableCell>
    </TableRow>
  )
}

const SecretRow = ({ secret, index, setSecrets }: SecretRowProps) => {
  const [inputs, setInputs] = createSignal<Secret>(secret())
  const [isEditing, setIsEditing] = createSignal(false)

  const updateSecret = () => {
    setSecrets((secrets) => secrets.map((s, i) => (i === index ? inputs() : s)))
    setIsEditing(false)
  }

  const deleteSecret = () => setSecrets((secrets) => secrets.filter((_, i) => i !== index))

  return (
    <SecretTableRow {...{ isEditing, inputs, setInputs, secret }} type="edit">
      <div class="flex space-x-2">
        <Show
          when={isEditing()}
          fallback={<Button onClick={() => setIsEditing(true)}>Edit</Button>}
        >
          <Button onClick={updateSecret}>Save</Button>
        </Show>
        <Button onClick={deleteSecret}>Delete</Button>
      </div>
    </SecretTableRow>
  )
}

const AddSecretRow = ({ setSecrets }: { setSecrets: Setter<Secret[]> }) => {
  const [inputs, setInputs] = createSignal<Secret>({ name: '', value: '' })

  const addSecret = () => {
    setSecrets((secrets) => [...secrets, inputs()])
    setInputs({ name: '', value: '' })
  }

  return (
    <SecretTableRow isEditing={() => true} inputs={inputs} setInputs={setInputs} type="add">
      <Button disabled={!inputs().name || !inputs().value} onClick={addSecret}>
        Add
      </Button>
    </SecretTableRow>
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
