document.addEventListener('DOMContentLoaded', () => {
    // ── 1. Mobile Menu Toggle Handler ────────────────
    const menuToggle = document.querySelector('.menu-toggle');
    const navLinks = document.querySelector('.nav-links');

    if (menuToggle && navLinks) {
        menuToggle.addEventListener('click', () => {
            const isExpanded = menuToggle.getAttribute('aria-expanded') === 'true';
            menuToggle.setAttribute('aria-expanded', !isExpanded);
            navLinks.classList.toggle('active');
            
            // Toggle hamburger animation state
            const icon = menuToggle.querySelector('.hamburger-inner');
            if (icon) {
                icon.classList.toggle('active');
            }
        });
    }

    // ── 2. GitHub Projects API Hydrator ──────────────
    const projectsContainer = document.getElementById('github-projects');

    if (projectsContainer) {
        fetch('/api/projects')
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.json();
            })
            .then(repos => {
                if (!repos || repos.length === 0) {
                    projectsContainer.innerHTML = '<p class="projects-empty">No public repositories found.</p>';
                    return;
                }

                // Clear the loading message
                projectsContainer.innerHTML = '';

                // Loop through repositories and build high-end cards
                repos.forEach((repo, index) => {
                    const card = document.createElement('article');
                    card.className = 'project-card reveal';
                    
                    // Format index with a leading zero (e.g., 01, 02)
                    const displayIndex = String(index + 1).padStart(2, '0');
                    
                    // Fallback for empty descriptions
                    const description = repo.description || 'No description provided. Click source to view codebase.';
                    
                    // Fallback for missing primary language
                    const languageTag = repo.language ? `<span>${repo.language}</span>` : '';

                    card.innerHTML = `
                        <div class="project-header">
                            <span class="project-number">${displayIndex}</span>
                            <h3>${repo.name}</h3>
                        </div>
                        <p>${description}</p>
                        <div class="project-tech">
                            ${languageTag}
                            <span>GitHub</span>
                        </div>
                        <div class="project-links">
                            <a href="${repo.html_URL}" target="_blank" rel="noopener noreferrer">
                                Source ↗
                            </a>
                        </div>
                    `;

                    projectsContainer.appendChild(card);
                });
            })
            .catch(error => {
                console.error('Failed to load GitHub projects:', error);
                projectsContainer.innerHTML = `
                    <div class="projects-error">
                        <p>Failed to load projects directly from GitHub API.</p>
                        <a href="https://github.com/psiloconvalley" target="_blank" rel="noopener noreferrer" class="btn btn-ghost">
                            View on GitHub ↗
                        </a>
                    </div>
                `;
            });
    }
});
