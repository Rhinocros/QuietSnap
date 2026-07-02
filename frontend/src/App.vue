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
    cancel: "取消",
    autoDeleteSettings: "自动清理设置",
    policyNever: "从不删除",
    policyNeverDesc: "保留所有生成的截图文件。",
    policyYesterday: "删除昨天",
    policyYesterdayDesc: "自动清理昨天生成的截图文件。",
    policyPrevRun: "删除上一次运行",
    policyPrevRunDesc: "在每次开始截图任务时，自动清理上一次运行产生的所有图片。",
    policyKeepN: "保存最近 N 天",
    policyKeepNDesc: "仅保留最近几天生成的截图，更早的自动清理。",
    policyCommonNotice: "提示：系统会在每次写入图片前自动检查并执行清理。所有被清理的文件都将放入【回收站】，不会直接彻底删除，以防误删。",
    daysLabel: "天",
    confirmPolicyChange: "确定要将清理策略设置为",
    confirm: "确定",
    currPolicy: "当前策略:"
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
    cancel: "Cancel",
    autoDeleteSettings: "Auto-Delete Settings",
    policyNever: "Never Delete",
    policyNeverDesc: "Keep all generated screenshot files.",
    policyYesterday: "Delete Yesterday",
    policyYesterdayDesc: "Automatically clean up screenshots generated yesterday.",
    policyPrevRun: "Delete Previous Run",
    policyPrevRunDesc: "Automatically clean up all screenshots generated in the previous run when starting a task.",
    policyKeepN: "Keep for N Days",
    policyKeepNDesc: "Only keep screenshots generated in recent days, automatically clean older ones.",
    policyCommonNotice: "Note: Cleanup runs automatically before writing each image. Cleaned files are moved to the Recycle Bin, not permanently deleted.",
    daysLabel: "Days",
    confirmPolicyChange: "Are you sure you want to change the deletion policy to",
    confirm: "Confirm",
    currPolicy: "Current Policy:"
  }
};

watch(lang, (newLang, oldLang) => {
  document.documentElement.lang = newLang === 'zh' ? 'zh-CN' : 'en-US';
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

const t = (key: keyof typeof dict.zh) => dict[lang.value as keyof typeof dict][key] || key;

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

const showDeleteDialog = ref<string | null>(null);
const showDeleteConfirm = ref<string | null>(null);
const tempDeletePolicy = ref<string>('never');
const pendingDeletePolicy = ref<string>('never');
const pendingDeleteDays = ref<number>(7);
const tempDays = ref<number>(7);

const getPolicyName = (p: string) => {
  switch (p) {
    case 'never': return t('policyNever');
    case 'yesterday': return t('policyYesterday');
    case 'previous_run': return t('policyPrevRun');
    case 'keep_n_days': return t('policyKeepN');
    default: return p;
  }
};

const openDeleteDialog = (taskId: string) => {
  const task = tasks.value.find(t => t.id === taskId);
  if (task) {
    tempDays.value = task.deleteDays || 7;
    tempDeletePolicy.value = task.deletePolicy || 'never';
  }
  showDeleteDialog.value = taskId;
};

const preConfirmPolicy = () => {
  confirmPolicy(tempDeletePolicy.value, tempDays.value);
};

const handleDateInput = (task: any, field: 'startDate' | 'endDate', e: Event) => {
  const input = e.target as HTMLInputElement;
  const isDeleting = (e as InputEvent).inputType?.startsWith('delete');
  let val = input.value.replace(/[^\d-]/g, ''); 
  
  if (!isDeleting) {
    let parts = val.split('-');
    
    // Auto-pad month if first digit is 2-9
    if (parts.length === 2 && parts[1].length === 1) {
      if (parseInt(parts[1]) >= 2) {
        parts[1] = '0' + parts[1];
        val = parts.join('-');
      }
    }
    
    // Auto-pad day if first digit is 4-9
    if (parts.length === 3 && parts[2].length === 1) {
      if (parseInt(parts[2]) >= 4) {
        parts[2] = '0' + parts[2];
        val = parts.join('-');
      }
    }

    // Restrict month to <= 12
    if (parts.length >= 2 && parts[1].length === 2) {
      let m = parseInt(parts[1]);
      if (m > 12) {
        parts[1] = '12';
        val = parts.join('-');
      } else if (m === 0) {
        parts[1] = '01';
        val = parts.join('-');
      }
    }

    // Restrict day to <= 31
    if (parts.length === 3 && parts[2].length === 2) {
      let d = parseInt(parts[2]);
      if (d > 31) {
        parts[2] = '31';
        val = parts.join('-');
      } else if (d === 0) {
        parts[2] = '01';
        val = parts.join('-');
      }
    }

    if (val.length === 4 && !val.includes('-')) val += '-';
    else if (val.length === 7 && val.split('-').length === 2) val += '-';
  }
  
  task[field] = val;
  input.value = val;
};

const handleDateBlur = (task: any, field: 'startDate' | 'endDate', e: Event) => {
  const input = e.target as HTMLInputElement;
  let val = input.value;
  
  if (/^\d{8}$/.test(val)) {
    val = `${val.substring(0,4)}-${val.substring(4,6)}-${val.substring(6,8)}`;
  } else {
    let parts = val.split('-');
    if (parts.length === 3) {
      let y = parts[0];
      let m = parts[1].padStart(2, '0');
      let d = parts[2].padStart(2, '0');
      if (y.length === 4) {
        val = `${y}-${m}-${d}`;
      }
    }
  }
  
  task[field] = val;
  input.value = val;
};

const confirmPolicy = (policy: string, days?: number) => {
  pendingDeletePolicy.value = policy;
  if (days) pendingDeleteDays.value = days;
  showDeleteConfirm.value = policy;
};

const applyPolicy = () => {
  if (showDeleteDialog.value) {
    const task = tasks.value.find(t => t.id === showDeleteDialog.value);
    if (task) {
      task.deletePolicy = pendingDeletePolicy.value;
      if (pendingDeletePolicy.value === 'keep_n_days') {
        task.deleteDays = pendingDeleteDays.value;
      }
    }
  }
  showDeleteConfirm.value = null;
  showDeleteDialog.value = null;
};

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
    regionX: 0, regionY: 0, regionW: 800, regionH: 600,
    deletePolicy: 'never',
    deleteDays: 7
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
                  <input type="text" :value="task.startDate" @input="handleDateInput(task, 'startDate', $event)" @blur="handleDateBlur(task, 'startDate', $event)" :placeholder="t('datePlaceholder')" maxlength="10" />
                  <div class="calendar-icon-wrapper">
                    <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                    <input type="date" :value="task.startDate" @input="task.startDate = ($event.target as HTMLInputElement).value" onclick="try{this.showPicker()}catch(e){}" class="hidden-date-picker" />
                  </div>
                </div>
              </div>
              <div class="form-group">
                <label>{{ t('endDate') }}</label>
                <div class="date-input-wrapper">
                  <input type="text" :value="task.endDate" @input="handleDateInput(task, 'endDate', $event)" @blur="handleDateBlur(task, 'endDate', $event)" :placeholder="t('datePlaceholder')" maxlength="10" />
                  <div class="calendar-icon-wrapper">
                    <svg viewBox="0 0 24 24" width="16" height="16" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                    <input type="date" :value="task.endDate" @input="task.endDate = ($event.target as HTMLInputElement).value" onclick="try{this.showPicker()}catch(e){}" class="hidden-date-picker" />
                  </div>
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

            <div class="form-group row-group" style="margin-top: 15px;">
              <label>{{ t('autoDeleteSettings') }}</label>
              <div class="input-with-button">
                <input type="text" :value="getPolicyName(task.deletePolicy || 'never') + (task.deletePolicy === 'keep_n_days' ? ' (' + (task.deleteDays || 7) + ' ' + t('daysLabel') + ')' : '')" readonly />
                <button class="btn-secondary" @click="openDeleteDialog(task.id)">{{ t('autoDeleteSettings') }}</button>
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

    <!-- Auto Delete Settings Modal -->
    <div v-if="showDeleteDialog" class="modal-overlay" style="z-index: 1000">
      <div class="glass-modal" style="width: 600px; max-width: 90vw; text-align: left;">
        <h3 style="margin-bottom: 20px; text-align: center;">{{ t('autoDeleteSettings') }}</h3>
        
        <div class="policy-list" style="display: flex; flex-direction: column; gap: 15px; margin-bottom: 20px;">
          <div class="policy-item" :style="{ borderColor: tempDeletePolicy === 'never' ? '#4a90e2' : 'rgba(255,255,255,0.1)' }" style="padding: 10px; border-width: 2px; border-style: solid; border-radius: 8px; cursor: pointer; transition: all 0.2s ease;" @click="tempDeletePolicy = 'never'">
            <h4 style="margin: 0 0 5px 0; color: #fff;">{{ t('policyNever') }}</h4>
            <p style="margin: 0; font-size: 0.9em; opacity: 0.8;">{{ t('policyNeverDesc') }}</p>
          </div>
          
          <div class="policy-item" :style="{ borderColor: tempDeletePolicy === 'yesterday' ? '#4a90e2' : 'rgba(255,255,255,0.1)' }" style="padding: 10px; border-width: 2px; border-style: solid; border-radius: 8px; cursor: pointer; transition: all 0.2s ease;" @click="tempDeletePolicy = 'yesterday'">
            <h4 style="margin: 0 0 5px 0; color: #fff;">{{ t('policyYesterday') }}</h4>
            <p style="margin: 0; font-size: 0.9em; opacity: 0.8;">{{ t('policyYesterdayDesc') }}</p>
          </div>
          
          <div class="policy-item" :style="{ borderColor: tempDeletePolicy === 'previous_run' ? '#4a90e2' : 'rgba(255,255,255,0.1)' }" style="padding: 10px; border-width: 2px; border-style: solid; border-radius: 8px; cursor: pointer; transition: all 0.2s ease;" @click="tempDeletePolicy = 'previous_run'">
            <h4 style="margin: 0 0 5px 0; color: #fff;">{{ t('policyPrevRun') }}</h4>
            <p style="margin: 0; font-size: 0.9em; opacity: 0.8;">{{ t('policyPrevRunDesc') }}</p>
          </div>
          
          <div class="policy-item" :style="{ borderColor: tempDeletePolicy === 'keep_n_days' ? '#4a90e2' : 'rgba(255,255,255,0.1)' }" style="padding: 10px; border-width: 2px; border-style: solid; border-radius: 8px; cursor: pointer; transition: all 0.2s ease;" @click="tempDeletePolicy = 'keep_n_days'">
            <h4 style="margin: 0 0 5px 0; color: #fff;">{{ t('policyKeepN') }}</h4>
            <p style="margin: 0 0 10px 0; font-size: 0.9em; opacity: 0.8;">{{ t('policyKeepNDesc') }}</p>
            <div style="display: flex; gap: 10px; align-items: center;">
              <input type="number" v-model="tempDays" min="1" style="width: 80px; padding: 5px; border-radius: 4px; border: 1px solid rgba(255,255,255,0.2); background: rgba(0,0,0,0.2); color: white;" @click.stop /> 
              <span>{{ t('daysLabel') }}</span>
            </div>
          </div>
        </div>
        
        <div style="font-size: 0.85em; opacity: 0.7; margin-bottom: 20px; line-height: 1.4; color: #ccc;">
          {{ t('policyCommonNotice') }}
        </div>
        
        <div style="display: flex; justify-content: center; gap: 15px; margin-top: 20px;">
          <button class="btn-primary" @click="preConfirmPolicy">{{ t('confirm') }}</button>
          <button class="btn-secondary" @click="showDeleteDialog = null">{{ t('cancel') }}</button>
        </div>
      </div>
    </div>

    <!-- Confirm Policy Modal -->
    <div v-if="showDeleteConfirm" class="modal-overlay" style="z-index: 1001">
      <div class="glass-modal">
        <p style="font-size: 1.1em; margin-bottom: 20px;">{{ t('confirmPolicyChange') }} <strong>[{{ getPolicyName(pendingDeletePolicy) }}]</strong> ?</p>
        <div class="modal-actions">
          <button class="btn-primary" @click="applyPolicy()">{{ t('confirm') }}</button>
          <button class="btn-secondary" @click="showDeleteConfirm = null">{{ t('cancel') }}</button>
        </div>
      </div>
    </div>

    <!-- Alert Modal -->
    <div v-if="alertMessage" class="modal-overlay" style="z-index: 1002">
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
