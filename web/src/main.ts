import { createApp } from 'vue'
import App from './App.vue'
import { router } from './router'
import { applyTheme, readTheme } from './lib/theme'
import './styles/tokens.css'
import './styles/base.css'

// index.html already set the attribute; this also syncs the theme-color metas.
applyTheme(readTheme())

createApp(App).use(router).mount('#app')
