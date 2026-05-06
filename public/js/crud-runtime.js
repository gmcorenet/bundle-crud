// GMCore CRUD Bundle - AJAX Runtime
//
// (c) 2026 gmcore/crud-bundle
// MIT License

const GMCORE_CRUD_BUNDLE = true;

// CRUD AJAX engine bootstrap
class GmcoreCrud {
  constructor(options = {}) {
    this.options = {
      searchDelayMs: 250,
      csrfToken: '',
      baseUrl: '/_gmcore/crud',
      ...options
    };
    this.resources = new Map();
  }

  static init(options) {
    window.gmcore_crud = new GmcoreCrud(options);
    return window.gmcore_crud;
  }

  mount(resource, container) {
    if (this.resources.has(resource)) {
      return this.resources.get(resource);
    }
    const controller = new GmcoreCrudController(this, resource, container);
    this.resources.set(resource, controller);
    return controller;
  }

  getResource(name) {
    return this.resources.get(name);
  }

  async fetch(url, options = {}) {
    const headers = {
      'X-Requested-With': 'XMLHttpRequest',
      'X-CRUD-Ajax': '1',
      'Content-Type': 'application/json',
      ...options.headers
    };

    if (this.options.csrfToken) {
      headers['X-CSRF-Token'] = this.options.csrfToken;
    }

    const response = await fetch(url, {
      ...options,
      headers
    });

    if (!response.ok && response.status !== 422) {
      throw new Error(`CRUD request failed: ${response.status}`);
    }

    const contentType = response.headers.get('Content-Type') || '';
    if (contentType.includes('application/json')) {
      return response.json();
    }

    return response.text();
  }
}

class GmcoreCrudController {
  constructor(runtime, resource, container) {
    this.runtime = runtime;
    this.resource = resource;
    this.container = container;
    this.state = {
      page: 1,
      perPage: 25,
      search: '',
      sort: [],
      filters: {}
    };
  }

  async refresh() {
    const params = new URLSearchParams({
      page: this.state.page,
      per_page: this.state.perPage,
      ...(this.state.search ? { q: this.state.search } : {})
    });

    if (this.state.sort.length > 0) {
      params.set('sort', this.state.sort.join(','));
    }

    Object.entries(this.state.filters).forEach(([key, val]) => {
      if (val) params.set(`filter_${key}`, val);
    });

    const url = `${this.runtime.options.baseUrl}/${this.resource}?${params.toString()}`;
    return this.runtime.fetch(url);
  }

  onSearch(query) {
    this.state.search = query;
    this.state.page = 1;
    return this.refresh();
  }

  onSort(fields) {
    this.state.sort = fields;
    return this.refresh();
  }

  onPage(page) {
    this.state.page = page;
    return this.refresh();
  }

  onFilter(field, value) {
    this.state.filters[field] = value;
    this.state.page = 1;
    return this.refresh();
  }

  async create(data) {
    const url = `${this.runtime.options.baseUrl}/${this.resource}/new`;
    return this.runtime.fetch(url, {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  async edit(id, data) {
    const url = `${this.runtime.options.baseUrl}/${this.resource}/${id}/edit`;
    return this.runtime.fetch(url, {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  async delete(id, token) {
    const formData = new FormData();
    formData.append('_token', token);

    const url = `${this.runtime.options.baseUrl}/${this.resource}/${id}/delete`;
    return this.runtime.fetch(url, {
      method: 'POST',
      body: formData
    });
  }

  async bulk(action, ids) {
    const url = `${this.runtime.options.baseUrl}/${this.resource}/bulk`;
    return this.runtime.fetch(url, {
      method: 'POST',
      body: JSON.stringify({ action, ids })
    });
  }

  async loadRelationOptions(relation, query = '', page = 1) {
    const params = new URLSearchParams({ q: query, page, limit: 20 });
    const url = `${this.runtime.options.baseUrl}/${this.resource}/relations/${relation}?${params.toString()}`;
    return this.runtime.fetch(url);
  }

  async openFilters() {
    const url = `${this.runtime.options.baseUrl}/${this.resource}/filters/modal`;
    return this.runtime.fetch(url);
  }
}

export default GmcoreCrud;
