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
        :root { --green: #2d7a4e; }
        body { margin:0; font-family:system-ui,-apple-system,sans-serif; background:#0a0f0a; color:#e6e6e6; }
        .container { max-width:1280px; margin:0 auto; padding:20px; }
        h1 { color:var(--green); margin-bottom:20px; }
        input#q { width:100%; padding:16px; font-size:1.1rem; background:#1a2421; border:2px solid var(--green); border-radius:8px; color:white; margin-bottom:15px; box-sizing:border-box; }
        .controls { margin-bottom:20px; display:flex; gap:10px; flex-wrap:wrap; }
        .result-group { background:#141a2a; border:1px solid var(--green); border-radius:8px; margin-bottom:20px; overflow:hidden; }
        .profile-header { padding:14px 18px; background:#1f2a24; display:flex; justify-content:space-between; align-items:center; cursor:pointer; font-weight:bold; user-select:none; }
        .profile-header .arrow { transition: transform 0.2s ease; display:inline-block; }
        .profile-header.open .arrow { transform: rotate(180deg); }
        .channels { padding:12px; max-height:600px; overflow-y:auto; }
        .channel { display:grid; grid-template-columns: 1fr auto auto; gap:16px; padding:12px; border-bottom:1px solid #2a3a2f; align-items:center; }
        .pill { padding:6px 14px; border-radius:9999px; font-size:0.9em; font-weight:500; }
        .ok { background:#2d7a4e; color:white; }
        .bad { background:#9f3a38; color:white; }
        button { padding:6px 12px; border:none; border-radius:6px; cursor:pointer; font-size:0.9em; }
        .genre-btn { background:#3b5a7a; color:white; }
        .controls button { background:#1f2a24; color:#e6e6e6; border:1px solid var(--green); border-radius:6px; padding:8px 14px; cursor:pointer; }
        .controls button:hover { background:#2d3a30; }
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
// ─── Persistence helpers ───────────────────────────────────────────────────

const LS_QUERY_KEY   = 'stalkerhek_search_query';
const LS_EXPAND_KEY  = 'stalkerhek_search_expanded'; // stored as JSON array of profile IDs

/** Return the Set of profile IDs the user last had expanded. */
function loadExpandedSet() {
    try {
        const raw = localStorage.getItem(LS_EXPAND_KEY);
        if (!raw) return new Set();
        const arr = JSON.parse(raw);
        return new Set(Array.isArray(arr) ? arr.map(Number) : []);
    } catch(e) { return new Set(); }
}

/** Persist the current expanded set. */
function saveExpandedSet(set) {
    try {
        localStorage.setItem(LS_EXPAND_KEY, JSON.stringify(Array.from(set)));
    } catch(e) {}
}

/** Read persisted query string (empty string if none). */
function loadQuery() {
    try { return localStorage.getItem(LS_QUERY_KEY) || ''; } catch(e) { return ''; }
}

/** Persist the current query string. */
function saveQuery(q) {
    try { localStorage.setItem(LS_QUERY_KEY, q); } catch(e) {}
}

// ─── State ────────────────────────────────────────────────────────────────

let debounceTimer;
const qInput     = document.getElementById('q');
const resultsDiv = document.getElementById('results');

// Expanded set lives for the lifetime of the page; persisted on every change.
let expandedIds = loadExpandedSet();

// ─── Search ───────────────────────────────────────────────────────────────

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
            resultsDiv.innerHTML = '<p style="color:#ff6666">Error: ' + data.error + '</p>';
            return;
        }

        if (!data.results || data.results.length === 0) {
            resultsDiv.innerHTML = '<p>No matching channels found for <strong>"' + escHtml(q) + '"</strong></p>';
            return;
        }

        let html = '<h2>Results</h2>';
        data.results.forEach(function(group) {
            const pid      = group.profile_id;
            const isOpen   = expandedIds.has(pid);
            const openCls  = isOpen ? ' open' : '';
            const display  = isOpen ? 'block' : 'none';

            html += '<div class="result-group" data-profile-id="' + pid + '">' +
                '<div class="profile-header' + openCls + '" onclick="toggleGroup(this)">' +
                    '<span>' + escHtml(group.profile_name) + ' \u2014 ' + group.total + ' channels</span>' +
                    '<span class="arrow">\u25bc</span>' +
                '</div>' +
                '<div class="channels" style="display:' + display + '">' +
                    group.channels.map(function(ch) {
                        return '<div class="channel">' +
                            '<div><strong>' + escHtml(ch.title) + '</strong><br><small>' + escHtml(ch.genre) + '</small></div>' +
                            '<div><span class="pill ' + (ch.enabled ? 'ok' : 'bad') + '">' + (ch.enabled ? 'Enabled' : 'Disabled') + '</span></div>' +
                            '<div><button class="genre-btn" onclick="toggleGenre(event,' + pid + ',\'' + escAttr(ch.genre_id) + '\',\'' + escAttr(ch.genre) + '\')">Toggle Genre</button></div>' +
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

// ─── Group toggle ─────────────────────────────────────────────────────────

function toggleGroup(header) {
    const group   = header.closest('.result-group');
    const content = header.nextElementSibling;
    const pid     = group ? Number(group.getAttribute('data-profile-id')) : null;
    const opening = content.style.display === 'none';

    content.style.display = opening ? 'block' : 'none';
    header.classList.toggle('open', opening);

    if (pid !== null) {
        if (opening) {
            expandedIds.add(pid);
        } else {
            expandedIds.delete(pid);
        }
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

// ─── Genre toggle ─────────────────────────────────────────────────────────

async function toggleGenre(e, profileId, genreId, genreName) {
    e.stopImmediatePropagation();

    const confirmed = confirm(
        'Toggle entire genre "' + genreName + '"?\n\nThis will affect ALL channels in that genre for this profile.'
    );
    if (!confirmed) return;

    try {
        const form = new FormData();
        form.append('id', profileId);
        form.append('genre_id', genreId);
        const res = await fetch('/api/filters/toggle_genre', { method: 'POST', body: form });
        if (res.ok) {
            alert('Genre toggled successfully. Refreshing...');
            performSearch();
        } else {
            alert('Failed to toggle genre');
        }
    } catch(err) {
        alert('Error: ' + err.message);
    }
}

// ─── Escape helpers ───────────────────────────────────────────────────────

function escHtml(s) {
    return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

function escAttr(s) {
    return String(s || '').replace(/\\/g,'\\\\').replace(/'/g,"\\'");
}

// ─── Boot ─────────────────────────────────────────────────────────────────

// Restore persisted query and kick off a search if there's something saved.
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
