import Vue from 'vue'
import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'
import mavonEditor from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'
import App from './App.vue'
import './style.css'

Vue.use(ElementUI, { size: 'small' })
Vue.use(mavonEditor)

new Vue({
  render: h => h(App)
}).$mount('#app')
