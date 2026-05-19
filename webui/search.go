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
        input#q { width:100%; padding:16px; font-size:1.1rem; background:#1a2421; border:2px solid var(--green); border-radius:8px; color:white; margin-bottom:15px; }
        .controls { margin-bottom:20px; display:flex; gap:10px; flex-wrap:wrap; }
        .result-group { background:#141a2a; border:1px solid var(--green); border-radius:8px; margin-bottom:20px; overflow:hidden; }
        .profile-header { padding:14px 18px; background:#1f2a24; display:flex; justify-content:space-between; align-items:center; cursor:pointer; font-weight:bold; }
        .channels { padding:12px; max-height:600px; overflow-y:auto; }
        .channel { display:grid; grid-template-columns: 1fr auto auto; gap:16px; padding:12px; border-bottom:1px solid #2a3a2f; align-items:center; }
        .pill { padding:6px 14px; border-radius:9999px; font-size:0.9em; font-weight:500; }
        .ok { background:#2d7a4e; color:white; } 
        .bad { background:#9f3a38; color:white; }
        button { padding:6px 12px; border:none; border-radius:6px; cursor:pointer; font-size:0.9em; }
        .genre-btn { background:#3b5a7a; color:white; }
    </style>
</head>
<body>
<div class="container">
    <h1><i class="fa-solid fa-magnifying-glass"></i> Cross-Profile Channel Search</h1>
    <input type="text" id="q" placeholder="Search channel titles (regex supported, e.g. espn|cincinnati|fox)" autofocus>
    
    <div class="controls">
        <button onclick="expandAll()">Expand All</button>
        <button onclick="collapseAll()">Collapse All</button>
    </div>

    <div id="results">Enter a search term above...</div>
</div>

<script>
let debounceTimer;
const qInput = document.getElementById('q');
const resultsDiv = document.getElementById('results');

async function performSearch() {
    const q = qInput.value.trim();
    if (q.length < 2) {
        resultsDiv.innerHTML = '<p style="color:#888">Type at least 2 characters to search...</p>';
        return;
    }

    try {
        const res = await fetch('/api/search?q=' + encodeURIComponent(q));
        const data = await res.json();

        if (data.error) {
            resultsDiv.innerHTML = '<p style="color:#ff6666">Error: ' + data.error + '</p>';
            return;
        }

        if (!data.results || data.results.length === 0) {
            resultsDiv.innerHTML = '<p>No matching channels found for <strong>"' + q + '"</strong></p>';
            return;
        }

        let html = '<h2>Results</h2>';
        data.results.forEach(function(group) {
            html += '<div class="result-group">' +
                '<div class="profile-header" onclick="toggleGroup(this)">' +
                    '<span>' + group.profile_name + ' — ' + group.total + ' channels</span>' +
                    '<span>▼</span>' +
                '</div>' +
                '<div class="channels" style="display:none">' +
                    group.channels.map(function(ch) {
                        return '<div class="channel">' +
                            '<div><strong>' + ch.title + '</strong><br><small>' + ch.genre + '</small></div>' +
                            '<div><span class="pill ' + (ch.enabled ? 'ok' : 'bad') + '">' + (ch.enabled ? 'Enabled' : 'Disabled') + '</span></div>' +
                            '<div><button class="genre-btn" onclick="toggleGenre(event, ' + group.profile_id + ', \'' + ch.genre_id + '\', \'' + ch.genre.replace(/'/g, "\\'") + '\')">Toggle Genre</button></div>' +
                        '</div>';
                    }).join('') +
                '</div></div>';
        });
        resultsDiv.innerHTML = html;
    } catch(e) {
        resultsDiv.innerHTML = '<p style="color:red">Failed to load results</p>';
    }
}

function toggleGroup(header) {
    const content = header.nextElementSibling;
    content.style.display = (content.style.display === 'none') ? 'block' : 'none';
}

function expandAll() {
    document.querySelectorAll('.channels').forEach(function(el) { el.style.display = 'block'; });
}

function collapseAll() {
    document.querySelectorAll('.channels').forEach(function(el) { el.style.display = 'none'; });
}

async function toggleGenre(e, profileId, genreId, genreName) {
    e.stopImmediatePropagation();
    
    const action = confirm('Toggle entire genre "' + genreName + '"?\n\nThis will affect ALL channels in that genre for this profile.') 
        ? 'toggle' : null;
    if (!action) return;

    try {
        const form = new FormData();
        form.append('id', profileId);
        form.append('genre_id', genreId);
        // We let the backend flip the current state (no need to send disabled flag)
        const res = await fetch('/api/filters/toggle_genre', { 
            method: 'POST', 
            body: form 
        });

        if (res.ok) {
            alert('Genre toggled successfully. Refreshing...');
            performSearch();   // auto refresh
        } else {
            alert('Failed to toggle genre');
        }
    } catch(err) {
        alert('Error: ' + err.message);
    }
}

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
