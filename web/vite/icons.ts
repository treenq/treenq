import { promises as fs } from 'fs'
import path from 'path'

export default function IconSpritePlugin() {
  async function generateIconSprite() {
    try {
      // Read the SVG files in the static/icons folder
      const iconsDir = path.join(process.cwd(), 'public/static', 'icons')
      const files = await fs.readdir(iconsDir)
      let symbols = ''
      const iconNames: string[] = []

      // Build up the SVG sprite from the SVG files
      for (const file of files) {
        if (!file.endsWith('.svg')) continue
        let svgContent = await fs.readFile(path.join(iconsDir, file), 'utf8')
        const id = file.replace('.svg', '')
        iconNames.push(id)
        svgContent = svgContent
          .replace(/id="[^"]+"/g, '') // Remove any existing id
          .replace('<svg', `<symbol id="${id}"`) // Change <svg> to <symbol>
          .replace('</svg>', '</symbol>')
        symbols += svgContent + '\n'
      }

      // Write the SVG sprite to a file in the static folder
      const sprite = `<svg width="0" height="0" style="display: none">\n\n${symbols}</svg>`
      await fs.writeFile(path.join(process.cwd(), 'public/static', 'icon-sprite.svg'), sprite)

      // generate ts file for icon names
      const dtsContent = `export type IconName = ${iconNames.map((name) => `'${name}'`).join(' | ')};\n`
      const dtsFilePath = path.join(process.cwd(), 'src/components/icons', 'icon-names.ts')
      await fs.writeFile(dtsFilePath, dtsContent)
    } catch (error) {
      console.error('Error generating icon sprite:', error)
    }
  }

  return {
    name: 'icon-sprite-plugin',
    buildStart() {
      // Generate during build
      return generateIconSprite()
    },
    configureServer(server: any) {
      // Regenerate during development whenever an icon is added
      const iconsPath = path.join(process.cwd(), 'public/static/icons')
      server.watcher.add(iconsPath)
      server.watcher.on('change', async (changedPath: any) => {
        if (changedPath.includes('public/static/icons') && changedPath.endsWith('.svg')) {
          await generateIconSprite()
        }
      })
    },
  }
}
