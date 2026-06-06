document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('searchForm');
    const input = document.getElementById('q');
    const dropdown = document.getElementById('dropdown');
    const filterBar = document.getElementById('filterBar');
    const filterToggle = document.getElementById('filterToggle');
    const catalog = document.getElementById('catalog');
    const resultCount = document.getElementById('result-count');
    const emptyState = document.getElementById('empty-state');

    function escapeHtml(s) {
        return String(s).replace(/[&<>"']/g, function (c) {
            return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": "&#39;" }[c];
        });
    }

    function countLabel(n, noun) {
        return n === 1 ? '1 ' + noun : n + ' ' + noun + 's';
    }

    // Mirrors writeArtistCard in server.go so filtered cards match the page.
    function renderCard(r) {
        const id = r.id != null ? r.id : r.ID;
        const name = r.name || r.Name || '';
        const image = r.image || r.Image || '';
        const creation = r.creationDate || r.CreationDate || 0;
        const members = r.memberCount || r.MemberCount || 0;
        const locations = r.locationCount || r.LocationCount || 0;
        return '<li class="catalog-item"><a class="catalog-link" href="/artist/' + id + '">' +
            '<img src="' + escapeHtml(image) + '" alt="' + escapeHtml(name) + '" loading="lazy" decoding="async">' +
            '<div class="catalog-card"><h3>' + escapeHtml(name) + '</h3>' +
            '<p class="catalog-card__meta">Since ' + creation + '</p>' +
            '<ul class="card-stats"><li>' + escapeHtml(countLabel(members, 'member')) + '</li>' +
            '<li>' + escapeHtml(countLabel(locations, 'location')) + '</li></ul></div></a></li>';
    }

    // Collect the text query plus every active filter/sort control.
    function buildQuery() {
        const params = new URLSearchParams();
        const q = ((input && input.value) || '').trim();
        if (q) {
            params.set('q', q);
        }
        if (filterBar) {
            ['sort', 'minYear', 'maxYear', 'minMembers', 'maxMembers', 'country'].forEach(function (name) {
                const el = filterBar.elements[name];
                if (el && el.value) {
                    params.set(name, el.value);
                }
            });
        }
        return params;
    }

    /* ----- autocomplete dropdown state ----- */
    let activeIndex = -1;

    function options() {
        return dropdown ? Array.from(dropdown.querySelectorAll('.autocomplete-item')) : [];
    }

    function isOpen() {
        return dropdown && dropdown.style.display === 'block';
    }

    function openDropdown() {
        if (!dropdown) return;
        dropdown.style.display = 'block';
        if (input) input.setAttribute('aria-expanded', 'true');
    }

    function closeDropdown() {
        if (!dropdown) return;
        dropdown.style.display = 'none';
        activeIndex = -1;
        if (input) {
            input.setAttribute('aria-expanded', 'false');
            input.removeAttribute('aria-activedescendant');
        }
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
        } else if (input) {
            input.removeAttribute('aria-activedescendant');
        }
    }

    function populateDropdown(results) {
        if (!dropdown) return;
        dropdown.innerHTML = '';
        activeIndex = -1;
        const top = results.slice(0, 8);
        if (top.length === 0) {
            closeDropdown();
            return;
        }
        top.forEach(function (artist, i) {
            const item = document.createElement('div');
            item.className = 'autocomplete-item';
            item.id = 'ac-option-' + i;
            item.setAttribute('role', 'option');
            item.setAttribute('aria-selected', 'false');
            const name = artist.name || artist.Name || '';
            const match = artist.matchedDetail || '';
            item.textContent = match ? name + ' - ' + match : name;
            const target = '/artist/' + (artist.id != null ? artist.id : artist.ID);
            item.addEventListener('mouseenter', function () { setActive(i); });
            item.addEventListener('click', function () { window.location.href = target; });
            dropdown.appendChild(item);
        });
        openDropdown();
    }

    /* ----- the shared client-server search/filter request ----- */
    let searchTimer = null;

    async function runSearch(updateDropdown) {
        const params = buildQuery();
        let data;
        try {
            const res = await fetch('/api/search?' + params.toString());
            data = await res.json();
        } catch (err) {
            return; // keep the current grid on a network error
        }
        const results = data.results || [];

        if (catalog) {
            catalog.innerHTML = results.map(renderCard).join('');
        }
        if (emptyState) {
            emptyState.hidden = results.length !== 0;
        }
        if (resultCount) {
            const total = data.total != null ? data.total : results.length;
            resultCount.textContent = 'Showing ' + results.length + ' of ' + total + ' artists';
        }

        if (updateDropdown) {
            const q = ((input && input.value) || '').trim();
            if (q.length < 1) {
                closeDropdown();
            } else {
                populateDropdown(results);
            }
        }
    }

    function scheduleSearch(updateDropdown) {
        clearTimeout(searchTimer);
        searchTimer = setTimeout(function () { runSearch(updateDropdown); }, 180);
    }

    if (input) {
        input.addEventListener('input', function () {
            scheduleSearch(true);
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
    }

    if (form) {
        form.addEventListener('submit', function (ev) {
            ev.preventDefault();
            closeDropdown();
            runSearch(false);
        });
    }

    if (filterToggle && filterBar) {
        filterToggle.addEventListener('click', function () {
            const open = filterBar.classList.toggle('is-open');
            filterToggle.setAttribute('aria-expanded', open ? 'true' : 'false');
        });
    }

    if (filterBar) {
        filterBar.addEventListener('change', function () { scheduleSearch(false); });
        filterBar.addEventListener('input', function (ev) {
            if (ev.target && ev.target.type === 'number') {
                scheduleSearch(false);
            }
        });
        // The Clear button is a native form reset; re-run once values are cleared.
        filterBar.addEventListener('reset', function () {
            setTimeout(function () { runSearch(false); }, 0);
        });
    }

    document.addEventListener('click', function (ev) {
        if (form && !form.contains(ev.target)) {
            closeDropdown();
        }
    });
});
