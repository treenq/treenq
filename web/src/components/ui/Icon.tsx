interface PropsIson {
  name: string
  customClass?: string
  width?: string
  height?: string
}

export const Icon = (props: PropsIson) => {
  return (
    <svg class={props.customClass} height={props.height} width={props.width}>
      <use href={`/public/static/icon-sprite.svg#${props.name}`} />
    </svg>
  )
}
