document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('searchForm');
    const input = document.getElementById('q');
    const catalog = document.querySelector('.catalog');

    function renderResults(results) {
        if (!catalog) return;
        if (!results || results.length === 0) {
            catalog.innerHTML = '';
            const p = document.createElement('p');
            p.className = 'empty-state';
            p.textContent = 'No results found.';
            catalog.parentNode.appendChild(p);
            return;
        }
        // remove any existing empty-state message
        const existing = catalog.parentNode.querySelector('.empty-state');
        if (existing) existing.remove();

        catalog.innerHTML = results.map(r => {
            return `<li class="catalog-item"><a class="catalog-link" href="/artist/${r.id}"><img src="${r.image}" alt="${r.name}"><div class="catalog-card"><h3>${escapeHtml(r.name)}</h3><p class="catalog-card__info">Created ${r.creationDate} · First album: ${escapeHtml(r.firstAlbum)}</p><p class="catalog-card__members">Members: </p></div></a></li>`;
        }).join('');
    }

    function escapeHtml(s) {
        return String(s).replace(/[&<>"']/g, function (c) {
            return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": "&#39;" }[c];
        });
    }

    if (!form || !input) return;

    form.addEventListener('submit', function (ev) {
        ev.preventDefault();
        const q = input.value || '';
        fetch('/api/search?q=' + encodeURIComponent(q))
            .then(function (res) { return res.json(); })
            .then(function (data) {
                renderResults(data.results);
            })
            .catch(function () {
                // noop: keep UI stable on error
            });
    });
});
