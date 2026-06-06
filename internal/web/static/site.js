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
        let activeIndex = -1;

        function options() {
            return Array.from(dropdown.querySelectorAll('.autocomplete-item'));
        }

        function isOpen() {
            return dropdown.style.display === 'block';
        }

        function openDropdown() {
            dropdown.style.display = 'block';
            input.setAttribute('aria-expanded', 'true');
        }

        function closeDropdown() {
            dropdown.style.display = 'none';
            input.setAttribute('aria-expanded', 'false');
            input.removeAttribute('aria-activedescendant');
            activeIndex = -1;
        }

        function setActive(index) {
            const opts = options();
            opts.forEach(function (opt, i) {
                const active = i === index;
                opt.classList.toggle('is-active', active);
                opt.setAttribute('aria-selected', active ? 'true' : 'false');
            });
            activeIndex = index;
            if (index >= 0 && opts[index]) {
                input.setAttribute('aria-activedescendant', opts[index].id);
                opts[index].scrollIntoView({ block: 'nearest' });
            } else {
                input.removeAttribute('aria-activedescendant');
            }
        }

        input.addEventListener('input', function () {
            const q = input.value || '';
            if (q.length < 1) {
                closeDropdown();
                return;
            }

            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(async function () {
                try {
                    const res = await fetch('/api/search?q=' + encodeURIComponent(q));
                    const data = await res.json();
                    const results = data.results || [];

                    dropdown.innerHTML = '';
                    activeIndex = -1;
                    if (results.length > 0) {
                        results.forEach((artist, i) => {
                            const item = document.createElement('div');
                            item.className = 'autocomplete-item';
                            item.id = 'ac-option-' + i;
                            item.setAttribute('role', 'option');
                            item.setAttribute('aria-selected', 'false');

                            const name = (artist.name || artist.Name);
                            const match = (artist.matchedDetail || '');
                            item.textContent = match ? `${name} - ${match}` : name;

                            const target = '/artist/' + (artist.id || artist.ID);
                            item.addEventListener('mouseenter', () => setActive(i));
                            item.addEventListener('click', () => {
                                window.location.href = target;
                            });
                            dropdown.appendChild(item);
                        });
                        openDropdown();
                    } else {
                        closeDropdown();
                    }
                } catch (err) {
                    closeDropdown();
                }
            }, 150);
        });

        input.addEventListener('keydown', function (ev) {
            const opts = options();
            if (!isOpen() || opts.length === 0) {
                return;
            }
            switch (ev.key) {
                case 'ArrowDown':
                    ev.preventDefault();
                    setActive((activeIndex + 1) % opts.length);
                    break;
                case 'ArrowUp':
                    ev.preventDefault();
                    setActive((activeIndex - 1 + opts.length) % opts.length);
                    break;
                case 'Enter':
                    if (activeIndex >= 0 && opts[activeIndex]) {
                        ev.preventDefault();
                        opts[activeIndex].click();
                    }
                    break;
                case 'Escape':
                    closeDropdown();
                    break;
            }
        });

        document.addEventListener('click', function (ev) {
            if (!form.contains(ev.target)) {
                closeDropdown();
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