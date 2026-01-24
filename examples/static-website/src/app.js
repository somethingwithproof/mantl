/**
 * Simple JavaScript for static website
 * Demonstrates health check API integration and dynamic status updates
 */

// Smooth scrolling for anchor links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
            target.scrollIntoView({
                behavior: 'smooth',
                block: 'start'
            });
        }
    });
});

// Fetch and update platform status (optional backend integration)
async function updatePlatformStatus() {
    const statusGrid = document.getElementById('status-grid');
    if (!statusGrid) return;

    // In production, this would fetch from actual health check endpoints
    const services = [
        { name: 'API Services', status: 'ok' },
        { name: 'Databases', status: 'ok' },
        { name: 'Message Queue', status: 'ok' },
        { name: 'Monitoring', status: 'ok' }
    ];

    // Clear existing status items
    while (statusGrid.firstChild) {
        statusGrid.removeChild(statusGrid.firstChild);
    }

    // Create status items using safe DOM methods
    services.forEach(service => {
        const statusItem = document.createElement('div');
        statusItem.className = 'status-item';

        const indicator = document.createElement('span');
        indicator.className = `status-indicator status-${service.status}`;

        const label = document.createElement('span');
        label.className = 'status-label';
        label.textContent = service.name;

        statusItem.appendChild(indicator);
        statusItem.appendChild(label);
        statusGrid.appendChild(statusItem);
    });
}

// Update status on page load
document.addEventListener('DOMContentLoaded', () => {
    updatePlatformStatus();

    // Refresh status every 30 seconds
    setInterval(updatePlatformStatus, 30000);
});

// Add active class to current nav item based on scroll position
window.addEventListener('scroll', () => {
    const sections = document.querySelectorAll('section[id]');
    const navLinks = document.querySelectorAll('.nav-links a');

    let current = '';
    sections.forEach(section => {
        const sectionTop = section.offsetTop;
        const sectionHeight = section.clientHeight;
        if (window.scrollY >= sectionTop - 100) {
            current = section.getAttribute('id');
        }
    });

    navLinks.forEach(link => {
        link.classList.remove('active');
        if (link.getAttribute('href') === `#${current}`) {
            link.classList.add('active');
        }
    });
});
