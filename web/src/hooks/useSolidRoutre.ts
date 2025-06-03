import { useLocation, useParams } from '@solidjs/router'

export const useSolidRoute = <T = unknown>() => {
  const params = useParams()
  const location = useLocation<T>()

  const id = params.id

  return {
    id,
    location,
    stateRoute: location.state as T, // опционально, если хочешь вернуть только state
  }
}
