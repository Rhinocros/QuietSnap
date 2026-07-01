<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { GetConfig, SaveConfig, StartScheduler, StopScheduler, IsSchedulerRunning, SelectDirectory, MinimizeWindow, OpenRegionSelector, SubmitRegion, SetWindowTitle } from '../bindings/autoscreen/appservice.js';
import { Events } from '@wailsio/runtime';

const isSelectorMode = ref(window.location.search.includes('mode=selector'));

const lang = ref('zh');
const dict = {
  zh: {
    title: "QuietSnap 自动截屏",
    saveDir: "保存目录",
    saveDirHint: "例如: D:\\Screenshots",
    browse: "浏览...",
    startDate: "开始日期",
    endDate: "结束日期 (可选)",
    startTime: "每日开始时间",
    endTime: "每日结束时间 (可选)",
    interval: "执行周期 (分钟)",
    region: "截图范围",
    format: "图片格式",
    fullscreen: "全屏截图",
    regionSelect: "框选区域",
    pngDesc: "PNG (清晰, 体积大)",
    jpgDesc: "JPG (推荐, 体积小)",
    selectedRegion: "已选择区域",
    reselect: "重新框选",
    autoStart: "开机自动启动",
    showUiAuto: "开机启动时显示界面",
    saveSettings: "保存设置",
    startAuto: "开始自动截图",
    stopAuto: "停止自动截图",
    statusRunning: "截图中...",
    statusStopped: "已停止",
    alertSaveDir: "请先选择保存目录，确认后再开始截图！",
    alertStartDate: "请先选择开始日期，确认后再开始截图！",
    alertStartTime: "请先选择每日开始时间，确认后再开始截图！",
    alertSaved: "设置已保存！",
    alertSaveFail: "保存失败：",
    ok: "确定",
    langToggle: "English",
    datePlaceholder: "年 / 月 / 日"
  },
  en: {
    title: "QuietSnap",
    saveDir: "Save Directory",
    saveDirHint: "e.g. D:\\Screenshots",
    browse: "Browse...",
    startDate: "Start Date",
    endDate: "End Date (Opt)",
    startTime: "Daily Start",
    endTime: "Daily End (Opt)",
    interval: "Interval (mins)",
    region: "Capture Area",
    format: "Image Format",
    fullscreen: "Fullscreen",
    regionSelect: "Select Region",
    pngDesc: "PNG (High Quality)",
    jpgDesc: "JPG (Recommended)",
    selectedRegion: "Selected Region",
    reselect: "Reselect",
    autoStart: "Run on Startup",
    showUiAuto: "Show UI on Startup",
    saveSettings: "Save Settings",
    startAuto: "Start Capture",
    stopAuto: "Stop Capture",
    statusRunning: "Capturing...",
    statusStopped: "Stopped",
    alertSaveDir: "Please select a save directory first!",
    alertStartDate: "Please select a start date first!",
    alertStartTime: "Please select a daily start time first!",
    alertSaved: "Settings saved!",
    alertSaveFail: "Failed to save: ",
    ok: "OK",
    langToggle: "中文",
    datePlaceholder: "yyyy / mm / dd"
  }
};
const t = (key: keyof typeof dict.zh) => dict[lang.value as keyof typeof dict][key] || key;

watch(lang, (newLang) => {
  document.documentElement.lang = newLang === 'zh' ? 'zh-CN' : 'en';
  SetWindowTitle(newLang === 'zh' ? 'QuietSnap 自动截屏控制面板' : 'QuietSnap Control Panel');
}, { immediate: true });

const saveDirectory = ref('');
const startDate = ref('');
const endDate = ref('');
const dailyStartTime = ref('');
const dailyEndTime = ref('');
const intervalMinutes = ref(5);
const mode = ref('fullscreen');
const format = ref('png');
const regionX = ref(0);
const regionY = ref(0);
const regionW = ref(800);
const regionH = ref(600);
const autoStart = ref(false);
const showUIOnAutoStart = ref(true);

const isRunning = ref(false);
let statusInterval: number;
const alertMessage = ref('');

const loadConfig = async () => {
  try {
    const cfg = await GetConfig();
    saveDirectory.value = cfg.saveDirectory || '';
    startDate.value = cfg.startDate || '';
    endDate.value = cfg.endDate || '';
    dailyStartTime.value = cfg.dailyStartTime || '';
    dailyEndTime.value = cfg.dailyEndTime || '';
    intervalMinutes.value = cfg.intervalMinutes || 5;
    mode.value = cfg.mode || 'fullscreen';
    format.value = cfg.format || 'png';
    regionX.value = cfg.regionX || 0;
    regionY.value = cfg.regionY || 0;
    regionW.value = cfg.regionW || 800;
    regionH.value = cfg.regionH || 600;
    autoStart.value = cfg.autoStart || false;
    showUIOnAutoStart.value = cfg.showUIOnAutoStart !== false; // default true
  } catch (err) {
    console.error("Failed to load config:", err);
  }
};

const saveConfig = async (suppressAlert = false) => {
  try {
    await SaveConfig({
      saveDirectory: saveDirectory.value,
      startDate: startDate.value,
      endDate: endDate.value,
      dailyStartTime: dailyStartTime.value,
      dailyEndTime: dailyEndTime.value,
      intervalMinutes: intervalMinutes.value,
      mode: mode.value,
      format: format.value,
      regionX: regionX.value,
      regionY: regionY.value,
      regionW: regionW.value,
      regionH: regionH.value,
      autoStart: autoStart.value,
      showUIOnAutoStart: showUIOnAutoStart.value
    });
    if (!suppressAlert) {
      alertMessage.value = t('alertSaved');
    }
  } catch (err) {
    console.error("Failed to save config:", err);
    if (!suppressAlert) {
      alertMessage.value = t('alertSaveFail') + err;
    }
  }
};

const toggleScheduler = async () => {
  if (!isRunning.value) {
    if (!saveDirectory.value) {
      alertMessage.value = t('alertSaveDir');
      return;
    }
    if (!startDate.value) {
      alertMessage.value = t('alertStartDate');
      return;
    }
    if (!dailyStartTime.value) {
      alertMessage.value = t('alertStartTime');
      return;
    }
    // Auto-save the config before starting
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

const handleSelectDirectory = async () => {
  try {
    const path = await SelectDirectory();
    if (path) {
      saveDirectory.value = path;
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

const handleModeChange = () => {
  if (mode.value === 'region') {
    handleOpenRegionSelector();
  }
};

const handleOpenRegionSelector = async () => {
  OpenRegionSelector();
};

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
  return {
    left: x + 'px',
    top: y + 'px',
    width: w + 'px',
    height: h + 'px'
  };
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
    // Too small, ignore and close
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
    const data = e.data[0];
    if (data && data.w > 0 && data.h > 0) {
      regionX.value = data.x;
      regionY.value = data.y;
      regionW.value = data.w;
      regionH.value = data.h;
    }
  });
});

onUnmounted(() => {
  if (statusInterval) clearInterval(statusInterval);
});

</script>

<template>
  <div v-if="!isSelectorMode" class="app-container">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
      <h1 style="margin-bottom: 0;">{{ t('title') }} <span style="font-size: 12px; color: #94a3b8">Wails v3</span></h1>
      <button class="btn-icon" style="width: auto; padding: 4px 10px; font-size: 12px; border-radius: 4px;" @click="lang = lang === 'zh' ? 'en' : 'zh'">{{ t('langToggle') }}</button>
    </div>
    
    <div class="form-group">
      <label>{{ t('saveDir') }}</label>
      <div class="input-with-button">
        <input type="text" v-model="saveDirectory" :placeholder="t('saveDirHint')" readonly style="cursor: default;" />
        <button class="btn-icon" @click="handleSelectDirectory">{{ t('browse') }}</button>
      </div>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>{{ t('startDate') }}</label>
        <div class="date-input-wrapper">
          <input type="date" v-model="startDate" :class="{ 'empty-date': !startDate }" />
          <span class="date-placeholder" v-if="!startDate">{{ t('datePlaceholder') }}</span>
        </div>
      </div>
      <div class="form-group">
        <label>{{ t('endDate') }}</label>
        <div class="date-input-wrapper">
          <input type="date" v-model="endDate" :class="{ 'empty-date': !endDate }" />
          <span class="date-placeholder" v-if="!endDate">{{ t('datePlaceholder') }}</span>
        </div>
      </div>
    </div>
    
    <div class="form-row">
      <div class="form-group">
        <label>{{ t('startTime') }}</label>
        <input type="time" v-model="dailyStartTime" />
      </div>
      <div class="form-group">
        <label>{{ t('endTime') }}</label>
        <input type="time" v-model="dailyEndTime" />
      </div>
      <div class="form-group">
        <label>{{ t('interval') }}</label>
        <input type="number" v-model="intervalMinutes" min="1" />
      </div>
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>{{ t('region') }}</label>
        <select v-model="mode" @change="handleModeChange">
          <option value="fullscreen">{{ t('fullscreen') }}</option>
          <option value="region">{{ t('regionSelect') }}</option>
        </select>
      </div>
      <div class="form-group">
        <label>{{ t('format') }}</label>
        <select v-model="format">
          <option value="png">{{ t('pngDesc') }}</option>
          <option value="jpg">{{ t('jpgDesc') }}</option>
        </select>
      </div>
    </div>

    <div v-if="mode === 'region'" class="form-group">
      <label>{{ t('selectedRegion') }}</label>
      <div class="input-with-button">
        <input type="text" :value="`X:${regionX} Y:${regionY} W:${regionW} H:${regionH}`" readonly style="cursor: default;" />
        <button class="btn-icon" @click="handleOpenRegionSelector">{{ t('reselect') }}</button>
      </div>
    </div>

    <div class="form-row">
      <label class="checkbox-group">
        <input type="checkbox" v-model="autoStart" />
        <span>{{ t('autoStart') }}</span>
      </label>
      <label class="checkbox-group" v-if="autoStart">
        <input type="checkbox" v-model="showUIOnAutoStart" />
        <span>{{ t('showUiAuto') }}</span>
      </label>
    </div>

    <div class="actions">
      <button class="btn-secondary" @click="saveConfig(false)">{{ t('saveSettings') }}</button>
      <button :class="isRunning ? 'btn-danger' : 'btn-primary'" @click="toggleScheduler">
        {{ isRunning ? t('stopAuto') : t('startAuto') }}
      </button>
    </div>

    <div class="status-text">
      <span class="status-indicator" :class="{ running: isRunning }"></span>
      {{ isRunning ? t('statusRunning') : t('statusStopped') }}
    </div>

    <!-- Custom Modal Overlay -->
    <div v-if="alertMessage" class="modal-overlay">
      <div class="modal-content">
        <p>{{ alertMessage }}</p>
        <div class="modal-actions">
          <button class="btn-primary" @click="alertMessage = ''">{{ t('ok') }}</button>
        </div>
      </div>
    </div>
  </div>

  <div v-if="isSelectorMode" class="selector-overlay" @mousedown="onMouseDown" @mousemove="onMouseMove" @mouseup="onMouseUp">
    <div v-if="isDragging" class="selector-box" :style="getBoxStyle()"></div>
  </div>
</template>
