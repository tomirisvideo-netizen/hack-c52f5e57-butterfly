(() => {
  'use strict';

  const $ = (selector, root = document) => root.querySelector(selector);
  const escape = (value) => String(value ?? '').replace(/[&<>"']/g, (char) => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[char]));
  const icons = {
    grid:'<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>',
    file:'<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z"/><path d="M14 2v6h6M8 13h8M8 17h5"/>',
    plus:'<path d="M12 5v14M5 12h14"/>',
    chevron:'<path d="m9 5 7 7-7 7"/>',
    down:'<path d="m6 9 6 6 6-6"/>',
    arrow:'<path d="M5 12h14m-5-5 5 5-5 5"/>',
    check:'<path d="m5 12 4 4L19 6"/>',
    circleCheck:'<circle cx="12" cy="12" r="9"/><path d="m8 12 3 3 5-6"/>',
    clock:'<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
    calendar:'<rect x="3" y="5" width="18" height="16" rx="2"/><path d="M16 3v4M8 3v4M3 11h18M8 15h2"/>',
    people:'<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2m20 0v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/><circle cx="9" cy="7" r="4"/>',
    person:'<circle cx="12" cy="8" r="4"/><path d="M5 21v-2a7 7 0 0 1 14 0v2"/>',
    upload:'<path d="M12 16V3m-5 5 5-5 5 5M4 16v4a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-4"/>',
    download:'<path d="M12 3v13m-5-5 5 5 5-5M4 16v4a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-4"/>',
    mic:'<rect x="9" y="2" width="6" height="13" rx="3"/><path d="M5 10v2a7 7 0 0 0 14 0v-2M12 19v3m-3 0h6"/>',
    quote:'<path d="M10 6H3v7h5c0 4-2 5-4 5m17-12h-7v7h5c0 4-2 5-4 5"/>',
    sparkle:'<path d="m12 3 2.7 6.3L21 12l-6.3 2.7L12 21l-2.7-6.3L3 12l6.3-2.7L12 3Z"/>',
    settings:'<path d="m9 3-.6 2.3-2 .9-2.2-.7-2 3.5 1.6 1.7v2.6L2.2 15l2 3.5 2.2-.7 2 .9L9 21h4l.6-2.3 2-.9 2.2.7 2-3.5-1.6-1.7v-2.6L19.8 9l-2-3.5-2.2.7-2-.9L13 3Z"/><circle cx="11" cy="12" r="3"/>',
    help:'<circle cx="12" cy="12" r="9"/><path d="M9.5 8a2.6 2.6 0 0 1 5 .9c0 1.9-2.5 2-2.5 4.1m0 3h.01"/>',
    leaf:'<path d="M20 3c-2 3-8-1-13 4a7 7 0 0 0 10 10c4-4 1-9 3-14ZM4 21l10-10"/>',
    shield:'<path d="M12 3 3 7v5c0 5 9 9 9 9s9-4 9-9V7Z"/><path d="m8 12 3 3 5-6"/>',
    close:'<path d="m6 6 12 12M6 18 18 6"/>',
    menu:'<path d="M4 6h16M4 12h16M4 18h16"/>',
    telegram:'<path d="m21 3-4 18-6-5-4 3 1-7L21 3 3 10l5 2m3 4 5-8"/>',
    refresh:'<path d="M20 7v5h-5M4 17v-5h5M6.2 6a8 8 0 0 1 13 3M4.8 15a8 8 0 0 0 13 3"/>',
    globe:'<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3c4 4 4 14 0 18-4-4-4-14 0-18Z"/>',
    activity:'<path d="M2 12h5l3-8 4 16 3-8h5"/>',
    play:'<path d="m8 4 12 8-12 8Z"/>',
    save:'<path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h12l4 4v12a2 2 0 0 1-2 2Z"/><path d="M7 3v6h10V3M7 21v-8h10v8"/>',
  };
  const icon = (name, extra = '') => `<svg class="icon ${extra}" viewBox="0 0 24 24" aria-hidden="true">${icons[name] || icons.file}</svg>`;
  const butterfly = (extra = '') => `<svg class="butterfly ${extra}" viewBox="0 0 64 64" aria-hidden="true"><path d="M30.8 29.7C19.6 13.3 7.8 7.1 6.8 21.3 5.9 33.7 19 35.7 26.4 35.1 10.5 42.4 14.2 55.5 23.5 51.4c5.3-2.4 7.5-10.2 8.5-14.5 1.1 4.3 3.3 12.1 8.6 14.5 9.3 4.1 13-9-2.9-16.3 7.4.6 20.5-1.4 19.6-13.8-1-14.2-12.8-8-24 8.4l-.4-5.3c-.1-1.2-1.7-1.2-1.8 0l-.3 5.3Z"/></svg>`;
  const statuses = {
    awaiting_provider: {label:'Ожидает подключение',tone:'amber',description:'Аудио сохранено. Для расшифровки подключите распознавание речи. Настройка выполняется на сервере.'},
    queued: {label:'В очереди',tone:'amber',description:'Встреча в очереди обработки. Результаты появятся здесь после ответа сервера.'},
    extracting: {label:'Обрабатывается',tone:'amber',description:'Сервер обрабатывает встречу. Можно оставить страницу открытой — статус обновится автоматически.'},
    processing: {label:'Обрабатывается',tone:'amber',description:'Сервер обрабатывает встречу. Можно оставить страницу открытой — статус обновится автоматически.'},
    needs_review: {label:'На проверке',tone:'amber',description:''},
    completed: {label:'Завершена',tone:'green',description:''},
    failed: {label:'Ошибка обработки',tone:'red',description:'Не удалось завершить обработку. Исходные данные сохранены; можно повторить попытку.'},
  };
  const languageNames = {auto:'Автоопределение',ru:'Русский',kk:'Қазақша',mixed:'Русский + қазақша'};
  const processing = new Set(['queued','extracting','processing']);
  let storedApi = '';
  try { storedApi = localStorage.getItem('butterfly_api_url') || ''; } catch (_) { /* Storage can be disabled. */ }
  const state = {api:storedApi.replace(/\/+$/,''),meetings:[],selected:null,draft:null,health:null,loading:true,loadingMeeting:false,error:'',tab:'actions',mobileOpen:false,exportOpen:false,busy:false,dirty:false,view:'meetings',notifications:{},meetingFilter:'all',meetingSearch:''};
  let pollTimer = null;
  let toastTimer = null;
  let loadVersion = 0;
  let connectVersion = 0;
  let audioObjectUrl = null;

  function formatDate(value, options = {day:'numeric',month:'long'}) {
    // A meeting date is a calendar day; only timestamps should be converted from UTC.
    const date = new Date(/^\d{4}-\d{2}-\d{2}$/.test(value) ? `${value}T00:00:00` : value || Date.now());
    return Number.isNaN(date.getTime()) ? String(value || '') : date.toLocaleDateString('ru-RU',options);
  }
  function timecode(value) {
    const total = Math.max(0,Math.floor(Number(value) || 0));
    return `${String(Math.floor(total/60)).padStart(2,'0')}:${String(total%60).padStart(2,'0')}`;
  }
  function statusFor(value) { return statuses[value] || {label:'Сохранена',tone:'',description:''}; }
  function initials(name) { return String(name || '?').split(/\s+/).filter(Boolean).slice(0,2).map((part)=>part[0]).join('').toUpperCase(); }
  function url(path) {
    if (/^https?:\/\//i.test(path)) return path;
    return `${state.api}${path.startsWith('/')?'':'/'}${path}`;
  }
  async function request(path, options = {}) {
    const controller = new AbortController();
    const timer = setTimeout(()=>controller.abort(),options.body instanceof FormData ? 120000 : 30000);
    try {
      const response = await fetch(url(path),{credentials:'include',...options,signal:controller.signal,headers:{...(options.body && !(options.body instanceof FormData) ? {'Content-Type':'application/json'} : {}),...options.headers}});
      if (!response.ok) {
        let detail = '';
        try { const body = await response.json(); detail = typeof body.detail === 'string' ? body.detail : (body.message || body.error || ''); } catch (_) { /* Keep a useful HTTP fallback. */ }
        throw new Error(detail ? String(detail).slice(0,450) : `Сервер вернул ошибку ${response.status}. Попробуйте ещё раз.`);
      }
      if (options.blob) return response.blob();
      if (response.status === 204) return null;
      return response.json();
    } catch (error) {
      if (error.name === 'AbortError') throw new Error('Сервер отвечает дольше обычного. Проверьте соединение и повторите запрос.');
      if (error instanceof TypeError) throw new Error('Не удалось связаться с сервером. Проверьте адрес API в настройках и подключение к сети.');
      throw error;
    } finally { clearTimeout(timer); }
  }
  function notify(message, error = false) {
    const el = $('#toast');
    el.textContent = message;
    el.className = `visible${error?' error':''}`;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(()=>{el.className='';},5500);
  }
  function copyMeeting(meeting) {
    return {...meeting,transcript:Array.isArray(meeting.transcript)?meeting.transcript:[],actions:Array.isArray(meeting.actions)?meeting.actions.map((item)=>({...item})):[],events:Array.isArray(meeting.events)?meeting.events:[]};
  }
  function updateList(meeting) {
    const index = state.meetings.findIndex((item)=>item.id===meeting.id);
    if (index>=0) state.meetings[index] = {...state.meetings[index],...meeting};
    else state.meetings.unshift(meeting);
  }
  async function connect() {
    const version = ++connectVersion;
    state.loading=true; state.error=''; render();
    try {
      await request('/api/session',{method:'POST'});
      if(version!==connectVersion) return;
      const results = await Promise.allSettled([request('/api/meetings'),request('/health')]);
      if(version!==connectVersion) return;
      if (results[0].status==='rejected') throw results[0].reason;
      const meetings = results[0].value;
      state.meetings=Array.isArray(meetings)?meetings:(meetings.meetings || []);
      state.health=results[1].status==='fulfilled'?results[1].value:null;
      state.loading=false; render();
      // Opening the workspace shows the meeting overview; samples are never auto-created.
    } catch(error) { if(version!==connectVersion) return;state.loading=false; state.error=error.message; render(); }
  }
  async function loadMeeting(id, quiet = false) {
    const version = ++loadVersion;
    clearTimeout(pollTimer);
    if (!quiet) { state.view='meetings'; state.loadingMeeting=true; state.mobileOpen=false; state.exportOpen=false; render(); }
    try {
      const meeting = copyMeeting(await request(`/api/meetings/${encodeURIComponent(id)}`));
      if (version!==loadVersion) return;
      const changed = state.selected?.id!==meeting.id;
      state.selected=meeting;
      if (!state.dirty || changed) { state.draft=copyMeeting(meeting); state.dirty=false; }
      if (changed) { state.tab='actions'; if (audioObjectUrl) URL.revokeObjectURL(audioObjectUrl); audioObjectUrl=null; }
      updateList(meeting);
      state.loadingMeeting=false; state.error=''; render();
      if (processing.has(meeting.status) || meeting.status==='awaiting_provider') schedulePoll(meeting.id);
      if (meeting.audio_url && !audioObjectUrl) loadAudio(meeting);
    } catch(error) {
      if (version!==loadVersion) return;
      state.loadingMeeting=false;
      if (quiet) { notify(error.message,true); schedulePoll(id,10000); }
      else { state.error=error.message; render(); }
    }
  }
  function schedulePoll(id,delay = 4000) {
    clearTimeout(pollTimer);
    pollTimer=setTimeout(()=>{if(state.selected?.id===id && state.view==='meetings' && !state.busy && !document.hidden) loadMeeting(id,true);else if(state.selected?.id===id) schedulePoll(id,10000);},delay);
  }
  async function loadAudio(meeting) {
    try {
      const blob = await request(meeting.audio_url,{blob:true});
      if (state.selected?.id!==meeting.id) return;
      audioObjectUrl=URL.createObjectURL(blob);
      const audio=$('#meeting-audio');
      if(audio) audio.src=audioObjectUrl;
    } catch (_) { /* The transcript remains available when audio cannot be loaded. */ }
  }

  function library() {
    return `<aside class="library ${state.mobileOpen?'mobile-open':''}" aria-label="Рабочее пространство"><button class="wordmark" data-action="home" aria-label="Butterfly — главная">${butterfly()}<span>Butterfly</span></button><button class="icon-btn mobile-close" data-action="close-menu" aria-label="Закрыть меню">${icon('close')}</button><p class="workspace-label">Рабочее пространство</p><nav class="sidebar-nav" aria-label="Основная навигация"><button class="sidebar-link" data-action="home">${icon('grid')}<span>Обзор</span></button><button class="sidebar-link ${state.view==='meetings'?'active':''}" data-action="home">${icon('calendar')}<span>Встречи</span></button><button class="sidebar-link" data-action="tasks">${icon('circleCheck')}<span>Поручения</span></button><button class="sidebar-link" data-action="participants">${icon('people')}<span>Участники</span></button></nav><button class="btn primary new-meeting" data-action="new">${icon('plus')}<span>Новая встреча</span></button><button class="library-demo ${state.view==='fly'?'active':''}" data-action="fly-demo">${icon('activity')}<span>Fly demo<small>Эксперимент плагина</small></span>${icon('arrow')}</button><div class="library-footer"><button class="sidebar-link" data-action="settings">${icon('settings')}<span>Настройки</span></button><div class="workspace-account"><span class="avatar">BT</span><span>Моё пространство</span></div></div></aside>`;
  }
  function render() {
    const configured=state.health?.provider?.llm_configured;
    const mcp=state.health?.mcp;
    const mcpLabel=mcp?.connected?'Инструменты подключены':mcp?.configured?'Инструменты недоступны':'Инструменты не подключены';
    const connectionText=state.error?'Нет соединения':state.health?(configured?'Модель настроена':'Пространство подключено'):'Подключение к пространству';
    $('#app').innerHTML=`<div class="app-shell">
      <div class="mobile-shade ${state.mobileOpen?'visible':''}" data-action="close-menu"></div>${library()}
      <main class="main"><header class="topbar"><div class="breadcrumb"><button class="icon-btn menu-toggle" data-action="menu" aria-label="Открыть список встреч">${icon('menu')}</button><span>Рабочее пространство</span>${icon('chevron')}<strong>${state.view==='fly'?'Fly demo':state.selected?escape(state.selected.title):'Встречи'}</strong></div><div class="topbar-right">${mcp?`<button class="connection mcp-connection" data-action="settings" aria-label="${escape(mcpLabel)}" title="MCP: ${escape(mcpLabel)}"><span class="status-dot ${mcp.connected?'green':mcp.configured?'amber':''}"></span>${icon('activity')}<span class="mcp-label">${escape(mcpLabel)}</span></button>`:''}<button class="connection" data-action="settings" title="Настройки подключения"><span class="status-dot ${state.error?'red':state.health?'green':''}"></span><span class="connection-label">${escape(connectionText)}</span></button><button class="icon-btn" data-action="settings" aria-label="Настройки подключения">${icon('settings')}</button></div></header>
      <div class="main-content">${state.error?`<div class="error-banner" role="alert">${icon('help')}<p>${escape(state.error)}</p><button class="btn small secondary" data-action="reconnect">Повторить</button></div>`:''}${state.view==='fly'?flyDemoPage():state.loading || state.loadingMeeting?`<div class="loading-content"><div><div class="spinner"></div><p>${state.loadingMeeting?'Открываем встречу…':'Подключаем пространство…'}</p></div></div>`:state.selected?meetingPage():welcome()}<footer class="page-footer"><span class="footer-brand">butterfly · из разговора в ясный план</span><span>Создано для осмысленных встреч</span></footer></div></main>
    </div>`;
    if(audioObjectUrl && $('#meeting-audio')) $('#meeting-audio').src=audioObjectUrl;
    if(state.busy) $('#app').querySelectorAll('button, input, textarea').forEach((control)=>{control.disabled=true;});
  }
  // The embedded Fly experience comes from the existing plugin. It is an
  // interaction experiment, not evidence of ASR accuracy or a clinical model.
  function flyDemoPage() {
    return `<section class="fly-demo-page" aria-labelledby="fly-demo-title"><div class="page-heading"><div class="heading-copy"><p class="eyebrow">Лаборатория / Эксперимент</p><h1 id="fly-demo-title">Fly demo</h1><p class="fly-demo-description">Исследуйте интерактивное демо из плагина HackAlem.</p></div><button class="btn secondary" data-action="close-fly">${icon('close')} Вернуться к встречам</button></div><div class="notice fly-experiment">${icon('sparkle')}<span>Экспериментальная визуализация плагина. Она демонстрирует взаимодействие с Fly; качество распознавания речи проверяется отдельно.</span></div><div class="panel fly-demo-frame"><iframe src="${escape(url('/fly-demo'))}" title="Fly demo — интерактивный эксперимент плагина HackAlem" referrerpolicy="same-origin" allow="fullscreen"></iframe></div><p class="fly-demo-help">Если демо недоступно, проверьте подключение инструментов в настройках. <button class="fly-link" data-action="settings">Открыть настройки ${icon('arrow')}</button></p></section>`;
  }
  function providerCard() {
    const provider=state.health?.provider;
    const configured=provider?.llm_configured===true;
    const providerName=provider?.name || provider?.model || (configured?'Модель для протокола':'NVIDIA Nemotron');
    return `<section class="provider-card" aria-label="Провайдер обработки">${icon('activity')}<div><h3>${escape(providerName)}</h3><p>${configured?'Провайдер настроен на сервере':'Планируемый провайдер'}</p><span class="badge ${configured?'green':''}">${configured?'Настроен':provider?'Не подключён':'Статус неизвестен'}</span></div></section>`;
  }
  function meetingTable() {
    const filtered=state.meetings.filter((meeting)=>(state.meetingFilter==='all' || (state.meetingFilter==='review'?meeting.status==='needs_review':meeting.status==='completed')) && String(meeting.title || '').toLocaleLowerCase('ru').includes(state.meetingSearch.toLocaleLowerCase('ru')));
    return filtered.length?`<div class="meetings-table" aria-label="Последние встречи"><div class="meeting-row meeting-table-head"><span>Встреча</span><span>Дата</span><span>Поручения</span><span>Статус</span><span aria-hidden="true"></span></div>${filtered.map((meeting)=>{const status=statusFor(meeting.status);return `<button class="meeting-row" data-action="select" data-id="${escape(meeting.id)}" aria-label="Открыть встречу: ${escape(meeting.title)}"><span class="meeting-row-title">${icon('file')}<span>${escape(meeting.title || 'Встреча без названия')}${meeting.mode==='demo'?'<small>Синтетическое демо</small>':''}</span></span><span class="meeting-row-date">${escape(formatDate(meeting.date || meeting.created_at))}</span><span class="meeting-row-count">${Array.isArray(meeting.actions)?meeting.actions.length:Number.isFinite(meeting.action_count)?meeting.action_count:'—'}</span><span><span class="badge ${status.tone}">${escape(status.label)}</span></span>${icon('chevron')}</button>`;}).join('')}</div>`:`<div class="panel empty-meetings">${icon('calendar')}<h3>${state.meetings.length?'Встречи не найдены':'Здесь появятся ваши встречи'}</h3><p>${state.meetings.length?'Измените запрос или выберите другой статус.':'Загрузите первую запись или откройте синтетическое демо.'}</p></div>`;
  }
  function welcome() {
    return `<section class="overview-page"><div class="page-heading"><div class="heading-copy"><h1>От разговора к результату</h1><p class="page-description">Загрузите запись. Проверьте поручения. Сохраните протокол.</p></div></div><div class="onboarding-grid"><section class="panel upload-launcher" aria-labelledby="upload-title">${icon('upload')}<h2 id="upload-title">Новая встреча</h2><div class="quick-dropzone" id="quick-dropzone"><strong>Перетащите запись сюда</strong><p>Аудиозапись · обработка на сервере проекта</p><input class="sr-only" type="file" id="quick-file" accept="audio/*,.mp3,.wav,.m4a,.ogg,.webm,.mp4" aria-label="Выбрать запись встречи" /></div><div class="quick-upload-actions"><label class="btn primary" for="quick-file" tabindex="0">${icon('file')} Выбрать файл</label><button class="text-button" data-action="demo" ${state.busy?'disabled':''}>${state.busy?'Открываем демо…':'Открыть демо'}</button><button class="text-button" data-action="new">Вставить текст</button></div><label class="form-field quick-title"><span>Название встречи</span><input id="quick-title" maxlength="200" placeholder="Например, запуск городского сервиса" /></label></section><aside class="how-it-works"><h2>Как это работает</h2><div class="how-step"><span>01</span><div><h3>Загрузите запись</h3><p>Аудио или готовая расшифровка вашей встречи.</p></div></div><div class="how-step"><span>02</span><div><h3>Проверьте результат</h3><p>Уточните поручения, ответственных и сроки по стенограмме.</p></div></div><div class="how-step"><span>03</span><div><h3>Скачайте протокол</h3><p>Подтвердите результат и поделитесь им с командой.</p></div></div>${providerCard()}</aside></div><section class="recent-meetings" aria-labelledby="recent-title"><div class="section-heading"><h2 id="recent-title">Последние встречи</h2><label class="meeting-search">${icon('file')}<input type="search" id="meeting-search" value="${escape(state.meetingSearch)}" placeholder="Найти встречу" aria-label="Найти встречу" /></label></div><div class="meeting-filters" role="group" aria-label="Фильтр встреч">${[['all','Все'],['review','На проверке'],['completed','Готовые']].map(([value,label])=>`<button class="tab ${state.meetingFilter===value?'active':''}" data-action="filter" data-filter="${value}" aria-pressed="${state.meetingFilter===value}">${label}</button>`).join('')}</div><div id="meetings-table-container">${meetingTable()}</div></section><p class="welcome-note">${icon('shield')}<span>Демо содержит вымышленные данные. Статус модели определяется сервером; аудио обрабатывается после подключения распознавания речи.</span></p></section>`;
  }
  function meetingPage() {
    const meeting=state.selected,draft=state.draft || meeting,status=statusFor(meeting.status);
    const speakers=[...new Set(meeting.transcript.map((item)=>item.speaker).filter(Boolean))];
    const duration=Math.max(0,...meeting.transcript.map((item)=>Number(item.end) || 0));
    const confirmed=draft.actions.filter((action)=>action.status==='confirmed').length;
    const canComplete=meeting.status==='needs_review' && draft.actions.every((action)=>action.status==='confirmed');
    return `<section class="meeting-detail" aria-label="Протокол встречи"><div class="page-heading"><div class="heading-copy"><h1>${escape(meeting.title || 'Встреча без названия')}</h1><div class="meeting-meta"><span>${escape(formatDate(meeting.date || meeting.created_at,{day:'numeric',month:'long',year:'numeric'}))}</span>${duration?`<span>·</span><span>${timecode(duration)}</span>`:''}<span>·</span><span>${escape(languageNames[meeting.language] || meeting.language || 'Язык не указан')}</span><span class="badge ${status.tone}">${escape(status.label)}</span></div></div><div class="heading-actions"><div class="export-menu"><button class="btn secondary" data-action="export-menu" aria-expanded="${state.exportOpen}" aria-haspopup="true">${icon('upload')} Экспорт</button>${state.exportOpen?`<div class="export-options" role="group" aria-label="Формат экспорта">${['pdf','docx','json'].map((format)=>`<button data-action="export" data-format="${format}" ${format!=='json' && (meeting.status!=='completed' || state.dirty)?'disabled title="Сначала подтвердите протокол"':''}>${icon('file')}Скачать ${format.toUpperCase()}</button>`).join('')}${meeting.status!=='completed'?'<p class="export-hint">PDF и DOCX доступны после проверки.</p>':''}</div>`:''}</div></div></div>${meeting.mode==='demo'?`<div class="notice demo-notice">${icon('sparkle')}<span><strong>Демо · синтетическая встреча.</strong> Вымышленные данные показывают сценарий проверки. Реальная модель не использовалась.</span></div>`:''}${status.description?`<div class="notice ${meeting.status==='failed'?'error':'processing'}">${icon(meeting.status==='failed'?'help':'clock')}<span>${escape(meeting.error || status.description)}</span>${['failed','awaiting_provider'].includes(meeting.status)?`<button class="btn small secondary" data-action="retry" ${state.busy?'disabled':''}>${icon('refresh')}Повторить</button>`:''}</div>`:''}<div class="meeting-stats"><div><strong>${draft.actions.length}</strong><span>поручений</span></div><div><strong>${draft.actions.length-confirmed}</strong><span>требуют проверки</span></div><div><strong>${speakers.length}</strong><span>участников</span></div></div><div class="tabs meeting-tabs" role="tablist" aria-label="Материалы встречи">${[['actions','Поручения'],['transcript','Стенограмма'],['summary','Краткое содержание'],['events','История']].map(([id,label])=>`<button class="tab ${state.tab===id?'active':''}" role="tab" aria-selected="${state.tab===id}" data-action="tab" data-tab="${id}">${label}${id==='actions'?` <span class="tab-count">${draft.actions.length}</span>`:''}</button>`).join('')}</div><div class="meeting-grid"><div class="meeting-primary">${state.tab==='actions'?`<div class="review-guidance">${icon('file')}<div><h2>Проверьте ответственных и сроки</h2><p>Проверьте предложения по исходным репликам. Вы принимаете решение.</p></div></div><div class="actions-list" role="tabpanel" aria-label="Поручения">${draft.actions.length?draft.actions.map((action,index)=>actionCard(action,index,meeting.status)).join(''):`<div class="panel empty-small">${icon('file')}<p>${processing.has(meeting.status) || meeting.status==='awaiting_provider'?'Поручения появятся после обработки.':'Поручений пока нет. Добавьте решение из разговора.'}</p></div>`}</div><div class="review-footer"><button class="add-action btn secondary" data-action="add-action">${icon('plus')} Добавить поручение</button><div class="review-tools">${state.dirty?'<span class="save-indicator">Изменено</span>':''}<button class="btn secondary" data-action="save" ${state.busy || !state.dirty?'disabled':''}>${icon('save')} Сохранить</button></div></div><div class="complete-protocol"><p>${meeting.status==='completed'?'Протокол подтверждён. PDF и DOCX готовы к экспорту.':'Подтвердите каждое поручение, чтобы завершить протокол.'}</p>${meeting.status==='completed'?`<span class="badge green">${icon('circleCheck')} Подтверждено</span>`:`<button class="btn primary" data-action="complete" ${!canComplete || state.busy?'disabled':''}>${icon('check')} Подтвердить протокол</button>`}</div>`:`<div class="panel transcript-panel">${state.tab==='transcript'?transcriptContent(meeting):state.tab==='summary'?summaryContent(draft):eventsContent(meeting)}</div>`}</div><aside class="context-column" aria-label="Контекст встречи"><section class="participants-panel" id="participants-panel"><h2>Участники</h2>${speakers.length?speakers.map((speaker)=>`<div class="participant"><span class="speaker-avatar">${escape(initials(speaker))}</span><span>${escape(speaker)}</span></div>`).join(''):'<p class="muted">Участники ещё не определены.</p>'}</section>${providerCard()}${flyPanel(meeting,draft,confirmed)}</aside></div>${meeting.audio_url?`<div class="panel audio-player"><div class="audio-caption">${icon('play')}<div><strong>Запись встречи</strong><span>${meeting.mode==='demo'?'Демо · ':''}${duration?timecode(duration):'Аудиозапись'}</span></div></div><audio id="meeting-audio" controls preload="none" aria-label="Запись встречи"></audio></div>`:''}${meeting.status==='completed' && state.health?.telegram_configured?`<div class="delivery-bar"><span>${icon('telegram')} Отправьте статус и ссылку в подключённый чат.</span><button class="btn secondary" data-action="telegram-notify" ${state.busy || state.notifications[meeting.id]?'disabled':''}>${icon(state.notifications[meeting.id]?'check':'telegram')}${state.notifications[meeting.id]?'Статус отправлен':'Отправить статус в Telegram'}</button></div>`:''}</section>`;
  }
  function transcriptContent(meeting) {
    const speakers=[...new Set(meeting.transcript.map((item)=>item.speaker))];
    return `<div class="transcript" id="transcript-content" role="tabpanel" aria-label="Расшифровка">${meeting.transcript.length?meeting.transcript.map((segment,index)=>`<article class="segment" id="segment-${escape(segment.id)}" tabindex="-1"><div class="speaker-avatar speaker-${speakers.indexOf(segment.speaker)%4}">${escape(initials(segment.speaker || 'Участник'))}</div><div><div class="segment-header"><span class="speaker-name">${escape(segment.speaker || 'Участник')}</span><button class="timestamp" data-action="seek" data-time="${Number(segment.start) || 0}" ${!meeting.audio_url?'disabled':''} aria-label="Перейти к ${timecode(segment.start)}">${timecode(segment.start)}</button></div><p>${escape(segment.text)}</p></div></article>`).join(''):`<div class="empty-small">${icon('mic')}<p>${processing.has(meeting.status)?'Расшифровка готовится. Страница обновится автоматически.':meeting.status==='awaiting_provider'?'Аудио ожидает подключения распознавания речи.':'Расшифровка пока недоступна.'}</p></div>`}</div>`;
  }
  function summaryContent(draft) {
    return `<div class="summary-content" role="tabpanel" aria-label="Резюме"><div class="summary-intro"><h2>Главное из разговора</h2><button class="btn small secondary" data-action="save" ${state.busy || !state.dirty?'disabled':''}>${icon('save')}Сохранить</button></div><label class="sr-only" for="summary-text">Резюме встречи</label><textarea class="summary-text" id="summary-text" data-field="summary" placeholder="Добавьте краткое резюме встречи…">${escape(typeof draft.summary==='string'?draft.summary:JSON.stringify(draft.summary || '',null,2))}</textarea><p class="summary-help">Уточните резюме своими словами. Изменения сохранятся после нажатия «Сохранить».</p></div>`;
  }
  function eventsContent(meeting) {
    return `<div class="timeline" role="tabpanel" aria-label="История обработки">${meeting.events.length?meeting.events.map((event)=>`<article class="timeline-item"><span class="timeline-dot"></span><div><p>${escape(event.message || event.type)}</p><time>${escape(formatDate(event.created_at,{day:'numeric',month:'short',hour:'2-digit',minute:'2-digit'}))}</time></div></article>`).join(''):'<div class="empty-small">Событий обработки пока нет.</div>'}</div>`;
  }
  function flyPanel(meeting,draft,confirmed) {
    const eventCount=meeting.events.length;
    const position=Math.min(4,eventCount);
    const pending=draft.actions.find((action)=>action.status!=='confirmed');
    const missing=pending?[!pending.owner && 'ответственного',!pending.deadline && 'срок'].filter(Boolean):[];
    let message='Показываю события сервера и помогаю проверить протокол.';
    if (pending) message=missing.length?`В следующей задаче нужно уточнить ${missing.join(' и ')}. Проверьте исходную реплику.`:'Следующая задача готова к проверке. Сверьте формулировку с исходной репликой.';
    else if (draft.actions.length) message='Все задачи проверены. Протокол можно подтвердить и сохранить.';
    else if (processing.has(meeting.status)) message='Жду результат обработки. Новые события появятся в истории.';
    else if (meeting.status==='awaiting_provider') message='Встреча сохранена. Обработка начнётся после подключения провайдера.';
    if(meeting.status==='completed') message='Протокол подтверждён. Вы можете экспортировать его и вернуться к любой цитате.';
    return `<section class="panel fly-panel" aria-label="Ход проверки"><div class="fly-header"><h2 class="fly-name">${butterfly()} Ход проверки</h2><span class="fly-live"><span class="status-dot ${eventCount?'green':''}"></span> ${eventCount} событий</span></div><div class="fly-visual" aria-hidden="true"><svg class="fly-path" viewBox="0 0 260 50" preserveAspectRatio="none"><path d="M5 25C37 25 38 6 67 13s35 26 63 20 37-24 63-18 39 20 62 10" fill="none" stroke="currentColor" stroke-dasharray="3 4"/>${[5,67,130,193,255].map((x,index)=>`<circle class="fly-node ${index<eventCount?'visited':''}" cx="${x}" cy="${[25,13,33,15,25][index]}" r="3"/>`).join('')}</svg><div class="fly-butterfly ${processing.has(meeting.status)?'processing':''}" style="--fly-position:${position}">${butterfly()}</div></div><h3>${pending?'Давайте проверим детали':meeting.status==='completed'?'Всё на своих местах':'Из разговора — в действие'}</h3><p>${escape(message)}</p><div class="fly-footer"><span>${draft.actions.length?`${confirmed} из ${draft.actions.length} задач проверено`:'Движение по событиям, не прогноз'}</span><button class="fly-link" data-action="${pending?'next-review':'events'}">${pending?'К задаче':'История'} ${icon('arrow')}</button></div></section>`;
  }
  function actionCard(action,index,meetingStatus) {
    const isConfirmed=action.status==='confirmed';
    const missing=[!String(action.owner || '').trim() && 'ответственный',!String(action.deadline || '').trim() && 'срок'].filter(Boolean);
    const segment=state.selected.transcript.find((item)=>item.id===action.segment_id);
    return `<article class="action-card ${isConfirmed?'confirmed':'pending'}" id="action-${escape(action.id)}"><div class="action-card-number">${String(index+1).padStart(2,'0')}</div><div class="action-card-content"><div class="action-heading"><span class="badge ${isConfirmed?'green':'amber'}">${icon(isConfirmed?'circleCheck':'help')}${isConfirmed?'Подтверждено':'Нужна проверка'}</span><label class="sr-only" for="action-text-${index}">Формулировка поручения ${index+1}</label><textarea class="action-text" id="action-text-${index}" rows="1" data-field="text" data-index="${index}" placeholder="Что нужно сделать?">${escape(action.text)}</textarea></div><div class="action-fields"><label class="action-field"><span>${icon('person')} Ответственный</span><input data-field="owner" data-index="${index}" value="${escape(action.owner || '')}" placeholder="Назначить ответственного" aria-label="Ответственный за поручение ${index+1}" /></label><label class="action-field"><span>${icon('calendar')} Срок</span><input type="date" data-field="deadline" data-index="${index}" value="${escape(action.deadline || '')}" placeholder="Указать срок" aria-label="Срок поручения ${index+1}" /></label></div><div class="action-source-row">${action.evidence || action.segment_id?`<button class="evidence" data-action="evidence" data-id="${escape(action.segment_id || '')}" title="Открыть исходную реплику">${icon('quote')}<span>${escape(action.evidence || 'Открыть исходную реплику')}</span>${segment?`<time>${timecode(segment.start)}</time>`:icon('chevron')}</button>`:'<span class="evidence manual-evidence">Добавлено вручную · без привязанной цитаты</span>'}</div><div class="action-footer">${isConfirmed?`<span class="confirmed-label">${icon('check')} Проверено вами</span><button class="btn small secondary" data-action="unconfirm" data-index="${index}" ${state.busy?'disabled':''}>Изменить</button>`:`<span class="action-warning">${missing.length?`Не определены в записи: ${missing.join(', ')}`:'Проверьте формулировку и исходную реплику'}</span><button class="btn primary action-confirm" data-action="confirm-action" data-index="${index}" ${state.busy?'disabled':''}>Подтвердить</button>`}</div></div></article>`;
  }
  function openMeetingDialog(file = null) {
    const prefill=$('#quick-title')?.value || (file?.name?file.name.replace(/\.[^.]+$/,''):'');
    const dialog=$('#meeting-dialog');
    const today=new Date(Date.now()-new Date().getTimezoneOffset()*60000).toISOString().slice(0,10);
    dialog.innerHTML=`<form id="new-meeting-form"><header class="dialog-head"><div><h2 id="meeting-dialog-title">Новая встреча</h2><p>Сохраните разговор. Найдите главное.</p></div><button type="button" class="icon-btn" data-action="close-dialog" aria-label="Закрыть">${icon('close')}</button></header><div class="dialog-content"><label class="form-field"><span>Название встречи</span><input name="title" required maxlength="200" value="${escape(prefill)}" placeholder="Например, планирование команды" autofocus /></label><div class="form-row"><label class="form-field"><span>Дата</span><input type="date" name="date" value="${today}" required /></label><label class="form-field"><span>Язык встречи</span><select name="language"><option value="auto">Определить автоматически</option><option value="ru">Русский</option><option value="kk">Қазақша</option><option value="mixed">Русский + қазақша</option></select></label></div><label class="form-field upload-zone">${icon('upload')}<strong class="upload-file-name">${file?escape(file.name):'Добавьте аудиозапись'}</strong><input type="file" name="file" accept="audio/*,.mp3,.wav,.m4a,.ogg,.webm,.mp4" /><small>MP3, WAV, M4A, OGG, WEBM. Для обработки нужно распознавание речи.</small></label><div class="or-divider">или готовая расшифровка</div><label class="form-field"><span class="sr-only">Текст встречи</span><textarea name="transcript" placeholder="Анна: Давайте подготовим предложение к пятнице.&#10;Марат: Я возьму это на себя."></textarea><small>Выберите один источник: аудио или текст. Если модель не подключена, текст можно проверить вручную.</small></label><div id="meeting-form-error" role="alert"></div></div><footer class="dialog-footer"><button type="button" class="btn ghost" data-action="close-dialog">Отмена</button><button type="submit" class="btn primary">${icon('arrow')} Создать встречу</button></footer></form>`;
    dialog.showModal();
    if(file){const transfer=new DataTransfer();transfer.items.add(file);$('input[name="file"]',dialog).files=transfer.files;}
  }
  function openSettings() {
    const dialog=$('#settings-dialog'),provider=state.health?.provider,mcp=state.health?.mcp;
    dialog.innerHTML=`<form id="settings-form"><header class="dialog-head"><div><h2 id="settings-dialog-title">Настройки пространства</h2><p>Подключение и интеграции</p></div><button type="button" class="icon-btn" data-action="close-dialog" aria-label="Закрыть">${icon('close')}</button></header><div class="dialog-content"><label class="form-field"><span>Адрес API</span><input name="api" type="url" value="${escape(state.api)}" placeholder="Текущий сервер" /><small>Оставьте пустым для текущего сервера. Для отдельного фронтенда укажите HTTPS-адрес бэкенда. Встречи доступны в текущей сессии браузера.</small></label><div class="provider-row"><span>Модель для протокола</span><span>${provider?.llm_configured?'Подключена':provider?'Не подключена':'Статус неизвестен'}</span></div><div class="provider-row"><span>Распознавание речи</span><span>${provider?.asr_configured?'Подключено':provider?'Не подключено':'Статус неизвестен'}</span></div><div class="settings-info"><h3>${icon('activity')} Инструменты MCP</h3><div class="provider-row"><span>Подключение</span><span>${mcp?.connected?'Подключено':mcp?.configured?'Настроено, нет соединения':mcp?'Не настроено':'Статус неизвестен'}</span></div>${mcp?.error?`<p class="settings-error">${escape(mcp.error)}</p>`:''}<p>Fly demo использует существующий плагин через сервер пространства.</p></div><div class="settings-info"><h3>${icon('telegram')} Telegram</h3><div class="provider-row"><span>Интеграция</span><span>${state.health?.telegram_configured?'Настроена':'Не настроена'}</span></div><p>После подтверждения протокола нажмите «Отправить статус в Telegram». В подключённый чат уйдут статус и ссылка, без текста встречи. Настройка бота выполняется на сервере.</p></div><div class="settings-info"><h3>${icon('shield')} Ключи остаются на сервере</h3><p>Подключение NVIDIA Nemotron и ASR выполняется через серверную конфигурацию. API-ключи и токен Telegram не нужно вводить в браузере.</p></div><div id="settings-form-error" role="alert"></div></div><footer class="dialog-footer"><button type="button" class="btn ghost" data-action="close-dialog">Закрыть</button><button type="submit" class="btn primary">Сохранить подключение</button></footer></form>`;
    dialog.showModal();
  }
  function focusEvidence(id) {
    if(!id) { notify('У этой задачи нет привязки к реплике. Сверьте цитату с расшифровкой.'); return; }
    state.tab='transcript'; render();
    const segment=document.getElementById(`segment-${id}`);
    if(!segment) { notify('Исходная реплика не найдена в расшифровке.',true); return; }
    segment.scrollIntoView({behavior:'smooth',block:'center'});
    segment.focus({preventScroll:true});segment.classList.add('highlight');
    setTimeout(()=>segment.classList.remove('highlight'),4000);
  }
  async function saveDraft({silent=false} = {}) {
    if(state.busy || !state.selected || !state.draft) return false;
    const id=state.selected.id;
    const actions=state.draft.actions.map((action)=>({...action,text:String(action.text || '').trim(),owner:String(action.owner || '').trim(),deadline:String(action.deadline || '').trim()}));
    if(actions.some((action)=>!action.text)) { notify('Добавьте формулировку для каждой задачи.',true); return false; }
    // Invalidate an earlier poll so its stale response cannot replace the saved draft.
    loadVersion++;
    state.busy=true;render();
    try {
      const result=await request(`/api/meetings/${encodeURIComponent(id)}`,{method:'PATCH',body:JSON.stringify({summary:state.draft.summary || '',actions})});
      const meeting=result?.id?result:await request(`/api/meetings/${encodeURIComponent(id)}`);
      state.selected=copyMeeting(meeting);state.draft=copyMeeting(meeting);state.dirty=false;updateList(meeting);
      if(!silent) notify('Изменения сохранены');
      return true;
    } catch(error) { notify(error.message,true); return false; }
    finally { state.busy=false;render(); }
  }
  async function createDemo() {
    if(state.busy) return;
    if(state.dirty && !(await saveDraft({silent:true}))) return;
    state.busy=true;render();
    try {
      const result=await request('/api/demo',{method:'POST'});
      const meeting=result.meeting || result;
      if(!meeting.id) throw new Error('Сервер не вернул идентификатор демо-встречи.');
      state.dirty=false;await loadMeeting(meeting.id);notify('Открыта синтетическая демо-встреча');
    } catch(error) { notify(error.message,true); }
    finally {state.busy=false;render();}
  }
  async function completeMeeting() {
    if(state.busy || !state.selected) return;
    if(state.draft.actions.some((action)=>action.status!=='confirmed')) {notify('Сначала подтвердите каждую задачу.');return;}
    if(state.dirty && !(await saveDraft({silent:true}))) return;
    state.busy=true;render();
    try {await request(`/api/meetings/${encodeURIComponent(state.selected.id)}/confirm`,{method:'POST'});await loadMeeting(state.selected.id);notify('Протокол подтверждён');}
    catch(error){notify(error.message,true);}
    finally{state.busy=false;render();}
  }
  async function sendTelegramStatus() {
    if(state.busy || state.selected?.status!=='completed' || !state.health?.telegram_configured) return;
    const id=state.selected.id;
    state.busy=true;render();
    try {
      const result=await request(`/api/meetings/${encodeURIComponent(id)}/notify`,{method:'POST'});
      if(result?.status==='sent' || result?.status==='already_sent') {
        state.notifications[id]=true;
        notify(result.status==='already_sent'?'Этот статус уже отправлен':'Статус отправлен в Telegram');
      } else if(result?.status==='queued') notify('Отправка поставлена в очередь. Результат появится в истории.');
      else notify('Отправка пока недоступна. Проверьте интеграцию в настройках.',true);
    } catch(error) {notify(error.message,true);}
    finally {state.busy=false;render();}
  }
  async function exportMeeting(format) {
    if(state.busy || !state.selected) return;
    const {id,title}=state.selected;
    if(format!=='json' && (state.selected?.status!=='completed' || state.dirty)){notify('PDF и DOCX доступны после подтверждения протокола.');return;}
    state.exportOpen=false;render();
    if(state.dirty && !(await saveDraft({silent:true}))) return;
    try {
      const blob=await request(`/api/meetings/${encodeURIComponent(id)}/export?format=${format}`,{blob:true});
      const objectUrl=URL.createObjectURL(blob),link=document.createElement('a');
      link.href=objectUrl;link.download=`${(title || 'butterfly').replace(/[<>:"/\\|?*\u0000-\u001f]/g,'-').slice(0,90)}.${format}`;
      document.body.appendChild(link);link.click();link.remove();setTimeout(()=>URL.revokeObjectURL(objectUrl),10000);
      notify(`${format.toUpperCase()} готов к скачиванию`);
    } catch(error){notify(error.message,true);}
  }
  document.addEventListener('click',async(event)=>{
    const button=event.target.closest('[data-action]');
    if(!button || button.disabled) return;
    const action=button.dataset.action;
    if(state.busy && action!=='close-dialog') return;
    if(action==='new') openMeetingDialog();
    else if(action==='settings' || action==='help') openSettings();
    else if(action==='close-dialog') button.closest('dialog').close();
    else if(action==='menu') {state.mobileOpen=!state.mobileOpen;render();}
    else if(action==='close-menu') {state.mobileOpen=false;render();}
    else if(action==='select') {if(state.dirty && !(await saveDraft({silent:true})))return;await loadMeeting(button.dataset.id);}
    else if(action==='home') {if(state.dirty && !(await saveDraft({silent:true})))return;state.view='meetings';clearTimeout(pollTimer);loadVersion++;state.selected=null;state.draft=null;state.loadingMeeting=false;state.mobileOpen=false;render();}
    else if(action==='fly-demo') {state.view='fly';state.mobileOpen=false;state.exportOpen=false;clearTimeout(pollTimer);render();$('#fly-demo-title')?.scrollIntoView({block:'start'});}
    else if(action==='close-fly') {state.view='meetings';render();if(state.selected && (processing.has(state.selected.status) || state.selected.status==='awaiting_provider'))schedulePoll(state.selected.id);}
    else if(action==='telegram-notify') await sendTelegramStatus();
    else if(action==='tasks') {state.view='meetings';state.mobileOpen=false;if(state.selected){state.tab='actions';render();}else{state.meetingFilter='review';render();$('#recent-title')?.scrollIntoView({behavior:'smooth'});}}
    else if(action==='participants') {if(state.selected){state.view='meetings';state.mobileOpen=false;render();$('#participants-panel')?.scrollIntoView({behavior:'smooth',block:'center'});}else notify('Откройте встречу, чтобы увидеть её участников.');}
    else if(action==='filter') {state.meetingFilter=button.dataset.filter;render();}
    else if(action==='demo') await createDemo();
    else if(action==='reconnect') await connect();
    else if(action==='tab') {state.tab=button.dataset.tab;render();}
    else if(action==='events') {state.tab='events';render();$('.transcript-panel')?.scrollIntoView({behavior:'smooth',block:'start'});}
    else if(action==='evidence') focusEvidence(button.dataset.id);
    else if(action==='next-review') {state.tab='actions';render();const pending=state.draft.actions.find((item)=>item.status!=='confirmed');if(pending){const card=document.getElementById(`action-${pending.id}`);card?.scrollIntoView({behavior:'smooth',block:'center'});card?.querySelector('textarea')?.focus({preventScroll:true});}}
    else if(action==='seek') {const audio=$('#meeting-audio');if(audio?.src){audio.currentTime=Number(button.dataset.time)||0;audio.play().catch(()=>notify('Нажмите кнопку воспроизведения в аудиоплеере.'));}}
    else if(action==='save') await saveDraft();
    else if(action==='add-action') {state.tab='actions';state.draft.actions.push({id:globalThis.crypto?.randomUUID?.() || `manual-${Date.now()}`,text:'',owner:'',deadline:'',evidence:'',segment_id:null,status:'pending'});state.dirty=true;render();const textarea=$('.actions-list .action-card:last-child textarea');textarea?.scrollIntoView({behavior:'smooth',block:'center'});textarea?.focus();}
    else if(action==='confirm-action' || action==='unconfirm') {const item=state.draft.actions[Number(button.dataset.index)];if(item){const before=item.status;item.status=action==='confirm-action'?'confirmed':'pending';state.dirty=true;if(!(await saveDraft({silent:true}))){item.status=before;render();}else notify(action==='confirm-action'?'Задача подтверждена':'Задача открыта для проверки');}}
    else if(action==='complete') await completeMeeting();
    else if(action==='export-menu') {state.exportOpen=!state.exportOpen;render();}
    else if(action==='export') await exportMeeting(button.dataset.format);
    else if(action==='retry') {state.busy=true;render();try{await request(`/api/meetings/${encodeURIComponent(state.selected.id)}/retry`,{method:'POST'});await loadMeeting(state.selected.id);}catch(error){notify(error.message,true);}finally{state.busy=false;render();}}
  });
  document.addEventListener('input',(event)=>{
    if(state.busy) return;
    if(event.target.id==='meeting-search'){state.meetingSearch=event.target.value;$('#meetings-table-container').innerHTML=meetingTable();return;}
    const field=event.target.dataset.field;
    if(!field || !state.draft)return;
    if(field==='summary')state.draft.summary=event.target.value;
    else {const item=state.draft.actions[Number(event.target.dataset.index)];if(item){item[field]=event.target.value;item.status='pending';}}
    state.dirty=true;
    document.querySelectorAll('[data-action="save"]').forEach((button)=>{button.disabled=false;});
    const complete=$('[data-action="complete"]');if(complete)complete.disabled=true;
    const card=event.target.closest('.action-card');
    if(card?.classList.contains('confirmed')){card.classList.remove('confirmed');card.classList.add('pending');const label=$('.confirmed-label',card);if(label)label.textContent='Изменено · нужна проверка';const badge=$('.action-heading .badge',card);if(badge){badge.className='badge amber';badge.textContent='Нужна проверка';}const confirm=$('[data-action="unconfirm"]',card);if(confirm){confirm.dataset.action='confirm-action';confirm.textContent='Подтвердить';confirm.className='btn primary action-confirm';}}
    const tools=$('.review-tools');if(tools && !$('.save-indicator',tools))tools.insertAdjacentHTML('afterbegin','<span class="save-indicator">Изменено</span>');
  });
  document.addEventListener('change',(event)=>{
    if(event.target.id==='quick-file' && event.target.files?.[0]) openMeetingDialog(event.target.files[0]);
    if(event.target.matches('#new-meeting-form input[type="file"]')){const label=$('.upload-file-name',event.target.closest('form'));if(label)label.textContent=event.target.files?.[0]?.name || 'Добавьте аудиозапись';}
  });
  document.addEventListener('dragover',(event)=>{if(event.target.closest('#quick-dropzone')){event.preventDefault();event.target.closest('#quick-dropzone').classList.add('drag-active');}});
  document.addEventListener('dragleave',(event)=>{event.target.closest('#quick-dropzone')?.classList.remove('drag-active');});
  document.addEventListener('drop',(event)=>{if(event.target.closest('#quick-dropzone')){event.preventDefault();event.target.closest('#quick-dropzone').classList.remove('drag-active');if(event.dataTransfer?.files?.[0])openMeetingDialog(event.dataTransfer.files[0]);}});
  document.addEventListener('submit',async(event)=>{
    if(state.busy) {event.preventDefault();return;}
    if(event.target.id==='new-meeting-form') {
      event.preventDefault();
      const form=event.target,data=new FormData(form),file=data.get('file'),transcript=String(data.get('transcript') || '').trim();
      const hasFile=file instanceof File && file.size>0;
      const errorEl=$('#meeting-form-error');
      if(!hasFile && !transcript){errorEl.innerHTML='<p class="form-error">Добавьте аудиофайл или вставьте текст встречи.</p>';return;}
      if(hasFile && transcript){errorEl.innerHTML='<p class="form-error">Выберите один источник: аудиофайл или текст встречи.</p>';return;}
      if(!hasFile)data.delete('file');
      if(!transcript)data.delete('transcript');
      if(state.dirty && !(await saveDraft({silent:true}))) return;
      state.busy=true;render();
      const submit=$('button[type="submit"]',form);submit.disabled=true;submit.textContent='Создаём встречу…';errorEl.innerHTML='';
      try {const result=await request('/api/meetings',{method:'POST',body:data});const meeting=result.meeting || result;if(!meeting.id)throw new Error('Сервер не вернул идентификатор встречи.');$('#meeting-dialog').close();state.dirty=false;await loadMeeting(meeting.id);notify('Встреча сохранена');}
      catch(error){errorEl.innerHTML=`<p class="form-error">${escape(error.message)}</p>`;}
      finally{state.busy=false;render();submit.disabled=false;submit.innerHTML=`${icon('arrow')} Создать встречу`;}
    } else if(event.target.id==='settings-form') {
      event.preventDefault();
      const form=event.target,api=String(new FormData(form).get('api') || '').trim().replace(/\/+$/,'');
      try {if(api){const parsed=new URL(api);if(!['http:','https:'].includes(parsed.protocol) || parsed.username || parsed.password || parsed.search || parsed.hash)throw new Error('Укажите HTTP(S)-адрес API без пароля, параметров и якоря.');}}
      catch(error){$('#settings-form-error').innerHTML=`<p class="form-error">${escape(error.message || 'Не удалось сохранить настройки.')}</p>`;return;}
      if(state.dirty && !(await saveDraft({silent:true}))) return;
      let persistent=true;
      try {localStorage.setItem('butterfly_api_url',api);} catch (_) {persistent=false;}
      clearTimeout(pollTimer);loadVersion++;state.api=api;state.selected=null;state.draft=null;state.loadingMeeting=false;state.dirty=false;state.health=null;state.meetings=[];state.notifications={};
      if(audioObjectUrl) URL.revokeObjectURL(audioObjectUrl);audioObjectUrl=null;
      $('#settings-dialog').close();await connect();
      if(!persistent) notify('Подключение применено для этой страницы. Браузер не разрешает сохранить адрес API.');
    }
  });
  document.addEventListener('keydown',(event)=>{if(event.target.matches('.meeting-tabs .tab') && ['ArrowLeft','ArrowRight','Home','End'].includes(event.key)){event.preventDefault();const tabs=[...document.querySelectorAll('.meeting-tabs .tab')],index=tabs.indexOf(event.target),next=event.key==='Home'?0:event.key==='End'?tabs.length-1:(index+(event.key==='ArrowRight'?1:-1)+tabs.length)%tabs.length;state.tab=tabs[next].dataset.tab;render();document.querySelector('.meeting-tabs [data-tab="'+state.tab+'"]')?.focus();}if((event.key==='Enter' || event.key===' ') && event.target.matches('label[for="quick-file"]')){event.preventDefault();$('#quick-file')?.click();}if(event.key==='Escape' && state.exportOpen){state.exportOpen=false;render();}});
  document.querySelectorAll('dialog').forEach((dialog)=>dialog.addEventListener('click',(event)=>{if(event.target===dialog){const bounds=dialog.getBoundingClientRect();if(event.clientX<bounds.left || event.clientX>bounds.right || event.clientY<bounds.top || event.clientY>bounds.bottom)dialog.close();}}));
  window.addEventListener('beforeunload',(event)=>{if(state.dirty){event.preventDefault();event.returnValue='';}});
  connect();
})();
