/* @refresh reload */
import 'solid-devtools'
import { render } from 'solid-js/web'
import App from './App.tsx'
import './main.css'

const root = document.getElementById('root')

render(() => <App />, root!)

if (import.meta.env.MODE === 'development') {
  import('@stagewise/toolbar').then(({ initToolbar }) => {
    const stagewiseConfig = { plugins: [] }
    initToolbar(stagewiseConfig)
  })
}
