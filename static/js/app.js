
document.addEventListener("DOMContentLoaded", () => {
    // Service Worker
    if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/sw.js");
    }

    // --- Weather & Autocomplete ---
    const locationInput = document.getElementById("location");
    const searchBtn = document.getElementById("search-btn");
    const locateBtn = document.getElementById("locate-btn");
    const suggestionsList = document.getElementById("suggestions");
    const weatherDisplay = document.getElementById("weather-display");
    let debounceTimer;

    // Search input handler (debounce)
    if (locationInput) {
        locationInput.addEventListener("input", (e) => {
            clearTimeout(debounceTimer);
            const query = e.target.value.trim();
            if (query.length < 2) {
                suggestionsList.classList.remove("visible");
                return;
            }
            debounceTimer = setTimeout(() => {
                fetch(`/api/search?q=${encodeURIComponent(query)}`)
                    .then(r => r.json())
                    .then(places => {
                        suggestionsList.innerHTML = "";
                        if (places && places.length > 0) {
                            places.forEach(p => {
                                const li = document.createElement("li");
                                li.innerHTML = `<span class="place-name">${p.name}</span><span class="place-meta">${p.state} ${p.zip || ''}</span>`;
                                li.onclick = () => {
                                    locationInput.value = `${p.name}, ${p.state}`;
                                    suggestionsList.classList.remove("visible");
                                    fetchWeather(`lat=${p.latitude}&lon=${p.longitude}`);
                                };
                                suggestionsList.appendChild(li);
                            });
                            suggestionsList.classList.add("visible");
                        } else {
                            suggestionsList.classList.remove("visible");
                        }
                    })
                    .catch(console.error);
            }, 300);
        });

        // Hide suggestions on outside click
        document.addEventListener("click", (e) => {
            if (locationInput && suggestionsList && !locationInput.contains(e.target) && !suggestionsList.contains(e.target)) {
                suggestionsList.classList.remove("visible");
            }
        });

        // Enter key to search
        locationInput.addEventListener("keydown", (e) => {
            if (e.key === "Enter") {
                manualLocate();
            }
        });
    }

    // Manual Search Button
    if (searchBtn) {
        searchBtn.addEventListener("click", manualLocate);
    }

    // Locate Me Button
    if (locateBtn) {
        locateBtn.addEventListener("click", locateMe);
    }

    function manualLocate() {
        const query = locationInput.value.trim();
        if (!query) return;
        fetchWeather(`location=${encodeURIComponent(query)}`);
        suggestionsList.classList.remove("visible");
    }

    // Geolocation logic
    function locateMe(silent = false) {
        if (!navigator.geolocation) {
            if (!silent) alert("Geolocation is not supported by your browser");
            return;
        }
        const originalText = locateBtn.innerText;

        if (!silent) {
            locateBtn.innerText = "Locating...";
            locateBtn.disabled = true;
        }

        navigator.geolocation.getCurrentPosition(
            (position) => {
                let qs = `lat=${position.coords.latitude}&lon=${position.coords.longitude}`;
                if (!silent) {
                    qs += "&userInitiated=1";
                }
                fetchWeather(qs);

                if (!silent) {
                    locateBtn.innerText = originalText;
                    locateBtn.disabled = false;
                }
            },
            (error) => {
                if (!silent) {
                    alert("Unable to retrieve location.");
                    locateBtn.innerText = originalText;
                    locateBtn.disabled = false;
                } else {
                    console.log("Auto-location unavailable or denied");
                }
            }
        );
    }

    // Auto-locate if permission granted
    if (navigator.geolocation) {
        navigator.permissions.query({ name: "geolocation" }).then((result) => {
            if (result.state === "granted" || result.state === "prompt") {
                locateMe(true);
            }
        });
    }

    // Fetch Weather (HTML Fragment)
    function fetchWeather(qs) {
        weatherDisplay.classList.add("is-loading");
        fetch(`/api/weather?${qs}`)
            .then(r => {
                if (!r.ok) throw new Error(r.statusText);
                return r.text();
            })
            .then(html => {
                weatherDisplay.innerHTML = html;
                weatherDisplay.classList.remove("is-loading");
                // Update input if userInitiated
                if (qs.includes("userInitiated=1")) {
                    const resolved = weatherDisplay.querySelector("#resolved-location");
                    if (resolved && resolved.dataset.location) {
                        locationInput.value = resolved.dataset.location;
                    }
                }
            })
            .catch(e => {
                // Safely render error message without injecting HTML
                weatherDisplay.innerHTML = "";
                const errorCard = document.createElement("div");
                errorCard.className = "error-card";
                const heading = document.createElement("h3");
                heading.textContent = "Error";
                const messagePara = document.createElement("p");
                messagePara.textContent = e.message;
                errorCard.appendChild(heading);
                errorCard.appendChild(messagePara);
                weatherDisplay.appendChild(errorCard);
                weatherDisplay.classList.remove("is-loading");
            });
    }

    // --- App Interest Modal ---
    const modal = document.getElementById("app-interest-modal");
    const openLink = document.getElementById("app-interest-link");
    const closeBtn = document.getElementById("app-interest-close");
    const cancelBtn = document.getElementById("app-interest-cancel");
    const overlay = document.getElementById("app-interest-overlay");
    const form = document.getElementById("app-interest-form");
    const errorDiv = document.getElementById("app-interest-error");
    const thanksDiv = document.getElementById("app-interest-thanks");
    const countrySelect = document.getElementById("app-country");

    function showModal(e) {
        if (e) e.preventDefault();
        modal.setAttribute("aria-hidden", "false");
        loadCountries();
    }

    function hideModal() {
        modal.setAttribute("aria-hidden", "true");
        if (errorDiv) errorDiv.classList.remove("is-visible");
        if (form) form.classList.remove("u-hidden");
        if (thanksDiv) thanksDiv.classList.add("u-hidden");
        if (form) form.reset();
    }

    if (openLink) openLink.addEventListener("click", showModal);
    if (closeBtn) closeBtn.addEventListener("click", hideModal);
    if (cancelBtn) cancelBtn.addEventListener("click", hideModal);
    if (overlay) overlay.addEventListener("click", hideModal);

    // Country Loader
    let countriesLoaded = false;
    function loadCountries() {
        if (countriesLoaded || !countrySelect) return;
        countrySelect.innerHTML = '<option value="">Loading...</option>';
        fetch("/static/countries.json")
            .then(r => r.json())
            .then(data => {
                countrySelect.innerHTML = '<option value="">Select your country</option>';
                data.forEach(c => {
                    const opt = document.createElement("option");
                    opt.value = c;
                    opt.textContent = c;
                    countrySelect.appendChild(opt);
                });
                countriesLoaded = true;
            })
            .catch(() => {
                countrySelect.innerHTML = '<option value="">Unable to load</option>';
            });
    }

    // Form Submission
    if (form) {
        form.addEventListener("submit", (e) => {
            e.preventDefault();
            errorDiv.classList.remove("is-visible");
            const formData = new FormData(form);
            const payload = {
                email: formData.get("email"),
                android: form.querySelector("#app-android").checked,
                ios: form.querySelector("#app-ios").checked,
                country: formData.get("country")
            };

            // Validation
            if (!payload.email || !payload.country || (!payload.android && !payload.ios)) {
                errorDiv.textContent = "Please fill in all fields and select platform.";
                errorDiv.classList.add("is-visible");
                return;
            }

            const btn = form.querySelector('button[type="submit"]');
            btn.disabled = true;

            fetch("/api/app-interest", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload)
            })
                .then(r => {
                    if (!r.ok) throw new Error("Submission failed");
                    return r.json();
                })
                .then(() => {
                    form.classList.add("u-hidden");
                    thanksDiv.classList.remove("u-hidden");
                    setTimeout(hideModal, 2500);
                })
                .catch(err => {
                    errorDiv.textContent = err.message;
                    errorDiv.classList.add("is-visible");
                })
                .finally(() => {
                    btn.disabled = false;
                });
        });
    }
});
