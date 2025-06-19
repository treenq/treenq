import { createSignal, onCleanup } from 'solid-js'
export const useTimer = () => {
  const [time, setTime] = createSignal<{ second: number; minute: number }>({ second: 0, minute: 0 })

  const [timer, setTimer] = createSignal<NodeJS.Timeout | undefined>(undefined)

  const startTimer = () => {
    finishTimer()
    setTime({ second: 0, minute: 0 })

    const timer = setInterval(() => {
      setTime((prev) => {
        if (prev.second >= 60) {
          return {
            minute: prev.minute + 1,
            second: 0,
          }
        }

        return {
          minute: prev.minute,
          second: prev.second + 1,
        }
      })
    }, 1000)

    setTimer(timer)
  }

  const finishTimer = () => {
    const currentTimer = timer()
    if (currentTimer) {
      clearInterval(currentTimer)
      setTimer(undefined)
    }
  }
  onCleanup(() => finishTimer())

  return {
    startTimer,
    finishTimer,
    time,
    isRunningTimer: () => timer() !== undefined,
  }
}
