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
import { httpClient } from '@/services/client'
import { Accessor, createSignal, Index, JSX, onMount, Setter, Show } from 'solid-js'

type Secret = { key: string; value: string }

type SecretRowProps = {
  repoID: string
  secret: Accessor<Secret>
  index: number
  setSecrets: Setter<Secret[]>
}

type SecretsProps = { repoID: string }

type AddSecretRowProps = { setSecrets: Setter<Secret[]>; repoID: string }

type SecretTableRowProps = {
  isEditing: Accessor<boolean>
  inputs: Accessor<Secret>
  setInputs: Setter<Secret>
  secret?: Accessor<Secret>
  revealSecret: () => Promise<boolean>
  children: JSX.Element
  type: 'add' | 'edit'
}

const SecretTableRow = ({
  isEditing,
  inputs,
  setInputs,
  secret,
  revealSecret,
  children,
  type,
}: SecretTableRowProps) => {
  const [visible, setVisible] = createSignal(false)

  const toggleVisible = async () => {
    const revealed = await revealSecret()
    setVisible(revealed)
  }

  return (
    <TableRow>
      <TableCell>
        <Show when={isEditing()} fallback={secret && secret().key}>
          <TextField
            value={inputs().key}
            onChange={(key) => setInputs((inputs) => ({ ...inputs, key }))}
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
            type={visible() || isEditing() ? 'text' : 'password'}
          />
        </TextField>
        <Show when={type === 'edit'}>
          <Button onClick={() => (visible() ? setVisible(false) : toggleVisible())}>
            <Show when={visible()} fallback={<Eye />}>
              <EyeOff />
            </Show>
          </Button>
        </Show>
      </TableCell>
      <TableCell>{children}</TableCell>
    </TableRow>
  )
}

const SecretRow = ({ repoID, secret, index, setSecrets }: SecretRowProps) => {
  const [inputs, setInputs] = createSignal<Secret>(secret())
  const [isEditing, setIsEditing] = createSignal(false)

  const revealSecret = async () => {
    const response = await httpClient.revealSecret({ repoID, key: inputs().key })
    if (!('error' in response)) {
      setInputs((inputs) => ({ ...inputs, value: response.data.value }))
      return true
    }
    return false
  }

  const toggleEditMode = async () => {
    const revealed = await revealSecret()
    setIsEditing(revealed)
  }

  const updateSecret = async () => {
    const response = await httpClient.setSecret({
      repoID,
      key: inputs().key,
      value: inputs().value,
    })
    if (!('error' in response)) {
      setSecrets((secrets) => secrets.map((s, i) => (i === index ? inputs() : s)))
      setIsEditing(false)
    }
  }

  const deleteSecret = () => setSecrets((secrets) => secrets.filter((_, i) => i !== index))

  return (
    <SecretTableRow
      {...{
        isEditing,
        inputs,
        setInputs,
        secret,
        revealSecret,
      }}
      type="edit"
    >
      <div class="flex space-x-2">
        <Show when={isEditing()} fallback={<Button onClick={toggleEditMode}>Edit</Button>}>
          <Button onClick={updateSecret}>Save</Button>
        </Show>
        <Button onClick={deleteSecret}>Delete</Button>
      </div>
    </SecretTableRow>
  )
}

const AddSecretRow = ({ setSecrets, repoID }: AddSecretRowProps) => {
  const [inputs, setInputs] = createSignal<Secret>({ key: '', value: '' })

  const addSecret = async () => {
    const response = await httpClient.setSecret({
      repoID,
      key: inputs().key,
      value: inputs().value,
    })
    if (!('error' in response)) {
      setSecrets((secrets) => [...secrets, inputs()])
      setInputs({ key: '', value: '' })
    }
  }

  return (
    <SecretTableRow
      isEditing={() => true}
      revealSecret={async () => false}
      inputs={inputs}
      setInputs={setInputs}
      type="add"
    >
      <Button disabled={!inputs().key || !inputs().value} onClick={addSecret}>
        Add
      </Button>
    </SecretTableRow>
  )
}

const Secrets = ({ repoID }: SecretsProps) => {
  const [secrets, setSecrets] = createSignal<Secret[]>([])

  onMount(() => {
    const fetchSecrets = async () => {
      const response = await httpClient.getSecrets({ repoID })
      if (!('error' in response) && response.data.keys) {
        setSecrets(
          response.data.keys.map((key) => ({
            key,
            value: '********',
          })),
        )
      }
    }

    fetchSecrets()
  })

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
          {(secret, index) => (
            <SecretRow
              {...{
                repoID,
                secret,
                index,
                setSecrets,
              }}
            />
          )}
        </Index>
        <AddSecretRow repoID={repoID} setSecrets={setSecrets} />
      </TableBody>
    </Table>
  )
}

export default Secrets
