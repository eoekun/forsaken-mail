import { cp, mkdir, readdir, rm } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const scriptDir = dirname(fileURLToPath(import.meta.url))
const webDir = resolve(scriptDir, '..')
const distDir = resolve(webDir, 'dist')
const embedDir = resolve(webDir, '..', 'embed')

await mkdir(embedDir, { recursive: true })

for (const entry of await readdir(embedDir, { withFileTypes: true })) {
  if (entry.name === '.gitkeep') {
    continue
  }

  await rm(join(embedDir, entry.name), { recursive: true, force: true })
}

await cp(distDir, embedDir, { recursive: true })
