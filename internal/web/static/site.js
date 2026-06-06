document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('searchForm');
    const input = document.getElementById('q');
    const dropdown = document.getElementById('dropdown');
    const catalog = document.querySelector('.catalog');

    function escapeHtml(s) {
        return String(s).replace(/[&<>"']/g, function (c) {
            return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": "&#39;" }[c];
        });
    }

    function renderResults(results) {
        if (!catalog) return;
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

    function formatLocation(raw) {
        if (!raw) return raw;
        const cleaned = raw.replace(/[_-]+/g, ' ');
        return cleaned.split(' ').map(function (w) {
            if (!w) return w;
            return w.charAt(0).toUpperCase() + w.slice(1).toLowerCase();
        }).join(' ');
    }

    function formatExistingLocationsAndRelations() {
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

    if (input && dropdown) {
        let debounceTimer = null;
        input.addEventListener('input', function () {
            const q = input.value || '';
            if (q.length < 1) {
                dropdown.style.display = 'none';
                return;
            }

            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(async function () {
                try {
                    const res = await fetch('/api/search?q=' + encodeURIComponent(q));
                    const data = await res.json();
                    const results = data.results || [];

                    dropdown.innerHTML = '';
                    if (results.length > 0) {
                        results.forEach(artist => {
                            const item = document.createElement('div');
                            item.className = 'autocomplete-item';

                            // Updated display logic to include MatchedDetail
                            const name = (artist.name || artist.Name);
                            const match = (artist.matchedDetail || "");
                            item.textContent = match ? `${name} - ${match}` : name;

                            item.onclick = () => {
                                window.location.href = '/artist/' + (artist.ID || artist.id);
                            };
                            dropdown.appendChild(item);
                        });
                        dropdown.style.display = 'block';
                    } else {
                        dropdown.style.display = 'none';
                    }
                } catch (err) {
                    dropdown.style.display = 'none';
                }
            }, 150);
        });

        document.addEventListener('click', function (ev) {
            if (!form.contains(ev.target)) {
                dropdown.style.display = 'none';
            }
        });
    }

    if (form) {
        form.addEventListener('submit', function (ev) {
            ev.preventDefault();
            const q = input.value || '';
            fetch('/api/search?q=' + encodeURIComponent(q))
                .then(res => res.json())
                .then(data => renderResults(data.results))
                .catch(() => { });
        });
    }

    formatExistingLocationsAndRelations();
});