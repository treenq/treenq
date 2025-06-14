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
import { Accessor, createEffect, createSignal, Index, JSX, Setter, Show } from 'solid-js'

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
  visible: Accessor<boolean>
  children: JSX.Element
}

const secretButtonSizeClass = 'w-16'

const SecretTableRow = ({
  isEditing,
  inputs,
  setInputs,
  secret,
  children,
  visible,
}: SecretTableRowProps) => {
  return (
    <TableRow>
      <TableCell>
        <Show when={isEditing()} fallback={secret && secret().key}>
          <TextField
            value={inputs().key}
            onChange={(key) => setInputs((inputs) => ({ ...inputs, key }))}
          >
            <TextFieldInput placeholder="input name" />
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
            placeholder="and value here"
            type={visible() || isEditing() ? 'text' : 'password'}
          />
        </TextField>
      </TableCell>
      <TableCell>{children}</TableCell>
    </TableRow>
  )
}

const SecretRow = ({ repoID, secret, index, setSecrets }: SecretRowProps) => {
  const [inputs, setInputs] = createSignal<Secret>(secret())
  const [isEditing, setIsEditing] = createSignal(false)
  const [visible, setVisible] = createSignal(false)

  const toggleVisible = async () => {
    const revealed = await revealSecret()
    setVisible(revealed)
  }

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

  const deleteSecret = async () => {
    const res = await httpClient.removeSecret({ repoID, key: secret().key })
    if (!('error' in res)) {
      setSecrets((secrets) => secrets.filter((_, i) => i !== index))
    }
  }

  return (
    <SecretTableRow
      {...{
        isEditing,
        inputs,
        setInputs,
        secret,
        revealSecret,
        visible,
      }}
    >
      <div class="flex space-x-2">
        <Show
          when={isEditing()}
          fallback={
            <Button class={secretButtonSizeClass} variant="outline" onClick={toggleEditMode}>
              Edit
            </Button>
          }
        >
          <Button class={secretButtonSizeClass} variant="default" onClick={updateSecret}>
            Save
          </Button>
        </Show>
        <Button
          variant="outline"
          size="icon"
          onClick={() => (visible() ? setVisible(false) : toggleVisible())}
        >
          <Show when={visible()} fallback={<Eye />}>
            <EyeOff />
          </Show>
        </Button>
        <Button variant="destructive" onClick={deleteSecret}>
          Delete
        </Button>
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
      inputs={inputs}
      setInputs={setInputs}
      visible={() => true}
    >
      <Button
        class={secretButtonSizeClass}
        disabled={!inputs().key || !inputs().value}
        onClick={addSecret}
      >
        Add
      </Button>
    </SecretTableRow>
  )
}

const Secrets = ({ repoID }: SecretsProps) => {
  const [secrets, setSecrets] = createSignal<Secret[]>([])

  createEffect(() => {
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
          <TableHead class="w-3xs"></TableHead>
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
