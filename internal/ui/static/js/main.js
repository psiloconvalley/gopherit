// gopherit.dev — Minimal vanilla JS
// No external dependencies. Strict mode.

(() => {
    "use strict";

    // ── Theme Toggle ──────────────────────────────────
    const toggle = document.getElementById("theme-toggle");
    const html = document.documentElement;
    const STORAGE_KEY = "gopherit-theme";

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

    // ── Mobile Navigation Drawer ─────────────────────
    const menuToggle = document.getElementById("menu-toggle");
    const navLinks = document.getElementById("nav-links");
    const body = document.body;

    if (menuToggle && navLinks) {
        const toggleMenu = (forceClose = false) => {
            const isOpen = menuToggle.getAttribute("aria-expanded") === "true";
            const shouldOpen = forceClose ? false : !isOpen;

            menuToggle.setAttribute("aria-expanded", shouldOpen.toString());
            navLinks.classList.toggle("active", shouldOpen);
            body.classList.toggle("menu-open", shouldOpen);
        };

        menuToggle.addEventListener("click", () => toggleMenu());

        // Close menu when clicking any nav link
        navLinks.querySelectorAll("a").forEach(link => {
            link.addEventListener("click", () => toggleMenu(true));
        });

        // Close menu when clicking outside of the navigation bar
        document.addEventListener("click", (e) => {
            const isClickInside = navLinks.contains(e.target) || menuToggle.contains(e.target);
            if (!isClickInside && navLinks.classList.contains("active")) {
                toggleMenu(true);
            }
        });
    }

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
        reveals.forEach((el) => el.classList.add("visible"));
    }
})();
