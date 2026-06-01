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
        :root { --green: #2d7a4e; --green-hover: #3a8f5e; --red: #9f3a38; --red-hover: #b84644; --border:#1f2e23; }
        body { margin:0; font-family:system-ui,-apple-system,sans-serif; background:#0a0f0a; color:#e6e6e6; }
        .container { max-width:1280px; margin:0 auto; padding:20px; }
        h1 { color:var(--green); margin-bottom:20px; }
        input#q {
            width:100%; padding:16px; font-size:1.1rem; background:#1a2421;
            border:2px solid var(--green); border-radius:8px; color:white;
            margin-bottom:15px; box-sizing:border-box; outline:none;
        }
        input#q:focus { box-shadow:0 0 0 3px rgba(45,122,78,.25); }
        .controls { margin-bottom:20px; display:flex; gap:10px; flex-wrap:wrap; align-items:center; }
        .controls button {
            background:#1f2a24; color:#e6e6e6; border:1px solid var(--green);
            border-radius:6px; padding:8px 14px; cursor:pointer; font-size:0.9em;
            display:inline-flex; align-items:center; gap:6px;
            transition:background .15s, border-color .15s, color .15s;
        }
        .controls button:hover { background:#2d3a30; }
        .controls button.active { background:rgba(45,122,78,.35); border-color:#5fb970; color:#bfffd3; }
        .controls button.active:hover { background:rgba(45,122,78,.5); }

        /* Results */
        .result-group { background:#141a2a; border:1px solid var(--green); border-radius:8px; margin-bottom:20px; overflow:hidden; }
        .profile-header {
            padding:14px 18px; background:#1f2a24; display:flex;
            justify-content:space-between; align-items:center;
            cursor:pointer; font-weight:bold; user-select:none;
        }
        .profile-header .arrow { transition:transform 0.2s ease; display:inline-block; }
        .profile-header.open .arrow { transform:rotate(180deg); }
        .channels { padding:12px; max-height:600px; overflow-y:auto; }
        .channel {
            display:grid; grid-template-columns:1fr auto auto;
            gap:16px; padding:12px; border-bottom:1px solid #2a3a2f; align-items:center;
        }
        .channel:last-child { border-bottom:none; }
        .ch-toggle {
            padding:6px 14px; border-radius:9999px; font-size:0.9em; font-weight:500;
            border:none; cursor:pointer; transition:filter .15s, transform .1s; white-space:nowrap;
        }
        .ch-toggle:active { transform:scale(.96); }
        .ch-toggle.enabled        { background:var(--green); color:white; }
        .ch-toggle.enabled:hover  { filter:brightness(1.15); }
        .ch-toggle.disabled       { background:var(--red);   color:white; }
        .ch-toggle.disabled:hover { filter:brightness(1.15); }
        .genre-btn {
            padding:6px 12px; border:none; border-radius:6px; cursor:pointer;
            font-size:0.85em; font-weight:500; white-space:nowrap; transition:filter .15s;
        }
        .genre-btn.will-disable       { background:#3b5a7a; color:white; }
        .genre-btn.will-disable:hover { filter:brightness(1.2); }
        .genre-btn.will-enable        { background:#4a6a3a; color:white; }
        .genre-btn.will-enable:hover  { filter:brightness(1.2); }
        .all-hidden-notice {
            padding:12px 16px; color:#9aaa9a; font-size:0.9em; font-style:italic; display:none;
        }

        /* Export modal */
        .modal-overlay {
            display:none; position:fixed; inset:0;
            background:rgba(0,0,0,.72); z-index:100;
            align-items:center; justify-content:center; padding:20px;
        }
        .modal-overlay.open { display:flex; }
        .modal-box {
            background:linear-gradient(180deg,rgba(17,24,21,.98),rgba(13,20,16,.97));
            border:1px solid var(--border); border-radius:16px;
            padding:24px; width:100%; max-width:640px;
            box-shadow:0 24px 64px rgba(0,0,0,.6);
            display:flex; flex-direction:column; gap:14px;
            max-height:90vh;
        }
        .modal-header {
            display:flex; align-items:center; justify-content:space-between; gap:12px;
        }
        .modal-title { font-size:16px; font-weight:700; color:#e6f2e6; }
        .modal-sub   { font-size:13px; color:#9aaa9a; margin-top:2px; }
        .modal-close {
            background:none; border:none; color:#9aaa9a; font-size:20px;
            cursor:pointer; padding:4px 8px; border-radius:6px; line-height:1;
            transition:color .15s, background .15s;
        }
        .modal-close:hover { color:#e6e6e6; background:rgba(255,255,255,.08); }
        .modal-textarea {
            width:100%; flex:1 1 auto; min-height:260px; max-height:50vh;
            background:#0d1410; border:1px solid var(--border); border-radius:10px;
            color:#c8dcc8; font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,
                           Consolas,"Liberation Mono","Courier New",monospace;
            font-size:13px; line-height:1.55; padding:12px; resize:vertical;
            outline:none; box-sizing:border-box;
        }
        .modal-textarea:focus { border-color:var(--green); box-shadow:0 0 0 3px rgba(45,122,78,.2); }
        .modal-actions { display:flex; gap:10px; align-items:center; flex-wrap:wrap; }
        .modal-btn {
            background:#1f2a24; color:#e6e6e6; border:1px solid var(--border);
            border-radius:8px; padding:10px 16px; cursor:pointer; font-size:0.9em;
            display:inline-flex; align-items:center; gap:7px;
            transition:background .15s, border-color .15s;
        }
        .modal-btn:hover { background:#2d3a30; border-color:var(--green); }
        .modal-btn.primary { background:var(--green); border-color:var(--green); color:white; }
        .modal-btn.primary:hover { background:var(--green-hover); }
        .modal-count { font-size:12px; color:#9aaa9a; margin-left:auto; }
    </style>
</head>
<body>

<!-- Export modal -->
<div class="modal-overlay" id="exportModal" role="dialog" aria-modal="true" aria-labelledby="exportModalTitle">
    <div class="modal-box">
        <div class="modal-header">
            <div>
                <div class="modal-title" id="exportModalTitle">
                    <i class="fa-solid fa-clipboard"></i> Enabled Channel Names
                </div>
                <div class="modal-sub">Select All then Ctrl+C / Cmd+C to copy</div>
            </div>
            <button class="modal-close" onclick="closeExportModal()" aria-label="Close">&times;</button>
        </div>
        <textarea class="modal-textarea" id="exportTextarea" readonly
            spellcheck="false" autocomplete="off"></textarea>
        <div class="modal-actions">
            <button class="modal-btn primary" onclick="selectAllExport()">
                <i class="fa-solid fa-arrow-pointer"></i> Select All
            </button>
            <button class="modal-btn" onclick="closeExportModal()">
                <i class="fa-solid fa-xmark"></i> Close
            </button>
            <span class="modal-count" id="exportCount"></span>
        </div>
    </div>
</div>

<div class="container">
    <h1><i class="fa-solid fa-magnifying-glass"></i> Cross-Profile Channel Search</h1>
    <input type="text" id="q"
        placeholder="Search channel titles (regex supported, e.g. espn|cincinnati|fox)"
        autofocus>

    <div class="controls">
        <button onclick="expandAll()">
            <i class="fa-solid fa-angles-down"></i> Expand All
        </button>
        <button onclick="collapseAll()">
            <i class="fa-solid fa-angles-up"></i> Collapse All
        </button>
        <button id="hideDisabledBtn" onclick="toggleHideDisabled()">
            <i class="fa-solid fa-eye-slash"></i> Hide Disabled
        </button>
        <button id="exportBtn" onclick="exportEnabled()">
            <i class="fa-solid fa-clipboard"></i> Copy Enabled Names
        </button>
    </div>

    <div id="results">Enter a search term above...</div>
</div>

<script>
// ─── Persistence ──────────────────────────────────────────────────────────

const LS_QUERY_KEY         = 'stalkerhek_search_query';
const LS_EXPAND_KEY        = 'stalkerhek_search_expanded';
const LS_HIDE_DISABLED_KEY = 'stalkerhek_search_hide_disabled';

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
function loadQuery()  { try { return localStorage.getItem(LS_QUERY_KEY) || ''; } catch(e) { return ''; } }
function saveQuery(q) { try { localStorage.setItem(LS_QUERY_KEY, q); } catch(e) {} }
function loadHideDisabled()     { try { return localStorage.getItem(LS_HIDE_DISABLED_KEY) === '1'; } catch(e) { return false; } }
function saveHideDisabled(val)  { try { localStorage.setItem(LS_HIDE_DISABLED_KEY, val ? '1' : '0'); } catch(e) {} }

// ─── State ────────────────────────────────────────────────────────────────

let debounceTimer;
const qInput     = document.getElementById('q');
const resultsDiv = document.getElementById('results');
let expandedIds  = loadExpandedSet();
let hideDisabled = loadHideDisabled();

// ─── Escape helpers ───────────────────────────────────────────────────────

function escHtml(s) {
    return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
function escAttr(s) {
    return String(s||'').replace(/\\/g,'\\\\').replace(/'/g,"\\'");
}

// ─── M3U8 title sanitization (mirrors stalker/normalize.go) ──────────────
//
// Unicode ranges stripped — identical to Go StripSuperscripts():
//   U+00B2–U+00B3, U+00B9   legacy superscript digits ²³¹
//   U+02B0–U+02FF            Spacing Modifier Letters  (ᶠᵖˢ …)
//   U+1D00–U+1DBF            Phonetic Extensions       (ᴴᴰᴿᴬᵂ …)
//   U+2069                   Pop Directional Isolate
//   U+2070–U+209F            Superscripts & Subscripts block
//   U+2600–U+26FF            Miscellaneous Symbols     (☼ ★ …)
//   U+2C60–U+2C7F            Latin Extended-C          (ⱽ …)

function stripSuperscripts(s) {
    return s
        .replace(/[\u00B2\u00B3\u00B9\u02B0-\u02FF\u1D00-\u1DBF\u2069\u2070-\u209F\u2600-\u26FF\u2C60-\u2C7F]/g, '')
        .replace(/\s+/g, ' ')
        .trim();
}

function cleanTitleForM3U8(s) {
    s = stripSuperscripts(s);
    if (!s) return '';
    // Collapse orphaned punctuation runs surrounded by whitespace
    // mirrors: orphanMidPunctRE = \s[/\-_,;:.]+\s => single space
    s = s.replace(/\s[\/\-_,;:.]+\s/g, ' ');
    return s.trim();
}

// ─── Export modal ─────────────────────────────────────────────────────────

function exportEnabled() {
    const seen  = new Set();
    const names = [];

    // Walk ALL .channel rows across all groups — collapsed groups are hidden
    // visually but their rows remain in the DOM with buttons intact.
    document.querySelectorAll('.channel').forEach(function(row) {
        const toggle = row.querySelector('.ch-toggle');
        if (!toggle || !toggle.classList.contains('enabled')) return;
        const strong = row.querySelector('strong');
        if (!strong) return;
        const sanitized = cleanTitleForM3U8(strong.textContent || '');
        if (!sanitized || seen.has(sanitized)) return;
        seen.add(sanitized);
        names.push(sanitized);
    });

    const ta    = document.getElementById('exportTextarea');
    const count = document.getElementById('exportCount');

    if (names.length === 0) {
        ta.value       = '(no enabled channels found in current results)';
        count.textContent = '';
    } else {
        ta.value          = names.join('\n');
        count.textContent = names.length + ' channel' + (names.length === 1 ? '' : 's');
    }

    document.getElementById('exportModal').classList.add('open');

    // Auto-select so user can hit Ctrl+C immediately without an extra click
    setTimeout(function() { ta.select(); }, 60);
}

function selectAllExport() {
    const ta = document.getElementById('exportTextarea');
    ta.focus();
    ta.select();
}

function closeExportModal() {
    document.getElementById('exportModal').classList.remove('open');
}

// Close on overlay click (outside the box)
document.getElementById('exportModal').addEventListener('click', function(e) {
    if (e.target === this) closeExportModal();
});

// Close on Escape
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') closeExportModal();
});

// ─── Hide-Disabled filter ─────────────────────────────────────────────────

function applyHideDisabled() {
    const btn = document.getElementById('hideDisabledBtn');
    if (btn) btn.classList.toggle('active', hideDisabled);

    document.querySelectorAll('.result-group').forEach(function(group) {
        const rows  = group.querySelectorAll('.channel');
        let visible = 0;

        rows.forEach(function(row) {
            const toggle     = row.querySelector('.ch-toggle');
            const isDisabled = toggle && toggle.classList.contains('disabled');
            if (hideDisabled && isDisabled) {
                row.style.display = 'none';
            } else {
                row.style.display = '';
                visible++;
            }
        });

        let notice = group.querySelector('.all-hidden-notice');
        if (!notice) {
            notice = document.createElement('div');
            notice.className = 'all-hidden-notice';
            notice.textContent = 'All channels in this group are disabled and hidden.';
            const channelsDiv = group.querySelector('.channels');
            if (channelsDiv) channelsDiv.appendChild(notice);
        }
        notice.style.display = (hideDisabled && visible === 0) ? 'block' : 'none';
    });
}

function toggleHideDisabled() {
    hideDisabled = !hideDisabled;
    saveHideDisabled(hideDisabled);
    applyHideDisabled();
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
                        const enabledCls   = ch.enabled ? 'enabled'       : 'disabled';
                        const toggleLabel  = ch.enabled ? 'Enabled'        : 'Disabled';
                        const willDisable  = ch.enabled;
                        const genreBtnCls  = willDisable ? 'will-disable'  : 'will-enable';
                        const genreBtnLbl  = willDisable ? 'Disable Genre' : 'Enable Genre';
                        const disabledFlag = willDisable ? '1'             : '0';

                        return '<div class="channel">' +
                            '<div><strong>' + escHtml(ch.title) + '</strong>' +
                                '<br><small>' + escHtml(ch.genre) + '</small></div>' +
                            '<div>' +
                                '<button class="ch-toggle ' + enabledCls + '" ' +
                                    'onclick="toggleChannel(event,' + pid + ',\'' +
                                    escAttr(ch.cmd) + '\',' + ch.enabled + ',this)">' +
                                    toggleLabel +
                                '</button>' +
                            '</div>' +
                            '<div>' +
                                '<button class="genre-btn ' + genreBtnCls + '" ' +
                                    'data-genre-id="'   + escHtml(ch.genre_id) + '" ' +
                                    'data-genre-name="' + escHtml(ch.genre)    + '" ' +
                                    'onclick="toggleGenre(event,' + pid + ',\'' +
                                    escAttr(ch.genre_id) + '\',\'' +
                                    escAttr(ch.genre)    + '\',' + disabledFlag + ')">' +
                                    genreBtnLbl +
                                '</button>' +
                            '</div>' +
                        '</div>';
                    }).join('') +
                '</div>' +
            '</div>';
        });

        resultsDiv.innerHTML = html;
        applyHideDisabled();

    } catch(e) {
        resultsDiv.innerHTML = '<p style="color:red">Failed to load results</p>';
    }
}

// ─── Channel toggle ───────────────────────────────────────────────────────

async function toggleChannel(e, profileId, cmd, currentlyEnabled, btn) {
    e.stopPropagation();
    btn.disabled = true;

    const nowDisabled = currentlyEnabled ? '1' : '0';

    try {
        const fd = new URLSearchParams();
        fd.append('id',       profileId);
        fd.append('cmd',      cmd);
        fd.append('disabled', nowDisabled);

        const res = await fetch('/api/filters/toggle_channel', {
            method:  'POST',
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            body:    fd.toString()
        });
        if (!res.ok) throw new Error('HTTP ' + res.status);

        const nowEnabled = !currentlyEnabled;
        btn.className    = 'ch-toggle ' + (nowEnabled ? 'enabled' : 'disabled');
        btn.textContent  = nowEnabled ? 'Enabled' : 'Disabled';

        const row      = btn.closest('.channel');
        const genreBtn = row ? row.querySelector('.genre-btn') : null;
        if (genreBtn) {
            const willDisable    = nowEnabled;
            genreBtn.className   = 'genre-btn ' + (willDisable ? 'will-disable' : 'will-enable');
            genreBtn.textContent = willDisable ? 'Disable Genre' : 'Enable Genre';
            const gid   = genreBtn.getAttribute('data-genre-id')   || '';
            const gname = genreBtn.getAttribute('data-genre-name') || '';
            genreBtn.onclick = function(ev) {
                toggleGenre(ev, profileId, gid, gname, willDisable ? '1' : '0');
            };
        }

        btn.onclick = function(ev) { toggleChannel(ev, profileId, cmd, nowEnabled, btn); };
        applyHideDisabled();

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
    const confirmed   = confirm(
        (willDisable ? 'Disable' : 'Enable') + ' entire genre "' + genreName + '"?\n\n' +
        'This will ' + (willDisable ? 'disable' : 'enable') +
        ' ALL channels in that genre for this profile.'
    );
    if (!confirmed) return;

    try {
        const fd = new URLSearchParams();
        fd.append('id',       profileId);
        fd.append('genre_id', genreId);
        fd.append('disabled', willDisable ? '1' : '0');

        const res = await fetch('/api/filters/toggle_genre', {
            method:  'POST',
            headers: {'Content-Type': 'application/x-www-form-urlencoded'},
            body:    fd.toString()
        });
        if (!res.ok) throw new Error('HTTP ' + res.status);

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
    const btn = document.getElementById('hideDisabledBtn');
    if (btn) btn.classList.toggle('active', hideDisabled);

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
