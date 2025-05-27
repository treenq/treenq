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

  return <For each={secrets()}>{(secret) => <div>{secret.name}</div>}</For>
}

export default Secrets
