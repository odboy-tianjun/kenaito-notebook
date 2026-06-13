<template>
  <div id="app" class="app-container">
    <!-- 未登录：显示登录页面 -->
    <login-view v-if="!isLoggedIn" @login-success="onLoginSuccess" />

    <!-- 已登录：显示主界面 -->
    <el-container v-else style="height: 100vh">
      <el-aside width="280px" class="sidebar">
        <div class="sidebar-header">
          <div class="title-row">
            <el-dropdown trigger="click" @command="handleUserCommand">
              <span class="user-info">
                <i class="el-icon-user-solid"></i>
                {{ email }}
                <i class="el-icon-arrow-down"></i>
              </span>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="logout">
                  <i class="el-icon-switch-button"></i> 退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
          <div class="search-box">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索笔记..."
              prefix-icon="el-icon-search"
              clearable
              @input="handleSearch"
            />
          </div>
        </div>

        <note-tree
          :notes="searchKeyword ? searchResults : noteTree"
          :selected-id="selectedNoteId"
          @select="openNote"
          @delete="deleteNote"
          @context-menu="handleContextMenu"
        />

        <div class="sidebar-footer">
          <el-button
            :type="showTrash ? 'primary' : 'text'"
            icon="el-icon-delete"
            @click="openTrash"
          >
            回收站
          </el-button>
        </div>
      </el-aside>
      
      <el-main class="main-content">
        <!-- 回收站视图 -->
        <trash-view
          v-if="showTrash"
          :trash-notes="trashNotes"
          @restore="restoreNote"
          @permanent-delete="permanentDeleteNote"
          @empty-trash="emptyTrash"
          @close="showTrash = false"
        />

        <!-- 多标签页 -->
        <div v-else-if="tabs.length > 0" class="tabs-container">
          <div class="tabs-header">
            <div
              v-for="tab in tabs"
              :key="tab.id"
              class="tab-item"
              :class="{ 'is-active': activeTabId === tab.id }"
              @click="switchTab(tab.id)"
            >
              <span class="tab-title">{{ tab.title || '未命名' }}</span>
              <span class="tab-close" @click.stop="closeTab(tab.id)">
                <i class="el-icon-close"></i>
              </span>
            </div>
          </div>
          
          <!-- 编辑器内容 -->
          <div v-if="activeTab" class="editor-container">
            <div class="editor-header">
              <el-input
                v-model="activeTab.title"
                class="note-title-input"
                @blur="saveTab"
                placeholder="笔记标题"
              />
              <div class="editor-actions">
                <el-dropdown trigger="click" @command="handleExport">
                  <el-button size="small" icon="el-icon-download">
                    导出 <i class="el-icon-arrow-down el-icon--right"></i>
                  </el-button>
                  <el-dropdown-menu slot="dropdown">
                    <el-dropdown-item command="html">导出 HTML</el-dropdown-item>
                    <el-dropdown-item command="markdown">导出 Markdown</el-dropdown-item>
                  </el-dropdown-menu>
                </el-dropdown>
              </div>
            </div>
            
            <mavon-editor
              ref="md"
              v-model="activeTab.content"
              :toolbars="toolbars"
              :imageUpload="imageUpload"
              @change="onContentChange"
              @imgAdd="handleImgAdd"
              @imgDel="handleImgDel"
              style="height: calc(100vh - 180px)"
            />
          </div>
        </div>
        
        <!-- 空状态 -->
        <div v-else-if="!showTrash" class="empty-state">
          <i class="el-icon-document"></i>
          <p>选择或创建一个笔记开始编写</p>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script>
import NoteTree from './components/NoteTree.vue'
import TrashView from './components/TrashView.vue'
import LoginView from './components/LoginView.vue'

export default {
  name: 'App',
  components: {
    NoteTree,
    TrashView,
    LoginView
  },
  data() {
    return {
      isLoggedIn: false,
      email: '',
      noteTree: [],
      tabs: [],           // 打开的标签页列表
      activeTabId: null,  // 当前激活的标签页 ID
      searchKeyword: '',
      searchResults: [],
      saveTimers: {},     // 每个标签页的保存定时器
      showTrash: false,   // 是否显示回收站
      trashNotes: [],     // 回收站中的笔记
      toolbars: {
        bold: true,
        italic: true,
        header: true,
        underline: true,
        strikethrough: true,
        mark: true,
        quote: true,
        ol: true,
        ul: true,
        link: true,
        imagelink: true,
        code: true,
        table: true,
        fullscreen: true,
        htmlcode: true,
        undo: true,
        redo: true,
        preview: true,
        subfield: true,
        help: true
      },
      imageUpload: true
    }
  },
  computed: {
    selectedNoteId() {
      return this.activeTabId
    },
    activeTab() {
      return this.tabs.find(tab => tab.id === this.activeTabId)
    }
  },
  async mounted() {
    // 检查登录状态
    try {
      const loggedIn = await window.go.main.App.IsLoggedIn()
      if (loggedIn) {
        this.isLoggedIn = true
        await this.loadUserInfo()
        await this.loadNoteTree()
      }
    } catch (err) {
      console.error('Failed to check login status:', err)
    }

    // 监听全局粘贴和拖放事件
    document.addEventListener('paste', this.handlePaste)
    document.addEventListener('dragover', this.handleDragOver, true)
    document.addEventListener('drop', this.handleDrop, true)
  },
  beforeDestroy() {
    document.removeEventListener('paste', this.handlePaste)
    document.removeEventListener('dragover', this.handleDragOver, true)
    document.removeEventListener('drop', this.handleDrop, true)
  },
  methods: {
    async onLoginSuccess() {
      this.isLoggedIn = true
      await this.loadUserInfo()
      await this.loadNoteTree()
    },

    async loadUserInfo() {
      try {
        const user = await window.go.main.App.GetUser()
        this.email = user.email || ''
      } catch (err) {
        console.error('Failed to load user info:', err)
      }
    },

    handleUserCommand(command) {
      if (command === 'logout') {
        this.$confirm('确定要退出登录吗？', '提示', {
          type: 'info',
          confirmButtonText: '退出',
          cancelButtonText: '取消'
        }).then(() => {
          window.go.main.App.Logout()
          this.isLoggedIn = false
          this.email = ''
          this.tabs = []
          this.activeTabId = null
          this.noteTree = []
          this.$message.success('已退出登录')
        }).catch(() => {})
      }
    },

    async loadNoteTree() {
      try {
        this.noteTree = await window.go.main.App.GetNoteTree()
      } catch (err) {
        console.error('Failed to load note tree:', err)
        this.$message.error('加载笔记失败')
      }
    },
    
    // 打开笔记到标签页
    async openNote(note) {
      if (note.isFolder) return
      
      // 检查是否已经在标签页中
      const existingTab = this.tabs.find(tab => tab.id === note.id)
      if (existingTab) {
        this.activeTabId = note.id
        return
      }
      
      try {
        const fullNote = await window.go.main.App.GetNote(note.id)
        this.tabs.push({
          id: fullNote.id,
          title: fullNote.title,
          content: fullNote.content,
          parentId: fullNote.parentId,
          isDirty: false
        })
        this.activeTabId = fullNote.id
      } catch (err) {
        console.error('Failed to load note:', err)
        this.$message.error('加载笔记内容失败')
      }
    },
    
    // 切换标签页
    switchTab(tabId) {
      this.activeTabId = tabId
    },
    
    // 关闭标签页
    closeTab(tabId) {
      const tab = this.tabs.find(t => t.id === tabId)
      if (tab && tab.isDirty) {
        this.$confirm('有未保存的更改，确定要关闭吗？', '提示', {
          type: 'warning'
        }).then(() => {
          this.removeTab(tabId)
        }).catch(() => {})
      } else {
        this.removeTab(tabId)
      }
    },
    
    removeTab(tabId) {
      const index = this.tabs.findIndex(t => t.id === tabId)
      this.tabs.splice(index, 1)
      
      if (this.activeTabId === tabId) {
        // 切换到相邻的标签页
        if (this.tabs.length > 0) {
          const newIndex = Math.min(index, this.tabs.length - 1)
          this.activeTabId = this.tabs[newIndex].id
        } else {
          this.activeTabId = null
        }
      }
      
      // 清理定时器
      if (this.saveTimers[tabId]) {
        clearTimeout(this.saveTimers[tabId])
        delete this.saveTimers[tabId]
      }
    },
    
    async createNoteDirectly(parentId, isFolder) {
      try {
        const title = isFolder ? '未命名文件夹' : '未命名笔记'
        const newNote = await window.go.main.App.CreateNote(parentId, title, isFolder)
        await this.loadNoteTree()

        // 如果是新建笔记，自动打开
        if (!isFolder) {
          await this.openNote(newNote)
        }

        this.$message.success('创建成功')
      } catch (err) {
        console.error('Failed to create note:', err)
        this.$message.error('创建失败')
      }
    },
    
    async deleteNote(note) {
      // 如果标签页中有该笔记，先关闭
      const tabIndex = this.tabs.findIndex(t => t.id === note.id)
      if (tabIndex > -1) {
        this.removeTab(note.id)
      }

      try {
        await window.go.main.App.DeleteNote(note.id)
        await this.loadNoteTree()
        if (this.showTrash) await this.loadTrashNotes()
        this.$message.success('已移到回收站')
      } catch (err) {
        console.error('Failed to delete note:', err)
        this.$message.error('删除失败')
      }
    },
    
    saveTab() {
      if (!this.activeTab) return
      
      if (this.saveTimers[this.activeTabId]) {
        clearTimeout(this.saveTimers[this.activeTabId])
      }
      
      this.saveTimers[this.activeTabId] = setTimeout(async () => {
        try {
          await window.go.main.App.UpdateNote(
            this.activeTab.id,
            this.activeTab.title,
            this.activeTab.content,
            ''
          )
          this.activeTab.isDirty = false
          await this.loadNoteTree()
        } catch (err) {
          console.error('Failed to save note:', err)
        }
      }, 500)
    },
    
    onContentChange() {
      if (this.activeTab) {
        this.activeTab.isDirty = true
      }
      this.saveTab()
    },

    // 粘贴事件处理
    async handlePaste(e) {
      if (!this.activeTab) return

      const items = e.clipboardData?.items
      if (!items) return

      // 检查是否有文件数据
      const files = []
      for (let i = 0; i < items.length; i++) {
        if (items[i].kind === 'file') {
          const file = items[i].getAsFile()
          if (file) files.push(file)
        }
      }

      if (files.length === 0) return

      e.preventDefault()
      e.stopPropagation()

      for (const file of files) {
        const maxSize = 6 * 1024 * 1024 // 6MB
        if (file.size > maxSize) {
          this.$message.warning(`文件 ${file.name} 大小超过 6MB，已跳过`)
          continue
        }

        try {
          let base64Data
          const isImage = /\.(jpg|jpeg|png|gif|webp)$/i.test(file.name)
          if (isImage) {
            base64Data = await this.compressImage(file)
          } else {
            base64Data = await this.fileToBase64(file)
          }
          const fileUrl = await window.go.main.App.SaveFile(base64Data, file.name)
          this.insertFileLink(file.name, fileUrl)
          this.$message.success(`${file.name} 嵌入成功`)
        } catch (err) {
          console.error('Failed to upload file:', err)
          this.$message.error(`${file.name} 上传失败`)
        }
      }
    },

    // 拖放事件处理（全局 capture 阶段）
    handleDragOver(e) {
      if (!this.activeTab) return
      e.preventDefault()
      e.dataTransfer.dropEffect = 'copy'
    },

    async handleDrop(e) {
      if (!this.activeTab) return

      const files = e.dataTransfer?.files
      if (!files || files.length === 0) return

      e.preventDefault()
      e.stopPropagation()

      for (const file of files) {
        const maxSize = 6 * 1024 * 1024 // 6MB
        if (file.size > maxSize) {
          this.$message.warning(`文件 ${file.name} 大小超过 6MB，已跳过`)
          continue
        }

        try {
          let base64Data
          const isImage = /\.(jpg|jpeg|png|gif|webp)$/i.test(file.name)
          if (isImage) {
            base64Data = await this.compressImage(file)
          } else {
            base64Data = await this.fileToBase64(file)
          }
          const fileUrl = await window.go.main.App.SaveFile(base64Data, file.name)
          this.insertFileLink(file.name, fileUrl)
          this.$message.success(`${file.name} 嵌入成功`)
        } catch (err) {
          console.error('Failed to upload file:', err)
          this.$message.error(`${file.name} 上传失败`)
        }
      }
    },

    fileToBase64(file) {
      return new Promise((resolve, reject) => {
        const reader = new FileReader()
        reader.readAsDataURL(file)
        reader.onload = () => resolve(reader.result)
        reader.onerror = error => reject(error)
      })
    },

    insertFileLink(filename, filePath) {
      if (!this.activeTab) return

      // 根据文件类型决定插入格式
      const isImage = /\.(jpg|jpeg|png|gif|webp|svg)$/i.test(filename)
      let markdown = ''
      if (isImage) {
        markdown = `![${filename}](${filePath})`
      } else {
        markdown = `[${filename}](${filePath})`
      }

      // 在内容末尾插入
      this.activeTab.content += `\n\n${markdown}\n`
      this.saveTab()
    },

    async handleSearch() {
      if (!this.searchKeyword.trim()) {
        this.searchResults = []
        return
      }
      
      try {
        this.searchResults = await window.go.main.App.SearchNotes(this.searchKeyword)
      } catch (err) {
        console.error('Failed to search notes:', err)
      }
    },
    
    async handleExport(command) {
      if (!this.activeTab) return

      try {
        let filePath
        if (command === 'html') {
          filePath = await window.go.main.App.ExportHTML(
            this.activeTab.title,
            this.activeTab.content
          )
        } else {
          filePath = await window.go.main.App.ExportMarkdown(
            this.activeTab.title,
            this.activeTab.content
          )
        }

        if (filePath) {
          this.$confirm('导出成功！是否打开文件所在目录？', '导出完成', {
            confirmButtonText: '打开目录',
            cancelButtonText: '确定',
            type: 'success'
          }).then(() => {
            window.go.main.App.RevealInFinder(filePath)
          }).catch(() => {})
        }
      } catch (err) {
        // 用户取消选择不显示错误
        if (err && err.includes && err.includes('用户取消')) {
          return
        }
        console.error('Failed to export:', err)
        this.$message.error('导出失败')
      }
    },

    async exportFolder(folder) {
      try {
        const filePath = await window.go.main.App.ExportFolderMarkdown(folder.id)
        if (filePath) {
          this.$confirm('导出成功！是否打开文件所在目录？', '导出完成', {
            confirmButtonText: '打开目录',
            cancelButtonText: '确定',
            type: 'success'
          }).then(() => {
            window.go.main.App.RevealInFinder(filePath)
          }).catch(() => {})
        }
      } catch (err) {
        if (err && err.includes && err.includes('用户取消')) {
          return
        }
        console.error('Failed to export folder:', err)
        this.$message.error('导出失败')
      }
    },

    async exportAll() {
      try {
        const filePath = await window.go.main.App.ExportAllMarkdown()
        if (filePath) {
          this.$confirm('导出成功！是否打开文件所在目录？', '导出完成', {
            confirmButtonText: '打开目录',
            cancelButtonText: '确定',
            type: 'success'
          }).then(() => {
            window.go.main.App.RevealInFinder(filePath)
          }).catch(() => {})
        }
      } catch (err) {
        if (err && err.includes && err.includes('用户取消')) {
          return
        }
        console.error('Failed to export all:', err)
        this.$message.error('导出失败')
      }
    },

    handleContextMenu(event) {
      const { command, note } = event
      if (command === 'newChild') {
        this.createNoteDirectly(note.id, false)
      } else if (command === 'newFolder') {
        this.createNoteDirectly(note ? note.id : 0, true)
      } else if (command === 'newNote') {
        this.createNoteDirectly(0, false)
      } else if (command === 'delete') {
        this.deleteNote(note)
      } else if (command === 'trash') {
        this.showTrash = true
      } else if (command === 'renamed') {
        this.loadNoteTree()
      } else if (command === 'exportFolder') {
        this.exportFolder(note)
      } else if (command === 'exportAll') {
        this.exportAll()
      }
    },

    // 回收站相关方法
    async openTrash() {
      this.showTrash = !this.showTrash
    },

    async loadTrashNotes() {
      try {
        const notes = await window.go.main.App.GetTrashNotes()
        this.trashNotes = notes || []
      } catch (err) {
        console.error('Failed to load trash notes:', err)
        this.$message.error('加载回收站失败')
      }
    },

    async restoreNote(note) {
      try {
        await this.$confirm(`确定要恢复 "${note.title || '未命名'}" 吗？`, '提示', {
          type: 'info',
          confirmButtonText: '恢复',
          cancelButtonText: '取消'
        })

        await window.go.main.App.RestoreNote(note.id)
        await this.loadTrashNotes()
        await this.loadNoteTree()
        this.$message.success('已恢复')
      } catch (err) {
        if (err !== 'cancel') {
          console.error('Failed to restore note:', err)
          this.$message.error('恢复失败')
        }
      }
    },

    async permanentDeleteNote(note) {
      try {
        await this.$confirm(`确定要永久删除 "${note.title || '未命名'}" 吗？此操作不可恢复！`, '警告', {
          type: 'warning',
          confirmButtonText: '永久删除',
          cancelButtonText: '取消'
        })

        await window.go.main.App.PermanentDeleteNote(note.id)
        await this.loadTrashNotes()
        this.$message.success('已永久删除')
      } catch (err) {
        if (err !== 'cancel') {
          console.error('Failed to permanently delete note:', err)
          this.$message.error('删除失败')
        }
      }
    },

    async emptyTrash() {
      try {
        await this.$confirm('确定要清空回收站吗？所有笔记将被永久删除！', '警告', {
          type: 'warning',
          confirmButtonText: '清空',
          cancelButtonText: '取消'
        })

        await window.go.main.App.EmptyTrash()
        await this.loadTrashNotes()
        this.$message.success('回收站已清空')
      } catch (err) {
        if (err !== 'cancel') {
          console.error('Failed to empty trash:', err)
          this.$message.error('清空失败')
        }
      }
    },

    // 图片压缩为 JPEG 格式，质量 50
    compressImage(file) {
      return new Promise((resolve, reject) => {
        const reader = new FileReader()
        reader.readAsDataURL(file)
        reader.onload = () => {
          const img = new Image()
          img.src = reader.result
          img.onload = () => {
            const canvas = document.createElement('canvas')
            const ctx = canvas.getContext('2d')
            canvas.width = img.width
            canvas.height = img.height
            ctx.drawImage(img, 0, 0)
            // 压缩为 JPEG 格式，质量 50
            const compressedDataUrl = canvas.toDataURL('image/jpeg', 0.5)
            resolve(compressedDataUrl)
          }
          img.onerror = reject
        }
        reader.onerror = reject
      })
    },

    // mavon-editor 图片上传处理
    async handleImgAdd(pos, $file) {
      try {
        // 压缩图片
        const compressedDataUrl = await this.compressImage($file)
        // 保存到后端文件服务器
        const fileUrl = await window.go.main.App.SaveFile(compressedDataUrl, $file.name)
        // 替换编辑器中的图片链接
        this.$refs.md.$img2Url(pos, fileUrl)
        this.$message.success('图片上传成功')
      } catch (err) {
        console.error('Failed to upload image:', err)
        this.$message.error('图片上传失败')
        this.$refs.md.$img2Url(pos, '')
      }
    },

    handleImgDel(pos) {
      // 图片删除时可以做清理工作（可选）
      console.log('Image deleted at position:', pos)
    }
  },
  watch: {
    showTrash(newVal) {
      if (newVal) {
        this.loadTrashNotes()
      }
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}

.app-container {
  height: 100vh;
}

.sidebar {
  background-color: #f5f7fa;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  overflow: hidden !important;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.user-info:hover {
  background-color: #ebeef5;
}

.search-box {
  margin-top: 8px;
}

.sidebar-footer {
  margin-top: auto;
  padding: 12px 16px;
  border-top: 1px solid #e4e7ed;
}

.main-content {
  padding: 0;
  display: flex;
  flex-direction: column;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
}

.empty-state i {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-state p {
  font-size: 16px;
}

/* 标签页样式 */
.tabs-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.tabs-header {
  display: flex;
  background-color: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  overflow-x: auto;
  min-height: 40px;
}

.tab-item {
  display: flex;
  align-items: center;
  padding: 0 16px;
  height: 40px;
  border-right: 1px solid #e4e7ed;
  cursor: pointer;
  background-color: #f5f7fa;
  min-width: 120px;
  max-width: 200px;
  transition: all 0.2s;
}

.tab-item:hover {
  background-color: #ebeef5;
}

.tab-item.is-active {
  background-color: #fff;
  border-bottom: 2px solid #409eff;
}

.tab-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}

.tab-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  margin-left: 8px;
  opacity: 0;
  transition: all 0.2s;
}

.tab-item:hover .tab-close {
  opacity: 1;
}

.tab-close:hover {
  background-color: #dcdfe6;
}

.editor-container {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.editor-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  gap: 12px;
}

.note-title-input {
  flex: 1;
}

.note-title-input .el-input__inner {
  font-size: 20px;
  font-weight: 600;
  border: none;
  padding: 0;
  background: transparent;
}

.note-title-input .el-input__inner:focus {
  outline: none;
}

.editor-actions {
  display: flex;
  gap: 8px;
}

/* 覆盖 mavon-editor 样式 */
.v-note-wrapper {
  border: none !important;
  border-radius: 0 !important;
  z-index: 0 !important;
}

.v-note-op {
  border-bottom: 1px solid #e4e7ed !important;
}

.v-show-content {
  padding: 20px !important;
}
</style>
