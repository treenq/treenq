import { useLocation, useNavigate, useParams } from '@solidjs/router'

export const useSolidRoute = <ParamKeys extends string, T = unknown>(
  paramKeys: readonly ParamKeys[] = [],
) => {
  const routeParams = useParams()
  const location = useLocation<T>()
  const navigate = useNavigate()

  const params = {} as Record<ParamKeys, string>
  for (const key of paramKeys) {
    params[key] = routeParams[key]
  }

  return {
    params,
    location,
    navigate,
    stateRoute: location.state as T,
  }
}
