<template>
  <div class="note-tree" @contextmenu.prevent="showContextMenu($event, null)">
    <div v-if="!notes || notes.length === 0" class="empty-tree">
      <p>暂无笔记</p>
    </div>
    
    <div v-else class="tree-content">
      <div
        v-for="note in notes"
        :key="note.id"
        class="tree-item"
      >
        <div
          class="tree-node"
          :class="{
            'is-selected': selectedId === note.id,
            'is-folder': note.isFolder
          }"
          @click="$emit('select', note)"
          @contextmenu.prevent="showContextMenu($event, note)"
        >
          <span
            v-if="note.isFolder"
            class="expand-icon"
            :class="{ 'is-expanded': expandedIds.includes(note.id) }"
            @click.stop="toggleExpand(note.id)"
          >
            <i class="el-icon-arrow-right"></i>
          </span>
          <span v-else class="expand-icon-placeholder"></span>
          
          <span class="node-icon">
            <i :class="getNodeIcon(note)"></i>
          </span>

          <input
            v-if="renamingId === note.id"
            ref="renameInput"
            v-model="renameTitle"
            class="rename-input"
            @blur="confirmRename(note)"
            @keyup.enter="confirmRename(note)"
            @keyup.escape="cancelRename"
            @click.stop
          />
          <span v-else class="node-label-wrapper">
            <span class="node-label" :title="note.title">
              {{ note.title || '未命名' }}
            </span>
            <span v-if="note.createdAt" class="node-date">
              {{ formatDate(note.createdAt) }}
            </span>
          </span>
        </div>
        
        <div
          v-if="note.isFolder && note.children && expandedIds.includes(note.id)"
          class="tree-children"
        >
          <note-tree
            :notes="note.children"
            :selected-id="selectedId"
            @select="$emit('select', $event)"
            @delete="$emit('delete', $event)"
            @context-menu="$emit('context-menu', $event)"
          />
        </div>
      </div>
    </div>
    
    <!-- 右键菜单 -->
    <div
      v-if="contextMenu.visible"
      class="context-menu"
      :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }"
      @click.stop
    >
      <template v-if="contextMenu.note && contextMenu.note.isFolder">
        <div class="context-menu-item" @click.stop="handleContextAction('newChild')">
          <i class="el-icon-document-add"></i> 新建笔记
        </div>
        <div class="context-menu-item" @click.stop="handleContextAction('newFolder')">
          <i class="el-icon-folder-add"></i> 新建子文件夹
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click.stop="handleContextAction('rename')">
          <i class="el-icon-edit"></i> 重命名
        </div>
        <div class="context-menu-item" @click.stop="handleContextAction('exportFolder')">
          <i class="el-icon-download"></i> 导出为 Markdown
        </div>
        <div class="context-menu-item delete" @click.stop="handleContextAction('delete')">
          <i class="el-icon-delete"></i> 删除
        </div>
      </template>
      <template v-else-if="contextMenu.note">
        <div class="context-menu-item delete" @click.stop="handleContextAction('delete')">
          <i class="el-icon-delete"></i> 删除
        </div>
      </template>
      <template v-else>
        <div class="context-menu-item" @click.stop="handleContextAction('newNote')">
          <i class="el-icon-document-add"></i> 新建笔记
        </div>
        <div class="context-menu-item" @click.stop="handleContextAction('newFolder')">
          <i class="el-icon-folder-add"></i> 新建文件夹
        </div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item" @click.stop="handleContextAction('exportAll')">
          <i class="el-icon-download"></i> 导出所有笔记
        </div>
      </template>
    </div>
  </div>
</template>

<script>
export default {
  name: 'NoteTree',
  props: {
    notes: {
      type: Array,
      default: () => []
    },
    selectedId: {
      type: Number,
      default: null
    }
  },
  data() {
    return {
      expandedIds: [],
      renamingId: null,
      renameTitle: '',
      contextMenu: {
        visible: false,
        x: 0,
        y: 0,
        note: null
      }
    }
  },
  mounted() {
    document.addEventListener('click', this.hideContextMenu)
    document.addEventListener('contextmenu', this.hideContextMenu)
  },
  beforeDestroy() {
    document.removeEventListener('click', this.hideContextMenu)
    document.removeEventListener('contextmenu', this.hideContextMenu)
  },
  methods: {
    toggleExpand(id) {
      const index = this.expandedIds.indexOf(id)
      if (index > -1) {
        this.expandedIds.splice(index, 1)
      } else {
        this.expandedIds.push(id)
      }
    },
    
    getNodeIcon(note) {
      if (note.isFolder) {
        return this.expandedIds.includes(note.id) 
          ? 'el-icon-folder-opened' 
          : 'el-icon-folder'
      }
      return 'el-icon-document'
    },
    
    showContextMenu(event, note) {
      event.stopPropagation()
      this.contextMenu = {
        visible: true,
        x: event.clientX,
        y: event.clientY,
        note: note
      }
    },
    
    hideContextMenu() {
      this.contextMenu.visible = false
    },
    
    handleContextAction(command) {
      if (command === 'rename') {
        this.startRename(this.contextMenu.note)
      } else {
        this.$emit('context-menu', {
          command: command,
          note: this.contextMenu.note
        })
      }
      this.hideContextMenu()
    },

    startRename(note) {
      this.renamingId = note.id
      this.renameTitle = note.title
      this.$nextTick(() => {
        if (this.$refs.renameInput) {
          this.$refs.renameInput.focus()
          this.$refs.renameInput.select()
        }
      })
    },

    async confirmRename(note) {
      if (!this.renamingId) return
      const newTitle = this.renameTitle.trim()
      if (newTitle && newTitle !== note.title) {
        await window.go.main.App.UpdateNote(note.id, newTitle, note.content, note.tag)
        this.$emit('context-menu', { command: 'renamed', note })
      }
      this.renamingId = null
      this.renameTitle = ''
    },

    cancelRename() {
      this.renamingId = null
      this.renameTitle = ''
    },

    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      return `${year}-${month}-${day}`
    }
  },
  watch: {
    notes: {
      handler(newNotes) {
        // 自动展开第一级文件夹
        if (newNotes) {
          newNotes.forEach(note => {
            if (note.isFolder && note.children && note.children.length > 0) {
              if (!this.expandedIds.includes(note.id)) {
                this.expandedIds.push(note.id)
              }
            }
          })
        }
      },
      immediate: true
    }
  }
}
</script>

<style scoped>
.note-tree {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.empty-tree {
  padding: 20px;
  text-align: center;
  color: #909399;
}

.tree-content {
  padding: 0 8px;
}

.tree-item {
  margin: 2px 0;
}

.tree-node {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
  user-select: none;
}

.tree-node:hover {
  background-color: #ebeef5;
}

.tree-node.is-selected {
  background-color: #ecf5ff;
  color: #409eff;
}

.tree-node.is-folder {
  font-weight: 500;
}

.expand-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  transition: transform 0.2s;
}

.expand-icon.is-expanded {
  transform: rotate(90deg);
}

.expand-icon-placeholder {
  width: 20px;
}

.node-icon {
  margin-right: 8px;
  font-size: 16px;
}

.node-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-label-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.node-date {
  font-size: 11px;
  color: #909399;
  margin-top: 2px;
}

.rename-input {
  flex: 1;
  border: 1px solid #409eff;
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 14px;
  outline: none;
  background: white;
}

.tree-children {
  padding-left: 20px;
}

/* 右键菜单 */
.context-menu {
  position: fixed;
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  z-index: 1000;
  min-width: 160px;
  padding: 4px 0;
}

.context-menu-item {
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}

.context-menu-item:hover {
  background-color: #f5f7fa;
}

.context-menu-item.delete {
  color: #f56c6c;
}

.context-menu-divider {
  height: 1px;
  background-color: #e4e7ed;
  margin: 4px 0;
}
</style>
