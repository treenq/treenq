import { useNavigate, useParams } from '@solidjs/router'

export const Routes = {
  auth: createRoute('auth'),
  notFound: createRoute('404'),
  deploy: createRoute('deploy', param('id')),
  repos: createRoute('repos', param('id')),
}

type Param<Name extends string = string> = { param: Name }
type Segment = string | Param

type ParamsFrom<T extends readonly Segment[]> = {
  [K in T[number] as K extends Param<infer Name> ? Name : never]: string
}

function param<Name extends string>(name: Name): Param<Name> {
  return { param: name }
}

function createRoute<T extends readonly Segment[]>(...segments: T) {
  const path = '/' + segments.map((s) => (typeof s === 'string' ? s : `:${s.param}`)).join('/')

  const makeHref = (params: ParamsFrom<T>) => {
    return (
      '/' +
      segments
        .map((s) => (typeof s === 'string' ? s : params[s.param as keyof typeof params]))
        .join('/')
    )
  }

  return {
    path,
    makeHref,
    navigate: (useNavigateFunc: typeof useNavigate, params: ParamsFrom<T>) => {
      const path = makeHref(params)
      return useNavigateFunc()(path)
    },
    params: (useParamsFunc: typeof useParams) => useParamsFunc() as ParamsFrom<T>,
  }
}
