<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, toRaw } from 'vue';
import { GetConfig, SaveConfig, StartScheduler, StopScheduler, IsSchedulerRunning, SelectDirectory, MinimizeWindow, OpenRegionSelector, SubmitRegion, SetWindowTitle, HideWindow, QuitApp } from '../bindings/autoscreen/appservice.js';
import { Events } from '@wailsio/runtime';
import { TaskConfig } from '../bindings/autoscreen/pkg/config/models.js';
import CustomSelect from './components/CustomSelect.vue';

const isSelectorMode = ref(window.location.search.includes('mode=selector'));

const lang = ref('zh');
const dict = {
  zh: {
    title: "QuietSnap",
    subtitle: "自动截屏控制面板",
    saveDir: "保存目录",
    saveDirHint: "例如: D:\\Screenshots",
    browse: "浏览...",
    startDate: "开始日期",
    endDate: "结束日期",
    startTime: "开始时间",
    endTime: "结束时间",
    interval: "周期 (分)",
    region: "范围",
    format: "格式",
    fullscreen: "全屏",
    regionSelect: "框选",
    pngDesc: "PNG",
    jpgDesc: "JPG",
    selectedRegion: "已选区域",
    reselect: "框选",
    autoStart: "开机自启",
    showUiAuto: "启动时显示界面",
    saveSettings: "保存设置",
    startAuto: "开始运行",
    stopAuto: "停止运行",
    statusRunning: "运行中",
    statusStopped: "已停止",
    alertSaveDir: "请先为所有任务选择保存目录！",
    alertSaved: "设置已保存！",
    alertSaveFail: "保存失败：",
    ok: "确定",
    langToggle: "EN",
    addTask: "添加任务",
    deleteTask: "删除",
    enableTask: "启用",
    taskName: "任务名称",
    viewModeList: "列表视图",
    viewModeCard: "卡片视图",
    datePlaceholder: "年 / 月 / 日",
    editConfig: "编辑配置",
    hideConfig: "收起",
    closeAppTitle: "退出程序",
    closeAppMsg: "您是要缩小到托盘运行，还是直接退出程序？",
    minimizeToTray: "缩小到托盘",
    quitDirectly: "直接退出",
    cancel: "取消"
  },
  en: {
    title: "QuietSnap",
    subtitle: "Auto Screenshot Panel",
    saveDir: "Save Dir",
    saveDirHint: "e.g. D:\\Screenshots",
    browse: "Browse...",
    startDate: "Start Date",
    endDate: "End Date",
    startTime: "Start Time",
    endTime: "End Time",
    interval: "Interval (m)",
    region: "Area",
    format: "Format",
    fullscreen: "Fullscreen",
    regionSelect: "Region",
    pngDesc: "PNG",
    jpgDesc: "JPG",
    selectedRegion: "Selected",
    reselect: "Select",
    autoStart: "Run on Startup",
    showUiAuto: "Show UI on Startup",
    saveSettings: "Save Settings",
    startAuto: "Start Capture",
    stopAuto: "Stop Capture",
    statusRunning: "Running",
    statusStopped: "Stopped",
    alertSaveDir: "Please select save directories for all tasks!",
    alertSaved: "Settings saved!",
    alertSaveFail: "Failed to save: ",
    ok: "OK",
    langToggle: "中",
    addTask: "Add Task",
    deleteTask: "Delete",
    enableTask: "Enable",
    taskName: "Task Name",
    viewModeList: "List View",
    viewModeCard: "Card View",
    datePlaceholder: "yyyy / mm / dd",
    editConfig: "Edit Config",
    hideConfig: "Hide",
    closeAppTitle: "Exit Program",
    closeAppMsg: "Would you like to minimize to the system tray or quit directly?",
    minimizeToTray: "Minimize to Tray",
    quitDirectly: "Quit Directly",
    cancel: "Cancel"
  }
};
const t = (key: keyof typeof dict.zh) => dict[lang.value as keyof typeof dict][key] || key;

watch(lang, (newLang, oldLang) => {
  document.documentElement.lang = newLang === 'zh' ? 'zh-CN' : 'en';
  SetWindowTitle(newLang === 'zh' ? 'QuietSnap 自动截屏控制面板' : 'QuietSnap Control Panel');

  if (oldLang && tasks.value) {
    const oldPrefix = dict[oldLang as keyof typeof dict].taskName;
    const newPrefix = dict[newLang as keyof typeof dict].taskName;
    tasks.value.forEach(t => {
      if (t._isNew && t.name.startsWith(oldPrefix)) {
        t.name = t.name.replace(oldPrefix, newPrefix);
      }
    });
  }
}, { immediate: true });

// State
const tasks = ref<any[]>([]);
const autoStart = ref(false);
const showUIOnAutoStart = ref(true);
const viewMode = ref<'list' | 'card'>('card');
const editingTaskId = ref<string | null>(null);
const showCloseDialog = ref(false);

const isRunning = ref(false);
let statusInterval: number;
const alertMessage = ref('');

const loadConfig = async () => {
  try {
    const cfg = await GetConfig();
    tasks.value = cfg?.tasks || [];
    if (tasks.value.length === 0) {
      addTask();
    }
    autoStart.value = cfg?.autoStart || false;
    showUIOnAutoStart.value = cfg?.showUIOnAutoStart !== false;
  } catch (err) {
    console.error("Failed to load config:", err);
  }
};

const addTask = () => {
  tasks.value.push({
    id: Date.now().toString(),
    name: `${t('taskName')} ${tasks.value.length + 1}`,
    _isNew: true,
    enabled: true,
    saveDirectory: '',
    startDate: '',
    endDate: '',
    dailyStartTime: '',
    dailyEndTime: '',
    intervalMinutes: 5,
    mode: 'fullscreen',
    format: 'png',
    regionX: 0, regionY: 0, regionW: 800, regionH: 600
  });
};

const deleteTask = (id: string) => {
  tasks.value = tasks.value.filter(t => t.id !== id);
  if (tasks.value.length === 0) addTask();
};

const saveConfig = async (suppressAlert = false) => {
  try {
    tasks.value.forEach(t => delete t._isNew);
    
    await SaveConfig({
      tasks: toRaw(tasks.value),
      autoStart: autoStart.value,
      showUIOnAutoStart: showUIOnAutoStart.value
    });
    if (!suppressAlert) alertMessage.value = t('alertSaved');
  } catch (err) {
    console.error("Failed to save config:", err);
    if (!suppressAlert) alertMessage.value = t('alertSaveFail') + err;
  }
};

const toggleScheduler = async () => {
  if (!isRunning.value) {
    const hasEmptyDir = tasks.value.some(t => t.enabled && !t.saveDirectory);
    if (hasEmptyDir) {
      alertMessage.value = t('alertSaveDir');
      return;
    }
    await saveConfig(true);
  }

  try {
    if (isRunning.value) {
      await StopScheduler();
    } else {
      await StartScheduler();
      await MinimizeWindow();
    }
    checkStatus();
  } catch (err) {
    console.error("Toggle failed:", err);
  }
};

const handleSelectDirectory = async (task: any) => {
  try {
    const path = await SelectDirectory();
    if (path) {
      task.saveDirectory = path;
    }
  } catch (err) {
    console.error("Select directory failed:", err);
  }
};

const checkStatus = async () => {
  if (isSelectorMode.value) return;
  try {
    isRunning.value = await IsSchedulerRunning();
  } catch (err) {
    console.error("Status check failed:", err);
  }
};

const handleOpenRegionSelector = (taskId: string) => {
  editingTaskId.value = taskId;
  OpenRegionSelector();
};

// Region Selection logic
const startX = ref(0);
const startY = ref(0);
const currentX = ref(0);
const currentY = ref(0);
const isDragging = ref(false);

const getBoxStyle = () => {
  const x = Math.min(startX.value, currentX.value);
  const y = Math.min(startY.value, currentY.value);
  const w = Math.abs(currentX.value - startX.value);
  const h = Math.abs(currentY.value - startY.value);
  return { left: x + 'px', top: y + 'px', width: w + 'px', height: h + 'px' };
};

const onMouseDown = (e: MouseEvent) => {
  isDragging.value = true;
  startX.value = e.clientX;
  startY.value = e.clientY;
  currentX.value = e.clientX;
  currentY.value = e.clientY;
};

const onMouseMove = (e: MouseEvent) => {
  if (!isDragging.value) return;
  currentX.value = e.clientX;
  currentY.value = e.clientY;
};

const onMouseUp = () => {
  if (!isDragging.value) return;
  isDragging.value = false;
  const x = Math.min(startX.value, currentX.value);
  const y = Math.min(startY.value, currentY.value);
  const w = Math.abs(currentX.value - startX.value);
  const h = Math.abs(currentY.value - startY.value);
  
  if (w > 10 && h > 10) {
    SubmitRegion({ x, y, w, h });
  } else {
    SubmitRegion({ x: 0, y: 0, w: 0, h: 0 });
  }
};

onMounted(() => {
  if (isSelectorMode.value) {
    document.body.style.background = 'transparent';
    return;
  }
  
  loadConfig();
  checkStatus();
  statusInterval = window.setInterval(checkStatus, 2000);
  
  Events.On("region-selected", (e: any) => {
    let data = e.data;
    if (Array.isArray(e.data)) {
      data = e.data[0];
    }
    if (data && data.w > 0 && data.h > 0) {
      const task = tasks.value.find(t => t.id === editingTaskId.value);
      if (task) {
        task.regionX = data.x;
        task.regionY = data.y;
        task.regionW = data.w;
        task.regionH = data.h;
      }
    }
    editingTaskId.value = null;
  });

  Events.On("show-close-dialog", () => {
    showCloseDialog.value = true;
  });
});

onUnmounted(() => {
  if (statusInterval) clearInterval(statusInterval);
});
</script>

<template>
  <div v-if="!isSelectorMode" class="app-container">
    <header class="app-header">
      <div class="header-titles">
        <h1>{{ t('title') }}</h1>
        <span class="subtitle">{{ t('subtitle') }}</span>
      </div>
      <div class="header-actions">
        <button class="btn-icon mode-toggle" @click="viewMode = viewMode === 'card' ? 'list' : 'card'">
          {{ viewMode === 'card' ? t('viewModeList') : t('viewModeCard') }}
        </button>
        <button class="btn-icon lang-toggle" @click="lang = lang === 'zh' ? 'en' : 'zh'">{{ t('langToggle') }}</button>
      </div>
    </header>

    <div class="main-content">
      <div class="tasks-container" :class="viewMode">
        <div v-for="task in tasks" :key="task.id" class="task-item" :class="{ 'task-disabled': !task.enabled }">
          
          <div class="task-header">
            <input type="text" v-model="task.name" class="task-name-input" />
            <div class="task-controls">
              <label class="toggle-switch">
                <input type="checkbox" v-model="task.enabled">
                <span class="slider"></span>
              </label>
              <button class="btn-danger btn-sm" @click="deleteTask(task.id)">×</button>
            </div>
          </div>

          <div class="task-body" v-show="viewMode === 'card' || editingTaskId === task.id">
            
            <div class="form-group row-group">
              <label>{{ t('saveDir') }}</label>
              <div class="input-with-button">
                <input type="text" v-model="task.saveDirectory" :placeholder="t('saveDirHint')" readonly />
                <button class="btn-secondary" @click="handleSelectDirectory(task)">{{ t('browse') }}</button>
              </div>
            </div>

            <div class="form-row grid-2">
              <div class="form-group">
                <label>{{ t('startDate') }}</label>
                <div class="date-input-wrapper">
                  <input type="date" v-model="task.startDate" :class="{ 'empty-date': !task.startDate }" />
                  <span class="date-placeholder" v-if="!task.startDate">{{ t('datePlaceholder') }}</span>
                </div>
              </div>
              <div class="form-group">
                <label>{{ t('endDate') }}</label>
                <div class="date-input-wrapper">
                  <input type="date" v-model="task.endDate" :class="{ 'empty-date': !task.endDate }" />
                  <span class="date-placeholder" v-if="!task.endDate">{{ t('datePlaceholder') }}</span>
                </div>
              </div>
            </div>

            <div class="form-row grid-3">
              <div class="form-group">
                <label>{{ t('startTime') }}</label>
                <input type="time" v-model="task.dailyStartTime" />
              </div>
              <div class="form-group">
                <label>{{ t('endTime') }}</label>
                <input type="time" v-model="task.dailyEndTime" />
              </div>
              <div class="form-group">
                <label>{{ t('interval') }}</label>
                <input type="number" v-model="task.intervalMinutes" min="1" />
              </div>
            </div>

            <div class="form-row grid-2">
              <div class="form-group">
                <label>{{ t('region') }}</label>
                <CustomSelect 
                  v-model="task.mode" 
                  @change="(val) => { if(val === 'region') handleOpenRegionSelector(task.id) }"
                  :options="[
                    { value: 'fullscreen', label: t('fullscreen') },
                    { value: 'region', label: t('regionSelect') }
                  ]"
                />
              </div>
              <div class="form-group">
                <label>{{ t('format') }}</label>
                <CustomSelect 
                  v-model="task.format"
                  :options="[
                    { value: 'png', label: t('pngDesc') },
                    { value: 'jpg', label: t('jpgDesc') }
                  ]"
                />
              </div>
            </div>

            <div v-if="task.mode === 'region'" class="form-group region-info row-group">
              <label>{{ t('selectedRegion') }}</label>
              <div class="input-with-button">
                <input type="text" :value="`X:${task.regionX} Y:${task.regionY} W:${task.regionW} H:${task.regionH}`" readonly />
                <button class="btn-secondary" @click="handleOpenRegionSelector(task.id)">{{ t('reselect') }}</button>
              </div>
            </div>

          </div>

          <!-- Edit button for list mode -->
          <div class="task-footer" v-if="viewMode === 'list'">
             <button class="btn-text" @click="editingTaskId = editingTaskId === task.id ? null : task.id">
               {{ editingTaskId === task.id ? t('hideConfig') : t('editConfig') }}
             </button>
          </div>

        </div>
        
        <button class="btn-add-task" @click="addTask">+ {{ t('addTask') }}</button>
      </div>
    </div>

    <footer class="app-footer">
      <div class="global-settings">
        <label class="checkbox-group">
          <input type="checkbox" v-model="autoStart" />
          <span>{{ t('autoStart') }}</span>
        </label>
        <label class="checkbox-group" :class="{'disabled': !autoStart}">
          <input type="checkbox" v-model="showUIOnAutoStart" :disabled="!autoStart" />
          <span>{{ t('showUiAuto') }}</span>
        </label>
      </div>

      <div class="primary-actions">
        <div class="status-badge" :class="{ running: isRunning }">
          <span class="pulse"></span>
          {{ isRunning ? t('statusRunning') : t('statusStopped') }}
        </div>
        <button class="btn-primary outline" @click="saveConfig(false)">{{ t('saveSettings') }}</button>
        <button class="btn-primary" :class="{ 'btn-danger': isRunning }" @click="toggleScheduler">
          {{ isRunning ? t('stopAuto') : t('startAuto') }}
        </button>
      </div>
    </footer>

    <!-- Custom Close Dialog Modal -->
    <div v-if="showCloseDialog" class="modal-overlay">
      <div class="glass-modal">
        <h3>{{ t('closeAppTitle') }}</h3>
        <p>{{ t('closeAppMsg') }}</p>
        <div class="modal-actions">
          <button class="btn-primary" @click="showCloseDialog = false; HideWindow()">{{ t('minimizeToTray') }}</button>
          <button class="btn-danger" @click="QuitApp">{{ t('quitDirectly') }}</button>
          <button class="btn-secondary" @click="showCloseDialog = false">{{ t('cancel') }}</button>
        </div>
      </div>
    </div>

    <!-- Alert Modal -->
    <div v-if="alertMessage" class="modal-overlay">
      <div class="glass-modal">
        <p>{{ alertMessage }}</p>
        <button class="btn-primary" @click="alertMessage = ''">{{ t('ok') }}</button>
      </div>
    </div>
  </div>

  <div v-if="isSelectorMode" class="selector-overlay" @mousedown="onMouseDown" @mousemove="onMouseMove" @mouseup="onMouseUp">
    <div class="selector-instructions">框选截图区域</div>
    <div v-if="isDragging" class="selector-box" :style="getBoxStyle()"></div>
  </div>
</template>
