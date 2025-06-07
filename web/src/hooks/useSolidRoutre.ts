import { redirect, useLocation, useNavigate, useParams } from '@solidjs/router'

export const useSolidRoute = <T = unknown>() => {
  const params = useParams()
  const location = useLocation<T>()
  const navigate = useNavigate()

  return {
    params,
    location,
    navigate,
    stateRoute: location.state as T,
    backPage: () => navigate(-1),
    redirect,
  }
}
