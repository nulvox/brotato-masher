(() => {
  const state = { raw: null, original: null, summary: null };
  const $ = (id) => document.getElementById(id);
  const labels = { characters_unlocked: 'Characters', consumables_unlocked: 'Consumables', items_unlocked: 'Items', upgrades_unlocked: 'Upgrades', weapons_unlocked: 'Weapons', zones_unlocked: 'Zones', challenges_completed: 'Completed challenges' };
  const statLabels = { chal_hourglass_quit_wave: 'Hourglass quits', enemies_killed: 'Enemies killed', enemies_killed_far_away: 'Enemies killed far away', evil_mob_killed: 'Evil mobs killed', evil_mob_killed_by: 'Evil mobs killed by', fruit_eaten_full_hp: 'Fruit eaten at full HP', is_unlock_all_save: 'Unlock-all flag', materials_collected: 'Materials collected', run_started: 'Runs started', run_won: 'Runs won', steps_taken: 'Steps taken', trees_killed: 'Trees killed' };
  const counterLabels = { items_bought: 'Item purchase counters', killed_enemies: 'Killed-enemy counters', killed_by_enemies: 'Killed-by-enemy counters' };
  const status = (message, error = false) => { $('status').textContent = message; $('status').classList.toggle('error', error); };
  const call = (name, ...args) => JSON.parse(window.brotatoWasm[name](...args));
  let ready;
  async function startWasm() { const go = new Go(); const result = await WebAssembly.instantiateStreaming(fetch('brotato.wasm'), go.importObject); go.run(result.instance); }
  ready = startWasm().catch(() => status('The WASM editor could not load. Run the build script and serve the web directory.', true));

  function render() {
    const doc = JSON.parse(state.raw), counts = state.summary.value.counts;
    $('summary').textContent = `Save version ${state.summary.value.version} · ${Object.values(counts).reduce((a, b) => a + b, 0)} unlock/challenge entries · ${state.summary.value.difficulties} difficulty records`;
    $('collections').innerHTML = Object.keys(labels).map(key => {
      const ids = doc[key];
      return `<div class="collection"><h3>${labels[key]}</h3><p>${ids.length} loaded IDs — uncheck entries to remove them</p><button type="button" data-action="all" data-key="${key}">Select all</button> <button type="button" class="secondary" data-action="none" data-key="${key}">Clear</button><div class="ids">${ids.map((id, i) => `<label><input type="checkbox" data-collection="${key}" data-index="${i}" checked> ${id}</label>`).join('')}</div></div>`;
    }).join('');
    $('stats').innerHTML = Object.entries(state.summary.value.stats).map(([key, value]) => typeof value === 'boolean'
      ? `<label class="stat"><span>${statLabels[key] || key}</span><input type="checkbox" data-stat="${key}" ${value ? 'checked' : ''}></label>`
      : `<label class="stat"><span>${statLabels[key] || key}</span><input type="number" min="0" step="1" value="${value}" data-stat="${key}"></label>`).join('');
    $('counters').innerHTML = Object.keys(counterLabels).map(key => `<label class="wide-field"><span>${counterLabels[key]} (${Object.keys(doc[key]).length})</span><textarea rows="4" data-counter="${key}">${JSON.stringify(doc[key], null, 2)}</textarea></label>`).join('');
    $('announcements').value = JSON.stringify(doc.read_announcements, null, 2);
    $('difficulties').innerHTML = doc.difficulties_unlocked.flatMap((character, ci) => character.zones_difficulty_info.map((zone, zi) => `<tr><th scope="row">${character.character_id}</th><td>${zone.zone_id}</td><td><input type="number" min="0" step="1" data-difficulty="${ci}:${zi}:selected" value="${zone.difficulty_selected_value}"></td><td><input type="number" min="0" step="1" data-difficulty="${ci}:${zi}:max" value="${zone.max_selectable_difficulty}"></td><td>${zone.max_difficulty_beaten.wave_number}</td><td>${zone.max_endless_wave_beaten.wave_number}</td></tr>`)).join('');
    $('editor').classList.remove('hidden');
  }

  async function load(file) {
    await ready;
    const raw = await file.text(), check = call('validate', raw);
    if (!check.ok) { state.raw = null; $('editor').classList.add('hidden'); status(check.value, true); return; }
    state.raw = raw; state.original = raw; state.summary = call('summary', raw); render(); status(`Loaded ${file.name}. Changes stay in memory until you download.`);
  }

  function edits() {
    const doc = JSON.parse(state.raw), output = {};
    for (const key of Object.keys(labels)) output[key] = [...document.querySelectorAll(`[data-collection="${key}"]`)].filter(x => x.checked).map(x => Number(x.closest('label').textContent.trim()));
    output.stats = {};
    document.querySelectorAll('[data-stat]').forEach(input => output.stats[input.dataset.stat] = input.type === 'checkbox' ? input.checked : Number(input.value));
    for (const key of Object.keys(counterLabels)) output[key] = JSON.parse(document.querySelector(`[data-counter="${key}"]`).value);
    output.read_announcements = JSON.parse($('announcements').value);
    output.difficulties_unlocked = doc.difficulties_unlocked;
    document.querySelectorAll('[data-difficulty]').forEach(input => { const [ci, zi, field] = input.dataset.difficulty.split(':'); output.difficulties_unlocked[ci].zones_difficulty_info[zi][field === 'selected' ? 'difficulty_selected_value' : 'max_selectable_difficulty'] = Number(input.value); });
    return output;
  }

  $('save-file').addEventListener('change', e => e.target.files[0] && load(e.target.files[0]));
  $('drop-zone').addEventListener('dragover', e => e.preventDefault());
  $('drop-zone').addEventListener('drop', e => { e.preventDefault(); e.dataTransfer.files[0] && load(e.dataTransfer.files[0]); });
  $('collections').addEventListener('click', e => { const button = e.target.closest('[data-action]'); if (!button) return; document.querySelectorAll(`[data-collection="${button.dataset.key}"]`).forEach(x => x.checked = button.dataset.action === 'all'); });
  $('reset').addEventListener('click', () => { state.raw = state.original; state.summary = call('summary', state.raw); render(); status('Edits reset.'); });
  $('download').addEventListener('click', () => { try { const result = call('apply', state.raw, JSON.stringify(edits())); if (!result.ok) return status(result.value, true); const blob = new Blob([result.value], { type: 'application/json' }); const link = document.createElement('a'); link.href = URL.createObjectURL(blob); link.download = 'save_v3_0_edited.json'; link.click(); URL.revokeObjectURL(link.href); status('Downloaded a new save. The source file was not changed.'); } catch (error) { status(`Could not create save: ${error.message}`, true); } });
})();
