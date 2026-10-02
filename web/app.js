(() => {
  const state = { raw: null, original: null, summary: null };
  const $ = (id) => document.getElementById(id);
  const status = (message, error = false) => { $('status').textContent = message; $('status').classList.toggle('error', error); };
  const call = (name, ...args) => JSON.parse(window.brotatoWasm[name](...args));
  const labels = { characters_unlocked: 'Characters', consumables_unlocked: 'Consumables', items_unlocked: 'Items', upgrades_unlocked: 'Upgrades', weapons_unlocked: 'Weapons', zones_unlocked: 'Zones' };
  const statLabels = { chal_hourglass_quit_wave: 'Hourglass quits', enemies_killed: 'Enemies killed', enemies_killed_far_away: 'Enemies killed far away', evil_mob_killed: 'Evil mobs killed', evil_mob_killed_by: 'Evil mobs killed by', fruit_eaten_full_hp: 'Fruit eaten at full HP', is_unlock_all_save: 'Unlock-all flag', materials_collected: 'Materials collected', run_started: 'Runs started', run_won: 'Runs won', steps_taken: 'Steps taken', trees_killed: 'Trees killed' };
  let ready;
  async function startWasm() { const go = new Go(); const result = await WebAssembly.instantiateStreaming(fetch('brotato.wasm'), go.importObject); go.run(result.instance); }
  ready = startWasm().catch(() => status('The WASM editor could not load. Run the build script and serve the web directory.', true));
  function render() {
    const counts = state.summary.value.counts;
    $('summary').textContent = `Save version ${state.summary.value.version} · ${Object.values(counts).reduce((a,b)=>a+b,0)} unlock entries loaded`;
    $('collections').innerHTML = Object.entries(counts).map(([key]) => { const ids = JSON.parse(state.raw)[key]; return `<div class="collection"><h3>${labels[key]}</h3><p>${ids.length} loaded IDs — uncheck entries to remove them</p><button type="button" data-action="all" data-key="${key}">Select all</button> <button type="button" class="secondary" data-action="none" data-key="${key}">Clear</button><div class="ids">${ids.map((id,i)=>`<label><input type="checkbox" data-collection="${key}" data-index="${i}" checked> ${id}</label>`).join('')}</div></div>`; }).join('');
    $('stats').innerHTML = Object.entries(state.summary.value.stats).map(([key,value]) => `<label class="stat"><span>${statLabels[key] || key}</span><input type="number" min="0" step="1" value="${value}" data-stat="${key}"></label>`).join('');
    $('editor').classList.remove('hidden');
  }
  async function load(file) { await ready; const raw = await file.text(); const check = call('validate', raw); if (!check.ok) { state.raw = null; $('editor').classList.add('hidden'); status(check.value, true); return; } state.raw = raw; state.original = raw; state.summary = call('summary', raw); render(); status(`Loaded ${file.name}. Changes stay in memory until you download.`); }
  function edits() { const doc = JSON.parse(state.raw); const output = {}; for (const key of Object.keys(labels)) output[key] = [...document.querySelectorAll(`[data-collection="${key}"]`)].filter(x => x.checked).map(x => Number(x.closest('label').textContent.trim())); output.stats = {}; document.querySelectorAll('[data-stat]').forEach(input => output.stats[input.dataset.stat] = Number(input.value)); return output; }
  $('save-file').addEventListener('change', e => e.target.files[0] && load(e.target.files[0]));
  $('drop-zone').addEventListener('dragover', e => { e.preventDefault(); }); $('drop-zone').addEventListener('drop', e => { e.preventDefault(); e.dataTransfer.files[0] && load(e.dataTransfer.files[0]); });
  $('collections').addEventListener('click', e => { const button = e.target.closest('[data-action]'); if (!button) return; document.querySelectorAll(`[data-collection="${button.dataset.key}"]`).forEach(x => x.checked = button.dataset.action === 'all'); });
  $('reset').addEventListener('click', () => { state.raw = state.original; state.summary = call('summary', state.raw); render(); status('Edits reset.'); });
  $('download').addEventListener('click', () => { const result = call('apply', state.raw, JSON.stringify(edits())); if (!result.ok) return status(result.value, true); const blob = new Blob([result.value], {type:'application/json'}); const link = document.createElement('a'); link.href = URL.createObjectURL(blob); link.download = 'save_v3_0_edited.json'; link.click(); URL.revokeObjectURL(link.href); status('Downloaded a new save. The source file was not changed.'); });
})();
