import {defineConfig} from 'vitepress'
import { generateSidebar } from 'vitepress-sidebar'

// https://vitepress.dev/reference/site-config
export default defineConfig({
    title: "Wireset",
    description: "a collection of useful wireset for a next project",
    themeConfig: {
        nav: [
            {text: 'Guide', link: '/guide/'},
            {text: 'Wiresets', link: '/wiresets/'}
        ],

        sidebar: generateSidebar({
            documentRootPath: '/docs',
            useTitleFromFileHeading: true,
            capitalizeFirst: true
        }),
        socialLinks: [
            {icon: 'github', link: 'https://github.com/vuejs/vitepress'}
        ]
    }
})
