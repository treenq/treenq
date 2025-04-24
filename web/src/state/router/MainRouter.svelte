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
    <ProtectedRouter component={<Main />} satisfies={flag} fallback={Auth} />
  </Route>

  <Route key="auth" path="/auth">
    <Auth />
  </Route>
</Router>
