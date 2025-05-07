import { onMount } from 'solid-js/types/server/reactive.js'

const RedirectPage = () => {
  onMount(() => window.close())

  return (
    <div>
      <h1>Redirecting...</h1>
    </div>
  )
}

export default RedirectPage
