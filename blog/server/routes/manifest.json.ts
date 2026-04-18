import { getSettingGroup } from '~/composables/api/sysconfig'

export default defineEventHandler(async (event) => {
  try {
    const [blogConfig] = await Promise.all([
      getSettingGroup('blog')
    ])

    const processConfig = (config: any, prefix: string) => {
      const processed: Record<string, string> = {}
      Object.entries(config).forEach(([key, value]) => {
        if (key.startsWith(`${prefix}.`)) {
          processed[key.substring(prefix.length + 1)] = value as string
        }
      })
      return processed
    }

    const blog = processConfig(blogConfig, 'blog')

    const manifest = {
      name: blog.title || 'Flec Blog',
      short_name: blog.title?.substring(0, 12) || 'Flec',
      description: blog.description || 'Flec 个人博客',
      theme_color: '#f7f7f7',
      background_color: '#ffffff',
      display: 'standalone',
      start_url: '/',
      icons: [
        {
          src: blog.favicon || '/favicon.ico',
          type: blog.favicon?.endsWith('.png') ? 'image/png' : 'image/x-icon'
        },
        {
          src: blog.favicon || '/favicon.ico',
          type: blog.favicon?.endsWith('.png') ? 'image/png' : 'image/x-icon'
        }
      ]
    }

    setHeader(event, 'Content-Type', 'application/manifest+json')
    return manifest
  } catch (error) {
    return {
      name: 'Flec Blog',
      short_name: 'Flec',
      theme_color: '#f7f7f7',
      background_color: '#ffffff',
      display: 'standalone',
      start_url: '/',
      icons: [{ src: '/favicon.ico', sizes: '192x192', type: 'image/png' }]
    }
  }
})
