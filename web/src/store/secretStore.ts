import { httpClient } from '@/services/client'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

export type Secret = { key: string; value: string }

type SecretState = { secrets: Secret[] }

const newDefaultSecretState = (): SecretState => ({ secrets: [] })

function createSecretStore() {
  const [store, setStore] = createStore(newDefaultSecretState())

  return mergeProps(store, {
    addSecret: async (repoID: string, key: string, value: string) => {
      const res = await httpClient.setSecret({ repoID, key, value })
      if ('error' in res) return { success: false }
      setStore({ secrets: [...store.secrets, { key, value }] })
      return { success: true }
    },
    updateSecret: async (repoID: string, key: string, value: string) => {
      const res = await httpClient.setSecret({ repoID, key, value })
      if ('error' in res) return { success: false }
      setStore({
        secrets: store.secrets.map((s) =>
          s.key === key
            ? {
                key,
                value,
              }
            : s,
        ),
      })
      return { success: true }
    },
    getSecrets: async (repoID: string) => {
      const res = await httpClient.getSecrets({ repoID })
      if ('error' in res) return
      const secrets =
        res.data.keys?.map((k) => ({
          key: k,
          value: '******',
        })) || []
      setStore({ secrets })
    },
    revealSecret: async (repoID: string, key: string) => {
      const res = await httpClient.revealSecret({ repoID, key })
      if ('error' in res) return ''
      return res.data.value
    },
    removeSecret: async (repoID: string, key: string) => {
      const res = await httpClient.removeSecret({ repoID, key })
      if ('error' in res) return
      setStore({ secrets: store.secrets.filter((s) => s.key !== key) })
    },
  })
}

export const secretStore = createSecretStore()
