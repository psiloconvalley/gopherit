// gopherit.dev — Minimal vanilla JS
// No external dependencies. Strict mode.

(() => {
    "use strict";

    // ── Theme Toggle ──────────────────────────────────
    const toggle = document.getElementById("theme-toggle");
    const html = document.documentElement;
    const STORAGE_KEY = "gopherit-theme";

    // Restore saved preference
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
        html.setAttribute("data-theme", saved);
    }

    toggle?.addEventListener("click", () => {
        const current = html.getAttribute("data-theme");
        const next = current === "dark" ? "light" : "dark";
        html.setAttribute("data-theme", next);
        localStorage.setItem(STORAGE_KEY, next);
    });

    // ── Scroll Reveal ─────────────────────────────────
    const reveals = document.querySelectorAll(".reveal");

    if (reveals.length > 0 && "IntersectionObserver" in window) {
        const observer = new IntersectionObserver(
            (entries) => {
                entries.forEach((entry) => {
                    if (entry.isIntersecting) {
                        entry.target.classList.add("visible");
                        observer.unobserve(entry.target);
                    }
                });
            },
            { threshold: 0.15 }
        );

        reveals.forEach((el) => observer.observe(el));
    } else {
        // Fallback for browsers without IntersectionObserver
        reveals.forEach((el) => el.classList.add("visible"));
    }
})();
