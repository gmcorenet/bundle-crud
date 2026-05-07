// GMCore CRUD Bundle - Core JavaScript Runtime
// (c) 2026 gmcore/crud-bundle - MIT License

(function() {
  'use strict';

  class GmcoreCrudRuntime {
    constructor(options = {}) {
      this.options = Object.assign({
        searchDelayMs: 250,
        csrfToken: '',
        baseUrl: '/_gmcore/crud',
        defaultPerPage: 25,
      }, options);

      this.instances = new Map();
      this.modalHost = null;
      this.confirmHost = null;
      this.activeModal = null;
      this.assetsRendered = false;
    }

    static init(options) {
      if (window.gmcore_crud) return window.gmcore_crud;
      window.gmcore_crud = new GmcoreCrudRuntime(options);
      window.gmcore_crud.autoMount();
      return window.gmcore_crud;
    }

    autoMount() {
      document.querySelectorAll('[data-gmcrud-root]').forEach(el => {
        const resource = el.dataset.gmcrudResource;
        const instance = el.dataset.gmcrudInstance;
        if (resource && !this.instances.has(resource)) {
          this.mount(el);
        }
      });
    }

    mount(container) {
      const el = container.jquery ? container[0] : container;
      const resource = el.dataset.gmcrudResource;
      if (!resource) return null;
      if (this.instances.has(resource)) return this.instances.get(resource);

      const controller = new GmcoreCrudInstance(this, el, resource);
      controller.bind();
      this.instances.set(resource, controller);

      this.modalHost = this.modalHost || document.querySelector('[data-gmcrud-modal-host]');
      this.confirmHost = this.confirmHost || this.modalHost;

      return controller;
    }

    getInstance(name) {
      return this.instances.get(name);
    }

    async fetch(url, options = {}) {
      const headers = Object.assign({
        'X-Requested-With': 'XMLHttpRequest',
        'X-CRUD-Ajax': '1',
      }, options.headers || {});

      if (!(options.body instanceof FormData)) {
        headers['Content-Type'] = 'application/json';
      }

      if (this.options.csrfToken) {
        headers['X-CSRF-Token'] = this.options.csrfToken;
      }

      const response = await fetch(url, Object.assign({}, options, { headers }));

      if (!response.ok && response.status !== 422) {
        const text = await response.text().catch(() => '');
        throw new Error(text || `CRUD request failed: ${response.status}`);
      }

      const contentType = response.headers.get('Content-Type') || '';
      if (contentType.includes('application/json')) {
        return response.json();
      }
      return { ok: true, html: await response.text() };
    }

    openModal(html) {
      if (!this.modalHost) return;
      const modal = this.modalHost.querySelector('[data-gmcrud-modal]');
      const content = modal ? modal.querySelector('[data-gmcrud-modal-content]') : null;
      if (!modal || !content) return;

      content.innerHTML = html;
      modal.removeAttribute('hidden');
      this.activeModal = modal;
      this.bindModalClose(modal);
      this.bindModalForms(modal);
    }

    closeModal() {
      if (!this.modalHost) return;
      const modal = this.modalHost.querySelector('[data-gmcrud-modal]');
      if (modal) {
        modal.setAttribute('hidden', '');
        const content = modal.querySelector('[data-gmcrud-modal-content]');
        if (content) content.innerHTML = '';
      }
      this.activeModal = null;
    }

    bindModalClose(modal) {
      modal.querySelectorAll('[data-gmcrud-modal-close]').forEach(btn => {
        btn.addEventListener('click', (e) => {
          e.preventDefault();
          this.closeModal();
        });
      });
      modal.querySelector('[data-gmcrud-modal-close]') || modal.querySelector('.crud-gmcore-modal-backdrop')?.addEventListener('click', (e) => {
        this.closeModal();
      });
    }

    bindModalForms(modal) {
      modal.querySelectorAll('form[data-gmcrud-modal-form]').forEach(form => {
        form.addEventListener('submit', (e) => this.handleModalFormSubmit(e, form));
      });
    }

    async handleModalFormSubmit(e, form) {
      e.preventDefault();
      const mode = form.dataset.gmcrudFormMode || 'create';
      const formData = new FormData(form);
      const payload = {};
      formData.forEach((value, key) => { payload[key] = value; });

      try {
        const result = await this.fetch(form.action || window.location.href, {
          method: form.method || 'POST',
          body: JSON.stringify(payload),
        });

        if (result.ok) {
          this.closeModal();
          this.refreshAll();
          this.showToast(result.message || 'Saved successfully', result.type || 'success');
        } else {
          this.showToast(result.message || 'Failed', result.type || 'error');
        }
      } catch (err) {
        this.showToast(err.message || 'Request failed', 'error');
      }
    }

    openConfirm(message, onAccept) {
      if (!this.confirmHost) return;
      const confirmModal = this.confirmHost.querySelector('[data-gmcrud-confirm-modal]');
      if (!confirmModal) return;

      const msgEl = confirmModal.querySelector('[data-gmcrud-confirm-message]');
      const acceptBtn = confirmModal.querySelector('[data-gmcrud-confirm-accept]');
      const cancelBtns = confirmModal.querySelectorAll('[data-gmcrud-confirm-cancel]');

      if (msgEl) msgEl.textContent = message || 'Are you sure?';

      const cleanup = () => {
        confirmModal.setAttribute('hidden', '');
        if (acceptBtn) acceptBtn.replaceWith(acceptBtn.cloneNode(true));
      };

      cancelBtns.forEach(btn => {
        const handler = () => { cleanup(); };
        btn.addEventListener('click', handler, { once: true });
      });

      if (acceptBtn) {
        const newAccept = acceptBtn.cloneNode(true);
        acceptBtn.replaceWith(newAccept);
        newAccept.addEventListener('click', async () => {
          cleanup();
          if (onAccept) await onAccept();
        }, { once: true });
      }

      confirmModal.removeAttribute('hidden');
    }

    async refreshAll() {
      for (const [name, instance] of this.instances) {
        await instance.refreshTable();
      }
    }

    showToast(message, type = 'success') {
      const stack = document.querySelector('[data-gmcrud-toasts]');
      if (!stack) {
        alert(message);
        return;
      }

      const toast = document.createElement('div');
      toast.className = `gmcore-toast ${type}`;
      toast.textContent = message;
      toast.style.cssText = 'position:fixed;bottom:1rem;right:1rem;padding:0.75rem 1.25rem;border-radius:6px;color:#fff;font-size:0.875rem;z-index:2000;animation:gmcore-fade-in 0.2s ease;';
      if (type === 'success') toast.style.background = '#10b981';
      if (type === 'error' || type === 'danger') toast.style.background = '#ef4444';
      if (type === 'info') toast.style.background = '#3b82f6';
      if (type === 'warning') toast.style.background = '#f59e0b';

      document.body.appendChild(toast);
      setTimeout(() => {
        toast.style.transition = 'opacity 0.3s';
        toast.style.opacity = '0';
        setTimeout(() => toast.remove(), 300);
      }, 3000);
    }
  }

  class GmcoreCrudInstance {
    constructor(runtime, container, resource) {
      this.runtime = runtime;
      this.container = container;
      this.resource = resource;
      this.state = {
        page: 1,
        perPage: parseInt(container.querySelector('[name="per_page"]')?.value || 25),
        search: '',
        sort: [],
        filters: {},
        advancedFilters: {},
      };
      this.searchTimer = null;
    }

    bind() {
      this.bindSearch();
      this.bindSort();
      this.bindPagination();
      this.bindPerPage();
      this.bindFilters();
      this.bindReset();
      this.bindModalLinks();
      this.bindDeleteForms();
      this.bindBulkActions();
      this.bindBulkToggle();
      this.bindDebug();
      this.bindDesigner();
      this.bindActionMenu();
    }

    bindSearch() {
      const input = this.container.querySelector('[data-gmcrud-search]');
      if (!input) return;
      input.addEventListener('input', () => {
        clearTimeout(this.searchTimer);
        this.searchTimer = setTimeout(() => {
          this.state.search = input.value;
          this.state.page = 1;
          this.refreshTable();
        }, this.runtime.options.searchDelayMs);
      });
    }

    bindSort() {
      this.container.querySelectorAll('[data-gmcrud-ajax-link]').forEach(link => {
        link.addEventListener('click', (e) => {
          e.preventDefault();
          const url = link.dataset.gmcrudUrl;
          if (url) {
            this.state.page = parseInt(new URLSearchParams(url.split('?')[1] || '').get('page') || 1);
            this.state.sort = (new URLSearchParams(url.split('?')[1] || '').get('sort') || '').split(',');
            this.refreshTable();
          }
        });
      });
    }

    bindPagination() {
      this.container.querySelectorAll('.crud-gmcore-pagination a').forEach(link => {
        link.addEventListener('click', (e) => {
          e.preventDefault();
          const url = link.dataset.gmcrudUrl;
          if (url) {
            const page = parseInt(new URLSearchParams(url.split('?')[1] || '').get('page') || 1);
            this.state.page = page;
            this.refreshTable();
          }
        });
      });
    }

    bindPerPage() {
      const select = this.container.querySelector('[data-gmcrud-per-page]');
      if (!select) return;
      select.addEventListener('change', () => {
        this.state.perPage = parseInt(select.value);
        this.state.page = 1;
        this.refreshTable();
      });
    }

    bindReset() {
      const btn = this.container.querySelector('[data-gmcrud-reset]');
      if (!btn) return;
      btn.addEventListener('click', () => {
        this.state.search = '';
        this.state.page = 1;
        this.state.sort = [];
        this.state.filters = {};
        this.state.advancedFilters = {};
        const input = this.container.querySelector('[data-gmcrud-search]');
        if (input) input.value = '';
        this.refreshTable();
      });
    }

    bindFilters() {
      const openBtn = this.container.querySelector('[data-gmcrud-filters-open]');
      if (!openBtn) return;
      openBtn.addEventListener('click', async () => {
        const url = this.container.dataset.gmcrudFiltersUrl || `/_gmcore/crud/${this.resource}/filters/modal`;
        try {
          const result = await this.runtime.fetch(url);
          if (result.html) {
            this.runtime.openModal(result.html);
            const modal = this.runtime.activeModal;
            if (modal) this.bindFilterModal(modal);
          }
        } catch (e) {}
      });
    }

    bindFilterModal(modal) {
      const applyBtn = modal.querySelector('[data-gmcrud-filters-apply]');
      const clearBtn = modal.querySelector('[data-gmcrud-filters-clear]');
      const closeBtns = modal.querySelectorAll('[data-gmcrud-filters-close]');

      closeBtns.forEach(b => b.addEventListener('click', () => this.runtime.closeModal()));
      if (clearBtn) clearBtn.addEventListener('click', () => { this.state.advancedFilters = {}; this.runtime.closeModal(); this.refreshTable(); });
      if (applyBtn) applyBtn.addEventListener('click', () => {
        this.state.advancedFilters = {};
        modal.querySelectorAll('[data-gmcrud-filter-row]').forEach(row => {
          const field = row.dataset.gmcrudFilterRow;
          const op = row.querySelector('[data-gmcrud-filter-operator]')?.value || 'eq';
          const val = row.querySelector('[data-gmcrud-filter-value]')?.value || '';
          if (val) {
            this.state.advancedFilters[field] = { operator: op, value: val };
          }
        });
        this.state.page = 1;
        this.runtime.closeModal();
        this.refreshTable();
      });
    }

    bindModalLinks() {
      this.container.querySelectorAll('[data-gmcrud-modal-link]').forEach(link => {
        link.addEventListener('click', async (e) => {
          e.preventDefault();
          const url = link.dataset.gmcrudUrl || link.getAttribute('href');
          if (!url) return;
          try {
            const result = await this.runtime.fetch(url);
            if (result.html) {
              this.runtime.openModal(result.html);
            } else if (result.record) {
              this.runtime.openModal(this.renderRecordView(result));
            }
          } catch (e) {
            this.runtime.showToast(e.message, 'error');
          }
        });
      });
    }

    renderRecordView(result) {
      if (!result.record) return '';
      const fields = Object.entries(result.record).filter(([k]) => k !== 'id' && k !== 'ID').map(([k, v]) =>
        `<div class="crud-gmcore-detail-row"><div>${k}</div><div>${v != null ? v : '&mdash;'}</div></div>`
      ).join('');
      const editUrl = `/_gmcore/crud/${this.resource}/{{id}}/edit`.replace('{{id}}', result.record.id || '');
      return `
<div class="crud-gmcore-modal-panel"><div class="crud-gmcore-modal-header">
<div class="crud-gmcore-eyebrow">GMCore Admin</div><h2>${this.resource}</h2>
<button class="crud-gmcore-modal-close" data-gmcrud-modal-close>&times;</button>
</div><div class="crud-gmcore-modal-body"><div class="crud-gmcore-detail-card">${fields}</div></div>
<div class="gmui-form-actions"><button class="gmui-btn is-ghost" data-gmcrud-modal-close>Close</button>
<a href="#" class="gmui-btn is-primary" data-gmcrud-url="${editUrl}" data-gmcrud-modal-link>Edit</a></div></div>`;
    }

    bindDeleteForms() {
      this.container.querySelectorAll('[data-gmcrud-delete-form]').forEach(form => {
        form.addEventListener('submit', (e) => {
          e.preventDefault();
          const message = form.dataset.gmcrudConfirm || 'Are you sure?';
          this.runtime.openConfirm(message, async () => {
            try {
              const formData = new FormData(form);
              const result = await this.runtime.fetch(form.action, {
                method: 'POST',
                body: formData,
              });
              if (result.ok) {
                this.runtime.showToast(result.message || 'Deleted successfully', 'success');
                this.refreshTable();
              } else {
                this.runtime.showToast(result.message || 'Deletion failed', 'error');
              }
            } catch (err) {
              this.runtime.showToast(err.message, 'error');
            }
          });
        });
      });
    }

    bindBulkActions() {
      const applyBtn = this.container.querySelector('[data-gmcrud-bulk-apply]');
      const actionSelect = this.container.querySelector('[data-gmcrud-bulk-action]');
      if (!applyBtn || !actionSelect) return;

      applyBtn.addEventListener('click', async () => {
        const action = actionSelect.value;
        if (!action) {
          this.runtime.showToast('Please select an action', 'warning');
          return;
        }

        const checked = this.container.querySelectorAll('[data-gmcrud-bulk-item]:checked');
        const ids = Array.from(checked).map(cb => cb.value);
        if (ids.length === 0) {
          this.runtime.showToast('No items selected', 'warning');
          return;
        }

        if (action === 'delete_bulk') {
          this.runtime.openConfirm(`Delete ${ids.length} selected items?`, async () => {
            await this.executeBulk(action, ids);
          });
          return;
        }

        await this.executeBulk(action, ids);
      });
    }

    async executeBulk(action, ids) {
      try {
        const url = this.container.dataset.gmcrudBulkUrl || `/_gmcore/crud/${this.resource}/bulk`;
        const result = await this.runtime.fetch(url, {
          method: 'POST',
          body: JSON.stringify({ action, ids }),
        });
        if (result.ok) {
          this.runtime.showToast(result.message || 'Bulk action completed', 'success');
          this.refreshTable();
        } else {
          this.runtime.showToast(result.message || 'Bulk action failed', 'error');
        }
      } catch (err) {
        this.runtime.showToast(err.message, 'error');
      }
    }

    bindBulkToggle() {
      const toggle = this.container.querySelector('[data-gmcrud-bulk-toggle]');
      if (!toggle) return;
      toggle.addEventListener('change', () => {
        this.container.querySelectorAll('[data-gmcrud-bulk-item]').forEach(cb => {
          cb.checked = toggle.checked;
        });
      });
    }

    bindDebug() {
      const debugBtn = this.container.querySelector('[data-gmcrud-debug-open]');
      if (!debugBtn) return;
      debugBtn.addEventListener('click', async () => {
        const url = debugBtn.dataset.gmcrudDebugUrl || `/_gmcore/crud/${this.resource}/debug/modal`;
        try {
          const result = await this.runtime.fetch(url);
          if (result.html) this.runtime.openModal(result.html);
        } catch (e) {}
      });
    }

    bindDesigner() {
      const designerBtn = this.container.querySelector('[data-gmcrud-designer-open]');
      if (!designerBtn) return;
      designerBtn.addEventListener('click', async () => {
        const url = `/_gmcore/crud/${this.resource}/designer/modal`;
        try {
          const result = await this.runtime.fetch(url);
          if (result.html) this.runtime.openModal(result.html);
        } catch (e) {}
      });
    }

    bindActionMenu() {
      this.container.querySelectorAll('[data-gmcrud-actions-toggle]').forEach(btn => {
        btn.addEventListener('click', () => {
          const menu = btn.parentElement.querySelector('[data-gmcrud-actions-menu]');
          if (menu) {
            const isHidden = menu.hasAttribute('hidden');
            document.querySelectorAll('[data-gmcrud-actions-menu]').forEach(m => m.setAttribute('hidden', ''));
            if (isHidden) menu.removeAttribute('hidden');
            else menu.setAttribute('hidden', '');
          }
        });
      });

      document.addEventListener('click', (e) => {
        if (!e.target.closest('[data-gmcrud-actions-toggle]')) {
          document.querySelectorAll('[data-gmcrud-actions-menu]').forEach(m => m.setAttribute('hidden', ''));
        }
      });
    }

    async refreshTable() {
      const table = this.container.querySelector('[data-gmcrud-table]');
      if (!table) return;

      const params = new URLSearchParams({
        page: this.state.page,
        per_page: this.state.perPage,
      });
      if (this.state.search) params.set('q', this.state.search);
      if (this.state.sort.length) params.set('sort', this.state.sort.join(','));

      try {
        const url = `/_gmcore/crud/${this.resource}?${params.toString()}`;
        const result = await this.runtime.fetch(url);
        if (result.html) {
          table.innerHTML = result.html;
          this.rebindTable(table);
        } else if (result.records) {
          table.innerHTML = this.renderTable(result);
          this.rebindTable(table);
        }
      } catch (e) {
        console.error('CRUD refresh error:', e);
      }
    }

    renderTable(data) {
      const records = data.records || [];
      if (records.length === 0) return '<div class="gmui-empty">No records found</div>';
      const keys = Object.keys(records[0] || {}).filter(k => k !== 'id' && k !== 'ID');
      const headers = keys.map(k => `<th>${k}</th>`).join('');
      const rows = records.map(r => {
        const cells = keys.map(k => `<td>${r[k] != null ? r[k] : '&mdash;'}</td>`).join('');
        return `<tr>${cells}<td>Actions</td></tr>`;
      }).join('');
      return `<table class="gmui-table"><thead><tr><th>ID</th>${headers}<th>Actions</th></tr></thead><tbody>${rows}</tbody></table>`;
    }

    rebindTable(table) {
      this.bindSort();
      this.bindPagination();
      this.bindModalLinks();
      this.bindDeleteForms();
      this.bindBulkToggle();
      this.bindActionMenu();
    }
  }

  if (typeof window !== 'undefined') {
    window.GmcoreCrud = GmcoreCrudRuntime;
  }
  if (typeof module !== 'undefined' && module.exports) {
    module.exports = GmcoreCrudRuntime;
  }
})();
