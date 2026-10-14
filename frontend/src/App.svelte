<script>
  import { onMount, onDestroy } from 'svelte';
  import { initSettings, settings, saveSettings, refreshStatus } from './lib/store.js';
  import Status from './lib/Status.svelte';

  let cogOpen = false;
  let lastError = '';

  onMount(async () => {
    await initSettings();
    // After a laptop sleep/wake, WebView2's event dispatch can replay
    // or reorder pending clicks. Closing the popover whenever the
    // page becomes visible again guarantees the next interaction
    // starts from a clean slate — see handleWindowClick for the
    // primary defense, this is belt-and-suspenders.
    document.addEventListener('visibilitychange', onVisibility);
  });

  onDestroy(() => {
    if (typeof document !== 'undefined') {
      document.removeEventListener('visibilitychange', onVisibility);
    }
  });

  function onVisibility() {
    if (document.visibilityState === 'visible') {
      cogOpen = false;
    }
  }

  function toggleCog() {
    cogOpen = !cogOpen;
  }

  // Close the popover only when the click lands OUTSIDE the gear
  // button and OUTSIDE the popover itself. Using target.closest()
  // instead of stopPropagation is more robust under Svelte +
  // WebView2, where a window-level listener can still fire after
  // a child handler calls stopPropagation.
  function handleWindowClick(e) {
    if (!cogOpen) return;
    const target = e.target;
    if (target && target.closest && target.closest('.bcog, .popover')) return;
    cogOpen = false;
  }

  async function flush(key, value) {
    lastError = '';
    try {
      await saveSettings({ [key]: value });
    } catch (e) {
      lastError = String(e);
    }
  }

  async function setEnabled(v) {
    await flush('enabled', v);
    await refreshStatus();
  }

  async function swapDirection() {
    await flush('direction', $settings.direction === 't2s' ? 's2t' : 't2s');
  }

  $: rev = $settings.direction === 't2s';
</script>

<svelte:window on:click={handleWindowClick}/>

<main class:rev>
  <header class="bhead">
    <div class="bhead-row">
      <div class="bmark">{rev ? '繁简' : '简繁'}</div>
      <div class="bbrand">
        <div class="t1">{rev ? '繁简转换' : '简繁转换'}</div>
        <div class="t2" class:paused={!$settings.enabled}>
          {$settings.enabled ? '剪贴板中文自动转换' : '已暂停 · 在设置中恢复'}
        </div>
      </div>
      <button
        class="dir-pill"
        class:paused={!$settings.enabled}
        on:click={swapDirection}
        aria-label="切换转换方向（当前：{rev ? '繁转简' : '简转繁'}）"
        title="点击切换转换方向"
      >
        {rev ? '繁 → 简' : '简 → 繁'}
        <svg viewBox="0 0 24 24" width="11" height="11" fill="none"
             stroke="currentColor" stroke-width="2.5" stroke-linecap="round"
             stroke-linejoin="round" aria-hidden="true">
          <path d="M17 1l4 4-4 4"/><path d="M21 5H7"/>
          <path d="M7 23l-4-4 4-4"/><path d="M3 19h14"/>
        </svg>
      </button>
      <button
        class="bcog"
        class:open={cogOpen}
        on:click={toggleCog}
        aria-label="设置"
        aria-expanded={cogOpen}
        aria-haspopup="dialog"
      >
        <svg viewBox="0 0 24 24" width="17" height="17"
             fill="none" stroke="currentColor" stroke-width="2"
             stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          <circle cx="12" cy="12" r="3"/>
        </svg>
      </button>
    </div>
  </header>

  <Status />

  {#if cogOpen}
    <div class="popover" role="dialog" aria-label="设置">
      <div class="pop-title">转换</div>
      <label class="pop-row">
        <span>启用自动转换</span>
        <input type="checkbox" class="toggle" checked={$settings.enabled}
          on:change={(e) => setEnabled(e.target.checked)}/>
      </label>
      <label class="pop-row stacked">
        <span>自动替换剪贴板内容<br/>
          <small>关闭后仅记录结果，不修改剪贴板</small>
        </span>
        <input type="checkbox" class="toggle" checked={$settings.write_back_to_clipboard}
          on:change={(e) => flush('write_back_to_clipboard', e.target.checked)}/>
      </label>
      <div class="pop-title">启动</div>
      <label class="pop-row">
        <span>启动时隐藏主窗口</span>
        <input type="checkbox" class="toggle" checked={$settings.start_minimized}
          on:change={(e) => flush('start_minimized', e.target.checked)}/>
      </label>
      {#if lastError}<div class="err">{lastError}</div>{/if}
    </div>
  {/if}
</main>

<style>
  main {
    position: relative;
    display: flex; flex-direction: column;
    height: 100vh;
    background: var(--bg); color: var(--text-primary);
  }

  /* ---------- gradient top bar ---------- */
  .bhead {
    background: var(--grad);
    padding: 14px 16px 42px;
    color: #fff;
    flex-shrink: 0;
    transition: background 400ms ease;
  }
  main.rev .bhead { background: var(--grad-rev); }

  .bhead-row { display: flex; align-items: center; gap: 10px; }
  .bmark {
    width: 32px; height: 32px; border-radius: 8px;
    background: rgba(255,255,255,0.18);
    border: 1px solid rgba(255,255,255,0.3);
    display: flex; align-items: center; justify-content: center;
    font-size: 12px; font-weight: 700; color: #fff; flex-shrink: 0;
  }
  .bbrand { flex: 1; min-width: 0; }
  .t1 { font-size: 16px; font-weight: 700; letter-spacing: -0.01em; }
  .t2 { font-size: 11px; opacity: 0.85; margin-top: 1px; }
  .t2.paused { opacity: 1; font-weight: 600; color: #FFE9A8; }

  /* direction control: the single header action */
  .dir-pill {
    display: inline-flex; align-items: center; gap: 6px;
    background: #fff; color: var(--blue);
    border: 1px solid #fff;
    border-radius: 999px; padding: 5px 10px 5px 12px;
    font-size: 12px; font-weight: 700; letter-spacing: 0.02em;
    cursor: pointer; flex-shrink: 0; white-space: nowrap;
    box-shadow: 0 2px 8px rgba(0,0,0,0.12);
    transition: transform 150ms, box-shadow 150ms, color 400ms ease, opacity 200ms;
  }
  .dir-pill svg { opacity: 0.75; }
  .dir-pill:hover { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(0,0,0,0.18); }
  .dir-pill:active { transform: translateY(0); }
  .dir-pill.paused { opacity: 0.6; }
  main.rev .dir-pill { color: var(--green); }

  .bcog {
    width: 30px; height: 30px; border-radius: 8px; flex-shrink: 0;
    border: 1px solid transparent; background: transparent; cursor: pointer;
    color: rgba(255,255,255,0.85);
    display: flex; align-items: center; justify-content: center;
    transition: background 150ms;
  }
  .bcog:hover, .bcog.open { background: rgba(255,255,255,0.15); color: #fff; }

  /* ---------- settings popover ---------- */
  .popover {
    position: absolute; top: 52px; right: 12px; z-index: 10;
    background: rgba(255,255,255,0.92);
    backdrop-filter: blur(14px);
    -webkit-backdrop-filter: blur(14px);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: 0 12px 32px rgba(22,24,29,0.18);
    min-width: 248px; padding: 6px;
  }
  .pop-title {
    font-size: 10.5px; font-weight: 700; letter-spacing: .07em;
    color: var(--text-hint); text-transform: uppercase;
    padding: 7px 10px 4px;
  }
  .pop-row {
    display: flex; align-items: center; justify-content: space-between;
    gap: 12px; padding: 8px 10px; border-radius: var(--radius-sm);
    font-size: 13px; cursor: pointer; user-select: none;
  }
  .pop-row:hover { background: var(--surface); }
  .pop-row.stacked { align-items: flex-start; }
  .pop-row small {
    color: var(--text-hint); font-size: 11px; font-weight: 400;
    line-height: 1.5;
  }

  /* iOS-style toggle from a checkbox */
  .toggle {
    appearance: none; -webkit-appearance: none;
    width: 36px; height: 20px; border-radius: 999px;
    background: var(--border); position: relative;
    cursor: pointer; margin: 2px 0 0; flex-shrink: 0;
    transition: background 200ms;
  }
  .toggle:checked { background: var(--green-bright); }
  .toggle::after {
    content: ""; position: absolute; top: 2px; left: 2px;
    width: 16px; height: 16px; border-radius: 50%; background: #fff;
    box-shadow: 0 1px 3px rgba(0,0,0,.2);
    transition: transform 200ms cubic-bezier(0.16, 1, 0.3, 1);
  }
  .toggle:checked::after { transform: translateX(16px); }

  .err {
    color: #DC2626; font-size: 12px;
    padding: 6px 10px;
  }
</style>
