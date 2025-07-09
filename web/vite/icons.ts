import { promises as fs } from 'fs'
import path from 'path'

export default function IconSpritePlugin() {
  async function generateIconSprite() {
    // Read the SVG files in the static/icons folder
    const iconsDir = path.join(process.cwd(), 'public/static', 'icons')
    const files = await fs.readdir(iconsDir)
    let symbols = ''

    // Build up the SVG sprite from the SVG files
    for (const file of files) {
      if (!file.endsWith('.svg')) continue
      let svgContent = await fs.readFile(path.join(iconsDir, file), 'utf8')
      const id = file.replace('.svg', '')
      svgContent = svgContent
        .replace(/id="[^"]+"/, '') // Remove any existing id
        .replace('<svg', `<symbol id="${id}"`) // Change <svg> to <symbol>
        .replace('</svg>', '</symbol>')
      symbols += svgContent + '\n'
    }

    // Write the SVG sprite to a file in the static folder
    const sprite = `<svg width="0" height="0" style="display: none">\n\n${symbols}</svg>`
    await fs.writeFile(path.join(process.cwd(), 'public/static', 'icon-sprite.svg'), sprite)
  }

  return {
    name: 'icon-sprite-plugin',
    buildStart() {
      // Generate during build
      return generateIconSprite()
    },
    configureServer(server: any) {
      // Regenerate during development whenever an icon is added
      server.watcher.add(path.join(process.cwd(), 'static', 'icons', '*.svg'))
      server.watcher.on('change', async (changedPath: any) => {
        if (changedPath.endsWith('.svg')) return generateIconSprite()
      })
    },
  }
}
