<script>
  import { onMount, onDestroy } from 'svelte';
  import {
    settings, status, history,
    saveSettings, refreshStatus, resetCounter,
    formatTime, subscribeEvents,
  } from './store.js';

  let timer = null;
  let unsubscribe = () => {};
  let expandedId = null;

  onMount(async () => {
    await refreshStatus();
    unsubscribe = subscribeEvents();
    timer = setInterval(refreshStatus, 2000);
  });

  onDestroy(() => {
    unsubscribe();
    if (timer) clearInterval(timer);
  });

  async function reset() {
    await resetCounter();
  }

  function clearHistory() {
    history.set([]);
    expandedId = null;
  }

  function toggleExpand(id) {
    expandedId = expandedId === id ? null : id;
  }

  async function copyText(e, text) {
    e.stopPropagation();
    if (!text) return;
    try {
      await navigator.clipboard.writeText(text);
    } catch (err) {
      console.error('Copy failed:', err);
    }
  }

  function skipReasonText(r) {
    switch (r) {
      case 'empty':      return '空剪贴板';
      case 'unchanged':  return '未变化';
      case 'non-cjk':    return '无中文内容';
      case 'paused':     return '已暂停';
      default:           return r || '—';
    }
  }

  $: rev = $settings.direction === 't2s';
  $: lastEvent = $status.last_event;
  // The card shows the most recent *processed* item: its original
  // text and the outcome — not the live clipboard, which after
  // write-back would just echo the result.
  $: srcText = lastEvent?.source || '';
  $: isConvert = lastEvent?.kind === 'convert' && !!lastEvent.target;
  $: dstText = isConvert ? lastEvent.target : '';
  $: dstHint = !lastEvent
    ? '（等待复制）'
    : lastEvent.kind === 'skip'
      ? `（未转换：${skipReasonText(lastEvent.skip_reason)}）`
      : lastEvent.kind === 'error'
        ? '（转换出错）'
        : '（等待转换）';
</script>

<div class="status" class:rev>

  <!-- Floating glass conversion card (overlaps the gradient header) -->
  <div class="float-card">
    <div class="fc-title">最近转换</div>
    <div class="panes">
      <div class="pane" class:jian={!rev} class:fan={rev} class:empty={!srcText}>
        <span class="tag">原文 · {rev ? '繁体' : '简体'}</span>
        <div class="pane-body">{srcText || '（等待复制）'}</div>
      </div>
      <div class="flow" aria-hidden="true">→</div>
      <div class="pane" class:fan={!rev} class:jian={rev} class:empty={!dstText}>
        <span class="tag">结果 · {rev ? '简体' : '繁体'}</span>
        <div class="pane-body">{dstText || dstHint}</div>
      </div>
    </div>
  </div>

  <!-- History -->
  <div class="hist">
    <div class="hist-head">
      <span>历史记录</span>
      {#if $history.length > 0}
        <button class="clear" on:click={clearHistory} aria-label="清空历史记录">清空</button>
      {/if}
    </div>
    {#if $history.length === 0}
      <div class="hist-empty">尚无转换记录</div>
    {:else}
      <div class="hist-list">
        {#each $history as ev (ev.id)}
          <button
            class="row"
            class:open={expandedId === ev.id}
            on:click={() => toggleExpand(ev.id)}
            aria-expanded={expandedId === ev.id}
          >
            <div class="row-head">
              <span class="t">{formatTime(ev.when)}</span>
              {#if ev.kind === 'convert'}
                <span class="dir" class:t2s={ev.direction === 't2s'}>
                  {ev.direction === 't2s' ? '繁→简' : '简→繁'}
                </span>
                <span class="txt" title={ev.source || ''}>{ev.source || ''}</span>
              {:else if ev.kind === 'skip'}
                <span class="dir neutral">跳过</span>
                <span class="txt muted" title={ev.source || ''}>{ev.source || '（空）'}</span>
              {:else}
                <span class="dir errb">错误</span>
                <span class="txt muted">（转换出错）</span>
              {/if}
            </div>

            {#if expandedId === ev.id}
              <div class="row-body">
                {#if ev.kind === 'convert' && ev.target}
                  <div class="res" class:t2s={ev.direction === 't2s'}>
                    <span class="res-arrow" aria-hidden="true">→</span>
                    <span class="res-text">{ev.target}</span>
                    <button
                      class="copy"
                      on:click={(e) => copyText(e, ev.target)}
                      aria-label="复制转换结果"
                    >
                      复制
                    </button>
                  </div>
                {:else if ev.kind === 'skip'}
                  <div class="detail">原因：{skipReasonText(ev.skip_reason)}</div>
                {:else if ev.kind === 'error'}
                  <div class="detail errd">{ev.error || '未知错误'}</div>
                {/if}
              </div>
            {/if}
          </button>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Counter footer -->
  <div class="foot">
    <span class="muted">累计转换</span>
    <b>{$status.counter ?? 0}</b>
    <button class="reset" on:click={reset}>重置</button>
  </div>

</div>

<style>
  .status {
    display: flex; flex-direction: column;
    flex: 1; min-height: 0;
  }

  /* ---------- floating glass card ---------- */
  .float-card {
    margin: -30px 16px 0;
    background: rgba(255, 255, 255, 0.88);
    backdrop-filter: blur(14px);
    -webkit-backdrop-filter: blur(14px);
    border: 1px solid rgba(255,255,255,0.75);
    border-radius: var(--radius);
    box-shadow: 0 10px 28px rgba(22, 24, 29, 0.14);
    padding: 11px 13px 13px;
    position: relative; z-index: 2;
    flex-shrink: 0;
  }
  .panes {
    display: grid; grid-template-columns: 1fr auto 1fr;
    gap: 10px; align-items: stretch;
  }
  .pane {
    border-radius: var(--radius-sm);
    padding: 8px 10px;
    min-height: 72px; max-height: 108px; overflow-y: auto;
    font-family: var(--font-mono);
    font-size: 12.5px; line-height: 1.6;
    word-break: break-all;
    min-width: 0;
  }
  .pane .tag {
    font-family: inherit;
    display: inline-block; font-size: 10px; font-weight: 700;
    letter-spacing: 0.06em; margin-bottom: 4px;
    font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', sans-serif;
  }
  .pane.jian { background: var(--blue-soft); border: 1px solid rgba(79,142,247,.3); }
  .pane.jian .tag { color: var(--blue); }
  .pane.fan { background: var(--green-soft); border: 1px solid rgba(52,208,166,.35); }
  .pane.fan .tag { color: var(--green); }
  .pane.empty .pane-body {
    color: var(--text-hint); font-style: italic;
    font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', sans-serif;
    font-size: 12px;
  }

  .fc-title {
    font-size: 11px; font-weight: 600; letter-spacing: 0.07em;
    color: var(--text-secondary); text-transform: uppercase;
    margin-bottom: 9px;
  }

  /* static flow indicator between the panes */
  .flow {
    align-self: center;
    color: var(--text-hint);
    font-size: 15px; font-weight: 600;
    flex-shrink: 0;
  }

  /* ---------- history ---------- */
  .hist {
    flex: 1; min-height: 0;
    display: flex; flex-direction: column;
    margin: 13px 16px 0;
  }
  .hist-head {
    display: flex; align-items: center; justify-content: space-between;
    font-size: 11px; font-weight: 600; letter-spacing: 0.07em;
    color: var(--text-secondary); text-transform: uppercase;
    margin-bottom: 7px;
  }
  .hist-head .clear {
    border: none; background: none; cursor: pointer;
    font-size: 11px; font-weight: 500;
    color: var(--text-hint);
    letter-spacing: 0; text-transform: none;
    padding: 2px 6px; border-radius: var(--radius-sm);
  }
  .hist-head .clear:hover { color: var(--text-primary); background: var(--surface); }

  .hist-empty {
    color: var(--text-hint); font-size: 12.5px; font-style: italic;
    padding: 4px 0;
  }

  .hist-list {
    flex: 1; min-height: 0; overflow-y: auto;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
  }
  .row {
    display: flex; flex-direction: column; gap: 7px;
    width: 100%;
    background: transparent; border: none;
    border-bottom: 1px solid var(--border);
    padding: 9px 13px;
    cursor: pointer; text-align: left;
    font-family: inherit; color: var(--text-primary);
    transition: background 100ms;
  }
  .row:last-child { border-bottom: none; }
  .row:hover, .row.open { background: var(--surface); }

  .row-head {
    display: flex; align-items: center; gap: 9px;
    font-size: 12.5px; min-width: 0;
  }
  .t {
    font-family: var(--font-mono); font-size: 11px;
    color: var(--text-hint); flex-shrink: 0;
  }
  .dir {
    font-size: 10px; font-weight: 700; flex-shrink: 0;
    padding: 1px 6px; border-radius: 4px; letter-spacing: 0.02em;
    background: var(--green-soft); color: var(--green);
  }
  .dir.t2s { background: var(--blue-soft); color: var(--blue); }
  .dir.neutral { background: var(--surface); color: var(--text-secondary); border: 1px solid var(--border); }
  .dir.errb { background: #FEE2E2; color: #DC2626; }
  .txt {
    flex: 1; min-width: 0;
    overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  }
  .txt.muted { color: var(--text-hint); font-style: italic; }

  .row-body { padding-top: 0; }
  .res {
    display: flex; align-items: center; gap: 8px;
    background: var(--bg);
    border: 1px solid rgba(52,208,166,.45);
    border-left: 3px solid var(--green-bright);
    border-radius: var(--radius-sm);
    padding: 8px 10px;
    font-family: var(--font-mono); font-size: 12px;
    line-height: 1.5;
  }
  .res.t2s {
    border-color: rgba(79,142,247,.4);
    border-left-color: var(--blue-bright);
  }
  .res-arrow { color: var(--green); font-weight: 700; flex-shrink: 0; }
  .res.t2s .res-arrow { color: var(--blue); }
  .res-text { flex: 1; min-width: 0; word-break: break-all; }
  .copy {
    flex-shrink: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'PingFang SC', sans-serif;
    font-size: 11px;
    border: 1px solid var(--border-strong); border-radius: 6px;
    background: var(--bg); padding: 3px 10px; cursor: pointer;
    color: var(--text-secondary);
    transition: all 120ms;
  }
  .copy:hover { color: var(--green); border-color: var(--green-bright); }
  .res.t2s .copy:hover { color: var(--blue); border-color: var(--blue-bright); }

  .detail {
    font-size: 12px; color: var(--text-secondary);
    padding: 6px 10px;
    background: var(--bg);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
  }
  .detail.errd {
    color: #DC2626; border-color: #FCA5A5; background: #FEF2F2;
    font-family: var(--font-mono);
  }

  /* ---------- footer ---------- */
  .foot {
    display: flex; align-items: center; justify-content: center; gap: 8px;
    padding: 9px 16px 11px;
    font-size: 11.5px; color: var(--text-secondary);
    flex-shrink: 0;
  }
  .foot b {
    font-family: var(--font-mono); font-size: 13px; font-weight: 800;
    font-variant-numeric: tabular-nums;
    background: var(--grad);
    -webkit-background-clip: text; background-clip: text;
    color: transparent;
  }
  .status.rev .foot b { background: var(--grad-rev); -webkit-background-clip: text; background-clip: text; }
  .foot .reset {
    border: none; background: none;
    font-size: 11px; font-weight: 500;
    color: var(--text-hint); cursor: pointer;
    padding: 2px 6px; border-radius: var(--radius-sm);
  }
  .foot .reset:hover { color: var(--text-primary); background: var(--surface); }
</style>
