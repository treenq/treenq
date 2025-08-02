import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type Workspace = { id: string; name: string }
type WorkspaceState = { workspaces: Workspace[] }

const newDefaultWorkspaceState = (): WorkspaceState => ({ workspaces: [] })

function createWorkspaceStore() {
  const [store, setStore] = createStore(newDefaultWorkspaceState())

  return mergeProps(store, {
    addWorkspace: (workspaceName: string) =>
      setStore({
        workspaces: [
          ...store.workspaces,
          { id: String(store.workspaces.length), name: workspaceName },
        ],
      }),
  })
}

export const workspaceStore = createWorkspaceStore()
