<script lang="ts">
  import { init } from '@wjfe/n-savant'
  init()

  import ProtectedRouter from './ProtectedRouter.svelte'

  import { Router, Route } from '@wjfe/n-savant'

  import Auth from '@/components/pages/Auth.svelte'
  import Main from '@/components/pages/Main.svelte'

  let flag = $state(false)
</script>

<Router>
  <Route key="main" path="/">
    <ProtectedRouter satisfies={flag}>
      <Main
        onClick={() => {
          flag = !flag
        }}
      />
    </ProtectedRouter>
  </Route>

  <Route key="auth" path="/auth">
    <Auth />
  </Route>
</Router>
