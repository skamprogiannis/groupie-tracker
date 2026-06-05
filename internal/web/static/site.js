document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('searchForm');
    const input = document.getElementById('q');
    const catalog = document.querySelector('.catalog');

    function escapeHtml(s) {
        return String(s).replace(/[&<>"']/g, function (c) {
            return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": "&#39;" }[c];
        });
    }

    function renderResults(results) {
        if (!catalog) return;
        // remove any existing empty-state message
        const existing = catalog.parentNode.querySelector('.empty-state');
        if (existing) existing.remove();

        if (!results || results.length === 0) {
            catalog.innerHTML = '';
            const p = document.createElement('p');
            p.className = 'empty-state';
            p.textContent = 'No results found.';
            catalog.parentNode.appendChild(p);
            return;
        }

        catalog.innerHTML = results.map(r => {
            return `<li class="catalog-item"><a class="catalog-link" href="/artist/${r.id}"><img src="${escapeHtml(r.image)}" alt="${escapeHtml(r.name)}"><div class="catalog-card"><h3>${escapeHtml(r.name)}</h3><p class="catalog-card__info">Created ${r.creationDate} · First album: ${escapeHtml(r.firstAlbum)}</p></div></a></li>`;
        }).join('');
    }

    // Format location strings: replace underscores/hyphens with spaces and Title Case each word
    function formatLocation(raw) {
        if (!raw) return raw;
        const cleaned = raw.replace(/[_-]+/g, ' ');
        return cleaned.split(' ').map(function (w) {
            if (!w) return w;
            return w.charAt(0).toUpperCase() + w.slice(1).toLowerCase();
        }).join(' ');
    }

    function formatExistingLocationsAndRelations() {
        // Locations lists
        document.querySelectorAll('section').forEach(function (sec) {
            const h2 = sec.querySelector('h2');
            if (!h2) return;
            const title = h2.textContent.trim();
            if (title === 'Locations') {
                sec.querySelectorAll('ul li').forEach(function (li) {
                    li.textContent = formatLocation(li.textContent.trim());
                });
            }
            if (title === 'Relations') {
                sec.querySelectorAll('dt').forEach(function (dt) {
                    dt.textContent = formatLocation(dt.textContent.trim());
                });
            }
        });
    }

    // Autocomplete UI
    function createAutocomplete(formEl, inputEl) {
        const list = document.createElement('div');
        list.className = 'autocomplete-list';
        list.style.display = 'none';
        formEl.appendChild(list);

        function show(items) {
            if (!items || items.length === 0) { list.style.display = 'none'; list.innerHTML = ''; return; }
            list.innerHTML = items.map(function (it) {
                return `<div class="autocomplete-item" data-id="${it.id}">${escapeHtml(it.name)} <small style="color:#64748b; margin-left:8px;">${it.creationDate}</small></div>`;
            }).join('');
            list.style.display = 'block';
        }

        list.addEventListener('click', function (ev) {
            const item = ev.target.closest('.autocomplete-item');
            if (!item) return;
            const id = item.getAttribute('data-id');
            if (id) {
                window.location.href = '/artist/' + id;
            }
        });

        document.addEventListener('click', function (ev) {
            if (!formEl.contains(ev.target)) {
                list.style.display = 'none';
            }
        });

        return { show };
    }

    if (!form || !input) {
        // still format locations if on artist page
        formatExistingLocationsAndRelations();
        return;
    }

    const autocomplete = createAutocomplete(form, input);

    let debounceTimer = null;
    input.addEventListener('input', function () {
        const q = input.value || '';
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(function () {
            fetch('/api/search?q=' + encodeURIComponent(q))
                .then(function (res) { return res.json(); })
                .then(function (data) {
                    autocomplete.show((data.results || []).slice(0, 8));
                })
                .catch(function () { autocomplete.show([]); });
        }, 150);
    });

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

    // Format any locations/relations already present on the page (artist detail page)
    formatExistingLocationsAndRelations();
});
