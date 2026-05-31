package webui

import (
	"html/template"
	"net/http"
)

func RegisterSearchPage(mux *http.ServeMux) {
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		tpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Search - Stalkerhek</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.6.0/css/all.min.css">
    <style>
        :root { --green: #2d7a4e; --green-hover: #3a8f5e; --red: #9f3a38; --red-hover: #b84644; }
        body { margin:0; font-family:system-ui,-apple-system,sans-serif; background:#0a0f0a; color:#e6e6e6; }
        .container { max-width:1280px; margin:0 auto; padding:20px; }
        h1 { color:var(--green); margin-bottom:20px; }
        input#q { width:100%; padding:16px; font-size:1.1rem; background:#1a2421; border:2px solid var(--green); border-radius:8px; color:white; margin-bottom:15px; box-sizing:border-box; outline:none; }
        input#q:focus { box-shadow: 0 0 0 3px rgba(45,122,78,.25); }
        .controls { margin-bottom:20px; display:flex; gap:10px; flex-wrap:wrap; }
        .controls button { background:#1f2a24; color:#e6e6e6; border:1px solid var(--green); border-radius:6px; padding:8px 14px; cursor:pointer; font-size:0.9em; }
        .controls button:hover { background:#2d3a30; }
        .result-group { background:#141a2a; border:1px solid var(--green); border-radius:8px; margin-bottom:20px; overflow:hidden; }
        .profile-header { padding:14px 18px; background:#1f2a24; display:flex; justify-content:space-between; align-items:center; cursor:pointer; font-weight:bold; user-select:none; }
        .profile-header .arrow { transition:transform 0.2s ease; display:inline-block; }
        .profile-header.open .arrow { transform:rotate(180deg); }
        .channels { padding:12px; max-height:600px; overflow-y:auto; }
        .channel { display:grid; grid-template-columns:1fr auto auto; gap:16px; padding:12px; border-bottom:1px solid #2a3a2f; align-items:center; }
        .channel:last-child { border-bottom:none; }

        /* channel toggle button — styled like a pill but clickable */
        .ch-toggle {
            padding:6px 14px; border-radius:9999px; font-size:0.9em; font-weight:500;
            border:none; cursor:pointer; transition:filter .15s, transform .1s;
            white-space:nowrap;
        }
        .ch-toggle:active { transform:scale(.96); }
        .ch-toggle.enabled  { background:var(--green); color:white; }
        .ch-toggle.enabled:hover  { filter:brightness(1.15); }
        .ch-toggle.disabled { background:var(--red);   color:white; }
        .ch-toggle.disabled:hover { filter:brightness(1.15); }

        /* genre toggle button */
        .genre-btn {
            padding:6px 12px; border:none; border-radius:6px; cursor:pointer;
            font-size:0.85em; font-weight:500; white-space:nowrap;
            transition:filter .15s;
        }
        .genre-btn.will-disable { background:#3b5a7a; color:white; }
        .genre-btn.will-disable:hover { filter:brightness(1.2); }
        .genre-btn.will-enable  { background:#4a6a3a; color:white; }
        .genre-btn.will-enable:hover  { filter:brightness(1.2); }
    </style>
</head>
<body>
<div class="container">
    <h1><i class="fa-solid fa-magnifying-glass"></i> Cross-Profile Channel Search</h1>
    <input type="text" id="q" placeholder="Search channel titles (regex supported, e.g. espn|cincinnati|fox)" autofocus>

    <div class="controls">
        <button onclick="expandAll()"><i class="fa-solid fa-angles-down"></i> Expand All</button>
        <button onclick="collapseAll()"><i class="fa-solid fa-angles-up"></i> Collapse All</button>
    </div>

    <div id="results">Enter a search term above...</div>
</div>

<script>
// ─── Persistence ──────────────────────────────────────────────────────────

const LS_QUERY_KEY  = 'stalkerhek_search_query';
const LS_EXPAND_KEY = 'stalkerhek_search_expanded';

function loadExpandedSet() {
    try {
        const raw = localStorage.getItem(LS_EXPAND_KEY);
        if (!raw) return new Set();
        const arr = JSON.parse(raw);
        return new Set(Array.isArray(arr) ? arr.map(Number) : []);
    } catch(e) { return new Set(); }
}

function saveExpandedSet(set) {
    try { localStorage.setItem(LS_EXPAND_KEY, JSON.stringify(Array.from(set))); } catch(e) {}
}

function loadQuery() {
    try { return localStorage.getItem(LS_QUERY_KEY) || ''; } catch(e) { return ''; }
}

function saveQuery(q) {
    try { localStorage.setItem(LS_QUERY_KEY, q); } catch(e) {}
}

// ─── State ────────────────────────────────────────────────────────────────

let debounceTimer;
const qInput     = document.getElementById('q');
const resultsDiv = document.getElementById('results');
let expandedIds  = loadExpandedSet();

// ─── Escape helpers ───────────────────────────────────────────────────────

function escHtml(s) {
    return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
function escAttr(s) {
    return String(s || '').replace(/\\/g,'\\\\').replace(/'/g,"\\'");
}

// ─── Search & render ──────────────────────────────────────────────────────

async function performSearch() {
    const q = qInput.value.trim();
    saveQuery(q);

    if (q.length < 2) {
        resultsDiv.innerHTML = '<p style="color:#888">Type at least 2 characters to search...</p>';
        return;
    }

    try {
        const res  = await fetch('/api/search?q=' + encodeURIComponent(q));
        const data = await res.json();

        if (data.error) {
            resultsDiv.innerHTML = '<p style="color:#ff6666">Error: ' + escHtml(data.error) + '</p>';
            return;
        }

        if (!data.results || data.results.length === 0) {
            resultsDiv.innerHTML = '<p>No matching channels found for <strong>"' + escHtml(q) + '"</strong></p>';
            return;
        }

        let html = '<h2>Results</h2>';
        data.results.forEach(function(group) {
            const pid     = group.profile_id;
            const isOpen  = expandedIds.has(pid);
            const openCls = isOpen ? ' open' : '';
            const display = isOpen ? 'block' : 'none';

            html += '<div class="result-group" data-profile-id="' + pid + '">' +
                '<div class="profile-header' + openCls + '" onclick="toggleGroup(this)">' +
                    '<span>' + escHtml(group.profile_name) + ' \u2014 ' + group.total + ' channels</span>' +
                    '<span class="arrow">\u25bc</span>' +
                '</div>' +
                '<div class="channels" style="display:' + display + '">' +
                    group.channels.map(function(ch) {
                        const enabledCls  = ch.enabled ? 'enabled'      : 'disabled';
                        const toggleLabel = ch.enabled ? 'Enabled'       : 'Disabled';
                        // genre button: if channel is enabled the genre is at least partially
                        // enabled, so the action is to disable it — and vice versa.
                        const genreWillDisable = ch.enabled;
                        const genreBtnCls      = genreWillDisable ? 'will-disable' : 'will-enable';
                        const genreBtnLabel    = genreWillDisable ? 'Disable Genre' : 'Enable Genre';
                        const disabledFlag     = genreWillDisable ? '1' : '0';

                        return '<div class="channel">' +
                            '<div><strong>' + escHtml(ch.title) + '</strong><br><small>' + escHtml(ch.genre) + '</small></div>' +
                            '<div>' +
                                '<button class="ch-toggle ' + enabledCls + '" ' +
                                    'onclick="toggleChannel(event,' + pid + ',\'' + escAttr(ch.cmd) + '\',' + ch.enabled + ',this)">' +
                                    toggleLabel +
                                '</button>' +
                            '</div>' +
                            '<div>' +
                                '<button class="genre-btn ' + genreBtnCls + '" ' +
                                    'onclick="toggleGenre(event,' + pid + ',\'' + escAttr(ch.genre_id) + '\',\'' + escAttr(ch.genre) + '\',' + disabledFlag + ')">' +
                                    genreBtnLabel +
                                '</button>' +
                            '</div>' +
                        '</div>';
                    }).join('') +
                '</div>' +
            '</div>';
        });

        resultsDiv.innerHTML = html;
    } catch(e) {
        resultsDiv.innerHTML = '<p style="color:red">Failed to load results</p>';
    }
}

// ─── Channel toggle ───────────────────────────────────────────────────────

async function toggleChannel(e, profileId, cmd, currentlyEnabled, btn) {
    e.stopPropagation();
    btn.disabled = true;

    const nowDisabled = currentlyEnabled ? '1' : '0'; // flip

    try {
        const fd = new URLSearchParams();
        fd.append('id',       profileId);
        fd.append('cmd',      cmd);
        fd.append('disabled', nowDisabled);

        const res = await fetch('/api/filters/toggle_channel', {
            method: 'POST',
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            body: fd.toString()
        });

        if (!res.ok) throw new Error('HTTP ' + res.status);

        // Flip button appearance in-place without a full re-search
        const nowEnabled = !currentlyEnabled;
        btn.className    = 'ch-toggle ' + (nowEnabled ? 'enabled' : 'disabled');
        btn.textContent  = nowEnabled ? 'Enabled' : 'Disabled';

        // Update the genre button on the same row to match new channel state
        const row       = btn.closest('.channel');
        const genreBtn  = row ? row.querySelector('.genre-btn') : null;
        if (genreBtn) {
            const willDisable      = nowEnabled;
            genreBtn.className     = 'genre-btn ' + (willDisable ? 'will-disable' : 'will-enable');
            genreBtn.textContent   = willDisable ? 'Disable Genre' : 'Enable Genre';
            // Update the onclick to pass the new disabled flag
            const gid   = genreBtn.getAttribute('data-genre-id')   || '';
            const gname = genreBtn.getAttribute('data-genre-name')  || '';
            genreBtn.onclick = function(ev) {
                toggleGenre(ev, profileId, gid, gname, willDisable ? '1' : '0');
            };
        }

        // Patch the onclick on the button itself with the new state
        btn.onclick = function(ev) {
            toggleChannel(ev, profileId, cmd, nowEnabled, btn);
        };

    } catch(err) {
        alert('Failed to toggle channel: ' + err.message);
    } finally {
        btn.disabled = false;
    }
}

// ─── Genre toggle ─────────────────────────────────────────────────────────

async function toggleGenre(e, profileId, genreId, genreName, disabledFlag) {
    e.stopPropagation();

    const willDisable = disabledFlag === '1' || disabledFlag === 1;
    const action      = willDisable ? 'disable' : 'enable';

    const confirmed = confirm(
        (willDisable ? 'Disable' : 'Enable') + ' entire genre "' + genreName + '"?\n\n' +
        'This will ' + action + ' ALL channels in that genre for this profile.'
    );
    if (!confirmed) return;

    try {
        const fd = new URLSearchParams();
        fd.append('id',       profileId);
        fd.append('genre_id', genreId);
        fd.append('disabled', willDisable ? '1' : '0');

        const res = await fetch('/api/filters/toggle_genre', {
            method: 'POST',
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            body: fd.toString()
        });

        if (!res.ok) throw new Error('HTTP ' + res.status);

        // Full re-search so all channel pills in this genre reflect new state
        performSearch();

    } catch(err) {
        alert('Failed to toggle genre: ' + err.message);
    }
}

// ─── Group expand / collapse ──────────────────────────────────────────────

function toggleGroup(header) {
    const group   = header.closest('.result-group');
    const content = header.nextElementSibling;
    const pid     = group ? Number(group.getAttribute('data-profile-id')) : null;
    const opening = content.style.display === 'none';

    content.style.display = opening ? 'block' : 'none';
    header.classList.toggle('open', opening);

    if (pid !== null) {
        opening ? expandedIds.add(pid) : expandedIds.delete(pid);
        saveExpandedSet(expandedIds);
    }
}

function expandAll() {
    document.querySelectorAll('.result-group').forEach(function(group) {
        const header  = group.querySelector('.profile-header');
        const content = group.querySelector('.channels');
        const pid     = Number(group.getAttribute('data-profile-id'));
        if (content) content.style.display = 'block';
        if (header)  header.classList.add('open');
        if (pid)     expandedIds.add(pid);
    });
    saveExpandedSet(expandedIds);
}

function collapseAll() {
    document.querySelectorAll('.result-group').forEach(function(group) {
        const header  = group.querySelector('.profile-header');
        const content = group.querySelector('.channels');
        const pid     = Number(group.getAttribute('data-profile-id'));
        if (content) content.style.display = 'none';
        if (header)  header.classList.remove('open');
        if (pid)     expandedIds.delete(pid);
    });
    saveExpandedSet(expandedIds);
}

// ─── Boot ─────────────────────────────────────────────────────────────────

(function init() {
    const saved = loadQuery();
    if (saved) {
        qInput.value = saved;
        performSearch();
    }
})();

qInput.addEventListener('input', function() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(performSearch, 350);
});
</script>
</body>
</html>`;

		t := template.Must(template.New("search").Parse(tpl))
		t.Execute(w, nil)
	})
}
