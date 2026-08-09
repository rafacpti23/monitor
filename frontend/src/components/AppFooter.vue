<template>
  <footer class="app-footer">
    <div class="footer-content">
      <span class="footer-copyright">
        © {{ currentYear }} <a href="https://wa.me/5527999082624" target="_blank" rel="noopener noreferrer" class="footer-link">Rafa Martins</a>
      </span>
      <span class="footer-separator">·</span>
      <span class="footer-brand">P-mon</span>
    </div>
  </footer>
</template>

<script setup>
const currentYear = new Date().getFullYear()
</script>

<style scoped>
.app-footer {
  padding: 24px 0 16px;
  text-align: center;
  border-top: 1px solid var(--border);
  margin-top: 32px;
}

.footer-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.footer-link {
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.2s ease;
}

.footer-link:hover {
  color: var(--accent);
}

.footer-separator {
  opacity: 0.4;
}

.footer-brand {
  font-family: 'JetBrains Mono', monospace;
  font-weight: 600;
  color: var(--text-secondary);
}
</style>
