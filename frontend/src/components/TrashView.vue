<template>
  <div class="trash-view">
    <div class="trash-header">
      <h3>
        <i class="el-icon-delete"></i> 回收站
        <span v-if="trashNotes.length > 0" class="trash-count">({{ trashNotes.length }})</span>
      </h3>
      <div class="trash-header-actions">
        <el-button
          v-if="trashNotes.length > 0"
          type="danger"
          size="small"
          icon="el-icon-delete-solid"
          @click="emptyTrash"
        >
          清空回收站
        </el-button>
      </div>
    </div>

    <div v-if="trashNotes.length === 0" class="empty-trash">
      <i class="el-icon-folder-delete"></i>
      <p>回收站为空</p>
    </div>

    <div v-else class="trash-list">
      <div
        v-for="note in trashNotes"
        :key="note.id"
        class="trash-item"
      >
        <div class="trash-item-icon">
          <i :class="note.isFolder ? 'el-icon-folder' : 'el-icon-document'"></i>
        </div>
        <div class="trash-item-info">
          <div class="trash-item-title">{{ note.title || '未命名' }}</div>
          <div class="trash-item-meta">
            <span v-if="note.deletedAt" class="delete-time">
              删除于: {{ formatDate(note.deletedAt) }}
            </span>
          </div>
        </div>
        <div class="trash-item-actions">
          <el-button
            type="primary"
            size="mini"
            icon="el-icon-refresh"
            @click="restoreNote(note)"
          >
            恢复
          </el-button>
          <el-button
            type="danger"
            size="mini"
            icon="el-icon-delete"
            @click="permanentDelete(note)"
          >
            永久删除
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'TrashView',
  props: {
    trashNotes: {
      type: Array,
      default: () => []
    }
  },
  methods: {
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      return `${year}-${month}-${day} ${hours}:${minutes}`
    },

    restoreNote(note) {
      this.$emit('restore', note)
    },

    permanentDelete(note) {
      this.$emit('permanent-delete', note)
    },

    emptyTrash() {
      this.$emit('empty-trash')
    }
  }
}
</script>

<style scoped>
.trash-view {
  padding: 16px;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.trash-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.trash-header-actions {
  display: flex;
  gap: 8px;
}

.trash-header h3 {
  margin: 0;
  font-size: 18px;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 8px;
}

.trash-count {
  color: #909399;
  font-weight: normal;
}

.empty-trash {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.empty-trash i {
  font-size: 64px;
  margin-bottom: 16px;
}

.trash-list {
  flex: 1;
  overflow-y: auto;
}

.trash-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  margin-bottom: 8px;
  transition: all 0.2s;
}

.trash-item:hover {
  border-color: #dcdfe6;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.trash-item-icon {
  font-size: 24px;
  color: #909399;
  margin-right: 12px;
}

.trash-item-info {
  flex: 1;
  min-width: 0;
}

.trash-item-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trash-item-meta {
  font-size: 12px;
  color: #909399;
}

.trash-item-actions {
  display: flex;
  gap: 8px;
  margin-left: 12px;
}
</style>
