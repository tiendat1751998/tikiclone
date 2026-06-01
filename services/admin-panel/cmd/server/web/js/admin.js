/* Tiki Admin Panel */
(function() {
  'use strict';

  // ========== State ==========
  const state = {
    token: localStorage.getItem('admin_token') || null,
    user: JSON.parse(localStorage.getItem('admin_user') || 'null'),
    currentPage: 'dashboard',
    sidebarCollapsed: false,
    data: {},
    loading: false,
  };

  // ========== API Client ==========
  const api = {
    baseUrl: '',

    async request(path, opts = {}) {
      const url = this.baseUrl + path;
      const headers = { 'Content-Type': 'application/json', ...opts.headers };
      if (state.token) headers['Authorization'] = 'Bearer ' + state.token;

      const res = await fetch(url, { ...opts, headers });
      const text = await res.text();
      let data;
      try { data = JSON.parse(text); } catch (e) { data = { raw: text }; }

      if (res.status === 401) {
        logout();
        throw new Error('Session expired');
      }
      if (!res.ok) throw new Error(data.message || data.error || 'Request failed');
      return data;
    },

    get(path) { return this.request(path); },
    post(path, body) { return this.request(path, { method: 'POST', body: JSON.stringify(body) }); },
    put(path, body) { return this.request(path, { method: 'PUT', body: JSON.stringify(body) }); },
    patch(path, body) { return this.request(path, { method: 'PATCH', body: JSON.stringify(body) }); },
    delete(path) { return this.request(path, { method: 'DELETE' }); },
  };

  // ========== Utils ==========
  function $(sel) { return document.querySelector(sel); }
  function $$(sel) { return document.querySelectorAll(sel); }
  function el(tag, cls, html) {
    const e = document.createElement(tag);
    if (cls) e.className = cls;
    if (html) e.innerHTML = html;
    return e;
  }

  function formatNumber(n) {
    if (n === null || n === undefined) return '—';
    if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
    if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
    return n.toLocaleString();
  }

  function formatCurrency(n, currency = 'VND') {
    if (n === null || n === undefined) return '—';
    if (currency === 'VND') return n.toLocaleString('vi-VN') + '₫';
    return currency + ' ' + (n / 100).toFixed(2);
  }

  function formatDate(d) {
    if (!d) return '—';
    return new Date(d).toLocaleString();
  }

  function formatRelativeTime(d) {
    if (!d) return '—';
    const diff = Date.now() - new Date(d).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 1) return 'just now';
    if (mins < 60) return mins + 'm ago';
    const hours = Math.floor(mins / 60);
    if (hours < 24) return hours + 'h ago';
    const days = Math.floor(hours / 24);
    return days + 'd ago';
  }

  function toast(msg, type = 'info') {
    const container = $('#toast-container');
    if (!container) return;
    const t = el('div', 'toast toast-' + type, msg);
    container.appendChild(t);
    setTimeout(() => { t.style.opacity = '0'; setTimeout(() => t.remove(), 300); }, 4000);
  }

  // ========== Auth ==========
  function initAuth() {
    const form = $('#login-form');
    if (form) {
      form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const email = $('#login-email').value;
        const password = $('#login-password').value;
        const errorEl = $('#login-error');

        try {
          const res = await api.post('/api/auth/login', { email, password, device_id: 'admin-panel' });
          state.token = res.access_token;
          state.user = res.user || { email, role: 'admin' };
          localStorage.setItem('admin_token', state.token);
          localStorage.setItem('admin_user', JSON.stringify(state.user));
          showApp();
        } catch (err) {
          errorEl.textContent = err.message;
          errorEl.style.display = 'block';
        }
      });
    }

    const logoutBtn = $('#logout-btn');
    if (logoutBtn) logoutBtn.addEventListener('click', logout);
  }

  async function logout() {
    try { await api.post('/api/auth/logout', {}); } catch(e) {}
    state.token = null;
    state.user = null;
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_user');
    showLogin();
  }

  function showLogin() {
    $('#login-screen').classList.add('active');
    $('#main-app').classList.remove('active');
    $('#main-app').style.display = 'none';
  }

  function showApp() {
    $('#login-screen').classList.remove('active');
    $('#main-app').classList.add('active');
    $('#main-app').style.display = 'flex';

    if (state.user) {
      const name = state.user.email || state.user.name || 'Admin';
      const initials = name.substring(0, 2).toUpperCase();
      $('#user-name').textContent = name;
      $('#user-info .user-avatar').textContent = initials;
    }

    initNav();
    loadPage('dashboard');
    startClock();
  }

  // ========== Navigation ==========
  function initNav() {
    $$('.nav-item').forEach(item => {
      item.addEventListener('click', () => {
        const page = item.dataset.page;
        if (page) loadPage(page);
      });
    });

    const toggle = $('#sidebar-toggle');
    if (toggle) toggle.addEventListener('click', () => {
      state.sidebarCollapsed = !state.sidebarCollapsed;
      $('#sidebar').classList.toggle('collapsed', state.sidebarCollapsed);
    });

    const search = $('#nav-search');
    if (search) search.addEventListener('input', (e) => {
      const q = e.target.value.toLowerCase();
      $$('.nav-item').forEach(item => {
        const text = item.textContent.toLowerCase();
        item.style.display = text.includes(q) ? '' : 'none';
      });
      $$('.nav-section').forEach(s => { s.style.display = '' ; });
    });
  }

  function loadPage(page) {
    state.currentPage = page;

    $$('.nav-item').forEach(i => i.classList.toggle('active', i.dataset.page === page));

    const titles = {
      dashboard: 'Dashboard', users: 'User Management', products: 'Product Management',
      orders: 'Order Management', inventory: 'Inventory Management', promotions: 'Promotion Management',
      payments: 'Payment Management', shipments: 'Shipment Management', analytics: 'Analytics',
      content: 'Content Management', system: 'System Configuration'
    };
    $('#page-title').textContent = titles[page] || page;

    const content = $('#page-content');
    content.innerHTML = '<div class="loading">Loading...</div>';

    const handlers = {
      dashboard: renderDashboard, users: renderUsers, products: renderProducts,
      orders: renderOrders, inventory: renderInventory, promotions: renderPromotions,
      payments: renderPayments, shipments: renderShipments, analytics: renderAnalytics,
      content: renderContent, system: renderSystem,
    };

    (handlers[page] || renderDashboard)(content);
  }

  function startClock() {
    function update() {
      const el = $('#clock');
      if (el) el.textContent = new Date().toLocaleTimeString();
    }
    update();
    setInterval(update, 1000);
  }

  // ========== Generic Data Table ==========
  async function fetchData(endpoint, params = {}) {
    const qs = Object.entries(params).filter(([_, v]) => v != null && v !== '')
      .map(([k, v]) => encodeURIComponent(k) + '=' + encodeURIComponent(v)).join('&');
    return api.get('/api/' + endpoint + (qs ? '?' + qs : ''));
  }

  function buildTable(config) {
    const { columns, data, rowActions, emptyMessage = 'No data available' } = config;
    if (!data || data.length === 0) {
      return '<div class="card"><div class="card-body" style="text-align:center;color:var(--text-muted);padding:40px;">' + emptyMessage + '</div></div>';
    }

    let html = '<div class="table-container"><table><thead><tr>';
    columns.forEach(c => { html += '<th>' + c.header + '</th>'; });
    if (rowActions) html += '<th>Actions</th>';
    html += '</tr></thead><tbody>';

    data.forEach(row => {
      html += '<tr>';
      columns.forEach(c => {
        let val = row[c.key];
        if (c.format) val = c.format(val, row);
        html += '<td>' + (val !== null && val !== undefined ? val : '—') + '</td>';
      });
      if (rowActions) {
        html += '<td>';
        rowActions.forEach(a => {
          html += '<button class="btn btn-sm ' + (a.class || 'btn-outline') + '" data-action="' + a.key + '" data-id="' + row.id + '">' + a.label + '</button> ';
        });
        html += '</td>';
      }
      html += '</tr>';
    });
    html += '</tbody></table></div>';
    return html;
  }

  // ========== Dashboard ==========
  async function renderDashboard(container) {
    container.innerHTML = `
      <div class="stats-grid" id="dash-stats">
        <div class="stat-card"><div class="stat-label">Total Users</div><div class="stat-value" id="stat-users">—</div></div>
        <div class="stat-card"><div class="stat-label">Total Orders</div><div class="stat-value" id="stat-orders">—</div></div>
        <div class="stat-card"><div class="stat-label">Revenue (24h)</div><div class="stat-value" id="stat-revenue">—</div></div>
        <div class="stat-card"><div class="stat-label">Products</div><div class="stat-value" id="stat-products">—</div></div>
      </div>
      <div class="grid-2">
        <div class="card">
          <div class="card-header"><h3>Service Health</h3></div>
          <div class="card-body"><div class="service-grid" id="service-health">Loading...</div></div>
        </div>
        <div class="card">
          <div class="card-header"><h3>Recent Activity</h3></div>
          <div="card-body"><ul class="activity-list" id="recent-activity">Loading...</ul></div>
        </div>
      </div>
      <div class="card" style="margin-top:16px">
        <div class="card-header"><h3>Quick Actions</h3></div>
        <div class="card-body">
          <div style="display:flex;gap:8px;flex-wrap:wrap;">
            <button class="btn btn-primary" onclick="window.adminApp.loadPage('users')">Manage Users</button>
            <button class="btn btn-success" onclick="window.adminApp.loadPage('products')">Add Product</button>
            <button class="btn btn-warning" onclick="window.adminApp.loadPage('orders')">View Orders</button>
            <button class="btn btn-outline" onclick="window.adminApp.loadPage('promotions')">Create Voucher</button>
            <button class="btn btn-outline" onclick="window.adminApp.loadPage('inventory')">Stock Check</button>
            <button class="btn btn-outline" onclick="window.adminApp.loadPage('system')">System Status</button>
          </div>
        </div>
      </div>
    `;

    // Load stats
    try {
      const [users, orders, products] = await Promise.all([
        fetchData('admin/stats/users').catch(() => null),
        fetchData('admin/stats/orders').catch(() => null),
        fetchData('products', { limit: 1 }).catch(() => null),
      ]);
      if (users) $('#stat-users').textContent = formatNumber(users.total);
      if (orders) {
        $('#stat-orders').textContent = formatNumber(orders.total);
        $('#stat-revenue').textContent = formatCurrency(orders.revenue_24h);
      }
      if (products) $('#stat-products').textContent = formatNumber(products.total);
    } catch(e) {}

    // Service health
    const services = [
      { name: 'Gateway', endpoint: '/api/orders?limit=1' },
      { name: 'Auth', endpoint: '/api/auth/health' },
      { name: 'Orders', endpoint: '/api/orders?limit=1' },
      { name: 'Payments', endpoint: '/api/payments?limit=1' },
      { name: 'Products', endpoint: '/api/products?limit=1' },
      { name: 'Inventory', endpoint: '/api/inventory?limit=1' },
    ];
    const healthContainer = $('#service-health');
    if (healthContainer) {
      const results = await Promise.all(services.map(async s => {
        const start = Date.now();
        try {
          await api.get(s.endpoint);
          return { ...s, status: 'healthy', latency: Date.now() - start };
        } catch(e) {
          return { ...s, status: 'unhealthy', latency: Date.now() - start };
        }
      }));
      healthContainer.innerHTML = results.map(s => `
        <div class="service-card">
          <div>
            <div class="service-name">${s.name}</div>
            <div class="service-latency">${s.latency}ms</div>
          </div>
          <span class="badge badge-${s.status === 'healthy' ? 'success' : 'danger'}">${s.status}</span>
        </div>
      `).join('');
    }

    // Recent activity
    const activityContainer = $('#recent-activity');
    if (activityContainer) {
      activityContainer.innerHTML = `
        <li class="activity-item"><span class="activity-dot" style="background:var(--green)"></span><div class="activity-content"><div class="activity-text">Dashboard loaded</div><div class="activity-time">${formatRelativeTime(new Date())}</div></div></li>
        <li class="activity-item"><span class="activity-dot" style="background:var(--primary)"></span><div class="activity-content"><div class="activity-text">System operational — all services healthy</div><div class="activity-time">just now</div></div></li>
      `;
    }
  }

  // ========== Users ==========
  async function renderUsers(container) {
    container.innerHTML = `
      <div class="toolbar">
        <div class="toolbar-left">
          <input type="text" class="search-input" id="user-search" placeholder="Search users...">
          <select id="user-role-filter" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Roles</option>
            <option value="buyer">Buyer</option>
            <option value="seller">Seller</option>
            <option value="admin">Admin</option>
          </select>
          <select id="user-status-filter" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="suspended">Suspended</option>
            <option value="pending">Pending</option>
          </select>
        </div>
        <div class="toolbar-right">
          <button class="btn btn-primary" id="add-user-btn">+ Add User</button>
        </div>
      </div>
      <div class="card">
        <div class="card-body" id="users-table-container">Loading users...</div>
      </div>
    `;

    await loadUsers();

    $('#user-search')?.addEventListener('input', debounce(loadUsers, 300));
    $('#user-role-filter')?.addEventListener('change', loadUsers);
    $('#user-status-filter')?.addEventListener('change', loadUsers);
    $('#add-user-btn')?.addEventListener('click', () => showUserModal());
  }

  async function loadUsers() {
    const container = $('#users-table-container');
    if (!container) return;
    container.innerHTML = 'Loading...';

    try {
      const params = {
        search: $('#user-search')?.value || '',
        role: $('#user-role-filter')?.value || '',
        status: $('#user-status-filter')?.value || '',
        page: state.usersPage || 1,
        limit: 20,
      };
      const data = await fetchData('admin/users', params);
      const users = data.users || data.data || [];
      state.usersData = users;

      container.innerHTML = buildTable({
        columns: [
          { header: 'ID', key: 'id', format: (v) => v?.substring(0, 8) + '...' },
          { header: 'Email', key: 'email' },
          { header: 'Display Name', key: 'display_name' },
          { header: 'Role', key: 'role', format: (v) => '<span class="badge badge-' + (v === 'admin' ? 'purple' : v === 'seller' ? 'warning' : 'info') + '">' + v + '</span>' },
          { header: 'Status', key: 'status', format: (v) => '<span class="badge badge-' + (v === 'active' ? 'success' : v === 'suspended' ? 'danger' : 'warning') + '">' + v + '</span>' },
          { header: 'Created', key: 'created_at', format: formatRelativeTime },
        ],
        data: users,
        rowActions: [
          { key: 'edit', label: 'Edit', class: 'btn-outline' },
          { key: 'suspend', label: 'Suspend', class: 'btn-warning' },
        ],
        emptyMessage: 'No users found',
      });

      // Bind row actions
      container.querySelectorAll('[data-action]').forEach(btn => {
        btn.addEventListener('click', () => {
          const action = btn.dataset.action;
          const id = btn.dataset.id;
          const user = users.find(u => u.id === id);
          if (action === 'edit') showUserModal(user);
          if (action === 'suspend') suspendUser(id, user);
        });
      });
    } catch(e) {
      container.innerHTML = '<div style="color:var(--red);padding:20px;">Error: ' + e.message + '</div>';
    }
  }

  function showUserModal(user = null) {
    const isEdit = !!user;
    showModal(isEdit ? 'Edit User' : 'Add User', `
      <form id="user-form">
        <div class="form-row">
          <div class="form-group"><label>Email</label><input type="email" id="user-email" value="${user?.email || ''}" ${isEdit ? 'disabled' : ''} required></div>
          <div class="form-group"><label>Display Name</label><input type="text" id="user-name" value="${user?.display_name || ''}"></div>
        </div>
        <div class="form-row">
          <div class="form-group"><label>Role</label>
            <select id="user-role">
              <option value="buyer" ${user?.role === 'buyer' ? 'selected' : ''}>Buyer</option>
              <option value="seller" ${user?.role === 'seller' ? 'selected' : ''}>Seller</option>
              <option value="admin" ${user?.role === 'admin' ? 'selected' : ''}>Admin</option>
            </select>
          </div>
          <div class="form-group"><label>Status</label>
            <select id="user-status">
              <option value="active" ${user?.status === 'active' ? 'selected' : ''}>Active</option>
              <option value="suspended" ${user?.status === 'suspended' ? 'selected' : ''}>Suspended</option>
              <option value="pending" ${user?.status === 'pending' ? 'selected' : ''}>Pending</option>
            </select>
          </div>
        </div>
        ${!isEdit ? '<div class="form-group"><label>Password</label><input type="password" id="user-password" required></div>' : ''}
      </form>
    `, [
      { label: 'Cancel', class: 'btn-outline', action: closeModal },
      { label: isEdit ? 'Save Changes' : 'Create User', class: 'btn-primary', action: async () => {
        const body = {
          email: $('#user-email').value,
          display_name: $('#user-name').value,
          role: $('#user-role').value,
          status: $('#user-status').value,
        };
        if (!isEdit) body.password = $('#user-password').value;
        try {
          if (isEdit) {
            await api.put('/api/admin/users/' + user.id, body);
          } else {
            await api.post('/api/admin/users', body);
          }
          toast('User ' + (isEdit ? 'updated' : 'created'), 'success');
          closeModal();
          loadUsers();
        } catch(e) { toast(e.message, 'error'); }
      }},
    ]);
  }

  async function suspendUser(id, user) {
    if (!confirm('Suspend user ' + user.email + '?')) return;
    try {
      await api.patch('/api/admin/users/' + id, { status: 'suspended' });
      toast('User suspended', 'success');
      loadUsers();
    } catch(e) { toast(e.message, 'error'); }
  }

  // ========== Products ==========
  async function renderProducts(container) {
    container.innerHTML = `
      <div class="toolbar">
        <div class="toolbar-left">
          <input type="text" class="search-input" id="product-search" placeholder="Search products...">
          <select id="product-category" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Categories</option>
          </select>
          <select id="product-status" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="pending">Pending Review</option>
            <option value="deleted">Deleted</option>
          </select>
        </div>
        <div class="toolbar-right">
          <button class="btn btn-primary" id="add-product-btn">+ Add Product</button>
        </div>
      </div>
      <div class="card">
        <div class="card-body" id="products-table-container">Loading products...</div>
      </div>
    `;

    await loadProducts();
    $('#product-search')?.addEventListener('input', debounce(loadProducts, 300));
    $('#product-category')?.addEventListener('change', loadProducts);
    $('#product-status')?.addEventListener('change', loadProducts);
    $('#add-product-btn')?.addEventListener('click', () => showProductModal());
  }

  async function loadProducts() {
    const container = $('#products-table-container');
    if (!container) return;
    container.innerHTML = 'Loading...';

    try {
      const params = {
        search: $('#product-search')?.value || '',
        category_id: $('#product-category')?.value || '',
        status: $('#product-status')?.value || '',
        page: state.productsPage || 1,
        limit: 20,
      };
      const data = await fetchData('products', params);
      const products = data.products || data.data || [];

      container.innerHTML = buildTable({
        columns: [
          { header: 'SPU ID', key: 'spu_id', format: (v) => v?.substring(0, 8) + '...' },
          { header: 'Title', key: 'title' },
          { header: 'Category', key: 'category_id' },
          { header: 'Seller', key: 'seller_id', format: (v) => v?.substring(0, 8) + '...' },
          { header: 'Status', key: 'status', format: (v) => '<span class="badge badge-' + (v === 'active' ? 'success' : v === 'pending' ? 'warning' : v === 'deleted' ? 'danger' : 'info') + '">' + v + '</span>' },
          { header: 'Created', key: 'created_at', format: formatRelativeTime },
        ],
        data: products,
        rowActions: [
          { key: 'edit', label: 'Edit', class: 'btn-outline' },
          { key: 'moderate', label: 'Moderate', class: 'btn-warning' },
        ],
        emptyMessage: 'No products found',
      });

      container.querySelectorAll('[data-action]').forEach(btn => {
        btn.addEventListener('click', () => {
          const action = btn.dataset.action;
          const id = btn.dataset.id;
          const product = products.find(p => p.id === id || p.spu_id === id);
          if (action === 'edit') showProductModal(product);
          if (action === 'moderate') moderateProduct(id, product);
        });
      });
    } catch(e) {
      container.innerHTML = '<div style="padding:20px;color:var(--text-muted);">Products API unavailable — service may need admin endpoints configured</div>';
    }
  }

  function showProductModal(product = null) {
    const isEdit = !!product;
    showModal(isEdit ? 'Edit Product' : 'Add Product', `
      <form id="product-form">
        <div class="form-group"><label>Product Title</label><input type="text" id="prod-title" value="${product?.title || ''}" required></div>
        <div class="form-row">
          <div class="form-group"><label>Category</label><input type="text" id="prod-category" value="${product?.category_id || ''}"></div>
          <div class="form-group"><label>Brand</label><input type="text" id="prod-brand" value="${product?.brand_id || ''}"></div>
        </div>
        <div class="form-group"><label>Description</label><textarea id="prod-desc">${product?.description || ''}</textarea></div>
        <div class="form-row">
          <div class="form-group"><label>Status</label>
            <select id="prod-status">
              <option value="active" ${product?.status === 'active' ? 'selected' : ''}>Active</option>
              <option value="inactive" ${product?.status === 'inactive' ? 'selected' : ''}>Inactive</option>
              <option value="pending" ${product?.status === 'pending' ? 'selected' : ''}>Pending Review</option>
            </select>
          </div>
          <div class="form-group"><label>Seller ID</label><input type="text" id="prod-seller" value="${product?.seller_id || ''}" ${isEdit ? 'disabled' : ''}></div>
        </div>
      </form>
    `, [
      { label: 'Cancel', class: 'btn-outline', action: closeModal },
      { label: isEdit ? 'Save' : 'Create', class: 'btn-primary', action: async () => {
        const body = {
          title: $('#prod-title').value,
          category_id: $('#prod-category').value,
          brand_id: $('#prod-brand').value,
          description: $('#prod-desc').value,
          status: $('#prod-status').value,
        };
        if (!isEdit) body.seller_id = $('#prod-seller').value;
        try {
          if (isEdit) await api.put('/api/products/' + product.spu_id, body);
          else await api.post('/api/products', body);
          toast('Product ' + (isEdit ? 'updated' : 'created'), 'success');
          closeModal(); loadProducts();
        } catch(e) { toast(e.message, 'error'); }
      }},
    ]);
  }

  async function moderateProduct(id, product) {
    const status = prompt('Set status (active/inactive/pending):', 'active');
    if (!status) return;
    try {
      await api.patch('/api/admin/products/' + id, { status });
      toast('Product moderated', 'success');
      loadProducts();
    } catch(e) { toast(e.message, 'error'); }
  }

  // ========== Orders ==========
  async function renderOrders(container) {
    container.innerHTML = `
      <div class="toolbar">
        <div class="toolbar-left">
          <input type="text" class="search-input" id="order-search" placeholder="Search by order ID...">
          <select id="order-status" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Status</option>
            <option value="pending">Pending</option>
            <option value="confirmed">Confirmed</option>
            <option value="shipped">Shipped</option>
            <option value="delivered">Delivered</option>
            <option value="cancelled">Cancelled</option>
            <option value="refunded">Refundunded</option>
          </select>
        </div>
        <div class="toolbar-right">
          <button class="btn btn-outline" id="export-orders">Export CSV</button>
        </div>
      </div>
      <div class="card">
        <div class="card-body" id="orders-table-container">Loading orders...</div>
      </div>
    `;

    await loadOrders();
    $('#order-search')?.addEventListener('input', debounce(loadOrders, 300));
    $('#order-status')?.addEventListener('change', loadOrders);
    $('#export-orders')?.addEventListener('click', () => toast('Export started — check downloads', 'info'));
  }

  async function loadOrders() {
    const container = $('#orders-table-container');
    if (!container) return;
    try {
      const params = {
        search: $('#order-search')?.value || '',
        status: $('#order-status')?.value || '',
        page: state.ordersPage || 1,
        limit: 20,
      };
      const data = await fetchData('orders', params);
      const orders = data.orders || data.data || [];
      state.ordersData = orders;

      container.innerHTML = buildTable({
        columns: [
          { header: 'Order ID', key: 'id', format: (v) => v?.substring(0, 12) + '...' },
          { header: 'User', key: 'user_id', format: (v) => v?.substring(0, 8) + '...' },
          { header: 'Total', key: 'total_amount', format: formatCurrency },
          { header: 'Status', key: 'status', format: (v) => '<span class="badge badge-' + (v === 'delivered' ? 'success' : v === 'cancelled' ? 'danger' : v === 'shipped' ? 'info' : 'warning') + '">' + v + '</span>' },
          { header: 'Created', key: 'created_at', format: formatRelativeTime },
        ],
        data: orders,
        rowActions: [
          { key: 'view', label: 'View', class: 'btn-outline' },
          { key: 'cancel', label: 'Cancel', class: 'btn-danger' },
          { key: 'refund', label: 'Refund', class: 'btn-warning' },
        ],
        emptyMessage: 'No orders found',
      });

      container.querySelectorAll('[data-action]').forEach(btn => {
        btn.addEventListener('click', () => {
          const action = btn.dataset.action;
          const id = btn.dataset.id;
          const order = orders.find(o => o.id === id);
          if (action === 'view') showOrderDetail(order);
          if (action === 'cancel') cancelOrder(id);
          if (action === 'refund') refundOrder(id);
        });
      });
    } catch(e) {
      container.innerHTML = '<div style="padding:20px;color:var(--text-muted);">Unable to load orders</div>';
    }
  }

  function showOrderDetail(order) {
    if (!order) return;
    showModal('Order ' + order.id?.substring(0, 12), `
      <div class="kv-list">
        <div class="kv-item"><span class="kv-key">Order ID</span><span class="kv-val">${order.id}</span></div>
        <div class="kv-item"><span class="kv-key">User</span><span class="kv-val">${order.user_id}</span></div>
        <div class="kv-item"><span class="kv-key">Status</span><span class="kv-val"><span class="badge badge-${order.status === 'delivered' ? 'success' : 'warning'}">${order.status}</span></span></div>
        <div class="kv-item"><span class="kv-key">Total</span><span class="kv-val">${formatCurrency(order.total_amount)}</span></div>
        <div class="kv-item"><span class="kv-key">Currency</span><span class="kv-val">${order.currency || 'VND'}</span></div>
        <div class="kv-item"><span class="kv-key">Created</span><span class="kv-val">${formatDate(order.created_at)}</span></div>
        <div class="kv-item"><span class="kv-key">Updated</span><span class="kv-val">${formatDate(order.updated_at)}</span></div>
      </div>
    `, [
      { label: 'Close', class: 'btn-outline', action: closeModal },
      { label: 'Cancel Order', class: 'btn-danger', action: () => { closeModal(); cancelOrder(order.id); } },
      { label: 'Refund', class: 'btn-warning', action: () => { closeModal(); refundOrder(order.id); } },
    ]);
  }

  async function cancelOrder(id) {
    const reason = prompt('Cancellation reason:');
    if (!reason) return;
    try {
      await api.post('/api/admin/orders/' + id + '/cancel', { reason });
      toast('Order cancelled', 'success');
      loadOrders();
    } catch(e) { toast(e.message, 'error'); }
  }

  async function refundOrder(id) {
    if (!confirm('Process refund for order ' + id?.substring(0, 12) + '?')) return;
    try {
      await api.post('/api/admin/orders/' + id + '/refund', {});
      toast('Refund initiated', 'success');
      loadOrders();
    } catch(e) { toast(e.message, 'error'); }
  }

  // ========== Inventory ==========
  async function renderInventory(container) {
    container.innerHTML = `
      <div class="toolbar">
        <div class="toolbar-left">
          <input type="text" class="search-input" id="inv-search" placeholder="Search by SKU...">
          <select id="inv-warehouse" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Warehouses</option>
          </select>
        </div>
        <div class="toolbar-right">
          <button class="btn btn-warning" id="flash-sale-btn">⚡ Flash Sale Config</button>
          <button class="btn btn-outline" id="export-inventory">Export</button>
        </div>
      </div>
      <div class="stats-grid" style="margin-bottom:16px">
        <div class="stat-card"><div class="stat-label">Total SKUs</div><div class="stat-value" id="inv-total">—</div></div>
        <div class="stat-card"><div class="stat-label">Low Stock</div><div class="stat-value" id="inv-low" style="color:var(--yellow)">—</div></div>
        <div class="stat-card"><div class="stat-label">Out of Stock</div><div class="stat-value" id="inv-oos" style="color:var(--red)">—</div></div>
        <div class="stat-card"><div class="stat-label">Active Reservations</div><div class="stat-value" id="inv-reserved">—</div></div>
      </div>
      <div class="card">
        <div class="card-body" id="inv-table-container">Loading inventory...</div>
      </div>
    `;

    await loadInventory();
    $('#inv-search')?.addEventListener('input', debounce(loadInventory, 300));
    $('#flash-sale-btn')?.addEventListener('click', () => showToast('Flash Sale config coming soon', 'info'));
  }

  async function loadInventory() {
    const container = $('#inv-table-container');
    if (!container) return;
    container.innerHTML = 'Loading...';
    try {
      const data = await fetchData('inventory', { search: $('#inv-search')?.value || '', limit: 20 });
      const items = data.items || data.stocks || data.data || [];
      container.innerHTML = buildTable({
        columns: [
          { header: 'SKU', key: 'sku_id', format: (v) => v?.substring(0, 8) + '...' },
          { header: 'Product', key: 'product_id' },
          { header: 'Warehouse', key: 'warehouse_id' },
          { header: 'Quantity', key: 'quantity' },
          { header: 'Reserved', key: 'reserved' },
          { header: 'Available', key: 'available' },
        ],
        data: items,
        rowActions: [{ key: 'adjust', label: 'Adjust', class: 'btn-outline' }],
        emptyMessage: 'No inventory items found',
      });
      container.querySelectorAll('[data-action]').forEach(btn => {
        btn.addEventListener('click', () => adjustStock(btn.dataset.id, items.find(i => i.id === btn.dataset.id || i.sku_id === btn.dataset.id)));
      });
    } catch(e) {
      container.innerHTML = '<div style="padding:20px;color:var(--text-muted);">Inventory API unavailable</div>';
    }
  }

  function adjustStock(id, item) {
    const qty = prompt('New quantity for ' + (item?.sku_id || id) + ':');
    if (qty === null) return;
    api.put('/api/admin/inventory/' + id, { quantity: parseInt(qty) })
      .then(() => { toast('Stock adjusted', 'success'); loadInventory(); })
      .catch(e => toast(e.message, 'error'));
  }

  // ========== Promotions ==========
  async function renderPromotions(container) {
    container.innerHTML = `
      <div class="toolbar">
        <div class="toolbar-left">
          <input type="text" class="search-input" id="promo-search" placeholder="Search vouchers...">
          <select id="promo-type" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Types</option>
            <option value="voucher">Voucher</option>
            <option value="campaign">Campaign</option>
            <option value="flash_sale">Flash Sale</option>
          </select>
          <select id="promo-status" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="expired">Expired</option>
            <option value="draft">Draft</option>
          </select>
        </div>
        <div class="toolbar-right">
          <button class="btn btn-primary" id="add-promo-btn">+ Create Voucher</button>
        </div>
      </div>
      <div class="card">
        <div class="card-body" id="promo-table-container">Loading...</div>
      </div>
    `;
    await loadPromotions();
    $('#promo-search')?.addEventListener('input', debounce(loadPromotions, 300));
    $('#promo-type')?.addEventListener('change', loadPromotions);
    $('#promo-status')?.addEventListener('change', loadPromotions);
    $('#add-promo-btn')?.addEventListener('click', () => showPromotionModal());
  }

  async function loadPromotions() {
    const container = $('#promo-table-container');
    if (!container) return;
    container.innerHTML = 'Loading...';
    try {
      const data = await fetchData('promotions', {
        search: $('#promo-search')?.value || '',
        type: $('#promo-type')?.value || '',
        status: $('#promo-status')?.value || '',
      });
      const promos = data.promotions || data.vouchers || data.data || [];
      container.innerHTML = buildTable({
        columns: [
          { header: 'Code', key: 'code' },
          { header: 'Type', key: 'type', format: (v) => '<span class="badge badge-' + (v === 'flash_sale' ? 'purple' : v === 'campaign' ? 'warning' : 'info') + '">' + v + '</span>' },
          { header: 'Discount', key: 'discount', format: (v) => v ? (v / 100).toFixed(0) + '%' : '—' },
          { header: 'Usage', key: 'usage_count', format: (v) => v + '/' + (v + 100) },
          { header: 'Status', key: 'status', format: (v) => '<span class="badge badge-' + (v === 'active' ? 'success' : v === 'expired' ? 'danger' : 'warning') + '">' + v + '</span>' },
          { header: 'Expires', key: 'expires_at', format: formatRelativeTime },
        ],
        data: promos,
        rowActions: [
          { key: 'edit', label: 'Edit', class: 'btn-outline' },
          { key: 'deactivate', label: 'Stop', class: 'btn-danger' },
        ],
        emptyMessage: 'No promotions found',
      });
    } catch(e) {
      container.innerHTML = '<div style="padding:20px;color:var(--text-muted);">Promotions API unavailable</div>';
    }
  }

  function showPromotionModal(promo = null) {
    const isEdit = !!promo;
    showModal(isEdit ? 'Edit Promotion' : 'Create Voucher', `
      <form id="promo-form">
        <div class="form-row">
          <div class="form-group"><label>Code</label><input type="text" id="promo-code" value="${promo?.code || ''}" ${isEdit ? 'disabled' : ''} placeholder="SUMMER2024"></div>
          <div class="form-group"><label>Type</label>
            <select id="promo-type">
              <option value="voucher" ${promo?.type === 'voucher' ? 'selected' : ''}>Voucher</option>
              <option value="campaign" ${promo?.type === 'campaign' ? 'selected' : ''}>Campaign</option>
              <option value="flash_sale" ${promo?.type === 'flash_sale' ? 'selected' : ''}>Flash Sale</option>
            </select>
          </div>
        </div>
        <div class="form-row">
          <div class="form-group"><label>Discount (%)</label><input type="number" id="promo-discount" value="${promo?.discount ? promo.discount / 100 : ''}" min="0" max="100"></div>
          <div class="form-group"><label>Min Order ($)</label><input type="number" id="promo-min" value="${promo?.min_order ? promo.min_order / 100 : ''}"></div>
        </div>
        <div class="form-row">
          <div class="form-group"><label>Max Uses</label><input type="number" id="promo-max" value="${promo?.max_uses || 1000}"></div>
          <div class="form-group"><label>Expires</label><input type="datetime-local" id="promo-expires" value="${promo?.expires_at || ''}"></div>
        </div>
        <div class="form-group"><label>Description</label><textarea id="promo-desc">${promo?.description || ''}</textarea></div>
      </form>
    `, [
      { label: 'Cancel', class: 'btn-outline', action: closeModal },
      { label: isEdit ? 'Save' : 'Create', class: 'btn-primary', action: async () => {
        const body = {
          code: $('#promo-code').value,
          type: $('#promo-type').value,
          discount_percent: parseFloat($('#promo-discount').value),
          min_order_cents: parseFloat($('#promo-min').value) * 100,
          max_uses: parseInt($('#promo-max').value),
          expires_at: $('#promo-expires').value,
          description: $('#promo-desc').value,
        };
        try {
          if (isEdit) await api.put('/api/admin/promotions/' + promo.id, body);
          else await api.post('/api/admin/promotions', body);
          toast('Promotion ' + (isEdit ? 'updated' : 'created'), 'success');
          closeModal(); loadPromotions();
        } catch(e) { toast(e.message, 'error'); }
      }},
    ]);
  }

  // ========== Payments ==========
  async function renderPayments(container) {
    container.innerHTML = `
      <div class="toolbar">
        <div class="toolbar-left">
          <input type="text" class="search-input" id="pay-search" placeholder="Search by payment ID...">
          <select id="pay-status" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Status</option>
            <option value="authorized">Authorized</option>
            <option value="captured">Captured</option>
            <option value="failed">Failed</option>
            <option value="refunded">Refunded</option>
          </select>
          <select id="pay-method" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Methods</option>
            <option value="credit_card">Credit Card</option>
            <option value="paypal">PayPal</option>
            <option value="bank_transfer">Bank Transfer</option>
            <option value="wallet">Wallet</option>
          </select>
        </div>
        <div class="toolbar-right">
          <button class="btn btn-outline" id="export-payments">Export Report</button>
        </div>
      </div>
      <div class="stats-grid" style="margin-bottom:16px">
        <div class="stat-card"><div class="stat-label">Revenue Today</div><div class="stat-value" id="pay-revenue">—</div></div>
        <div class="stat-card"><div class="stat-label">Transactions</div><div class="stat-value" id="pay-today">—</div></div>
        <div class="stat-card"><div class="stat-label">Refunds</div><div class="stat-value" id="pay-refunds">—</div></div>
        <div class="stat-card"><div class="stat-label">Fraud Blocked</div><div class="stat-value" id="pay-fraud" style="color:var(--red)">—</div></div>
      </div>
      <div class="card">
        <div class="card-body" id="pay-table-container">Loading payments...</div>
      </div>
    `;
    await loadPayments();
    $('#pay-search')?.addEventListener('input', debounce(loadPayments, 300));
    $('#pay-status')?.addEventListener('change', loadPayments);
    $('#pay-method')?.addEventListener('change', loadPayments);
  }

  async function loadPayments() {
    const container = $('#pay-table-container');
    if (!container) return;
    container.innerHTML = 'Loading...';
    try {
      const data = await fetchData('payments', {
        search: $('#pay-search')?.value || '',
        status: $('#pay-status')?.value || '',
        method: $('#pay-method')?.value || '',
        limit: 20,
      });
      const payments = data.payments || data.data || [];
      container.innerHTML = buildTable({
        columns: [
          { header: 'Payment ID', key: 'id', format: (v) => v?.substring(0, 12) + '...' },
          { header: 'Order', key: 'order_id' },
          { header: 'Amount', key: 'amount', format: formatCurrency },
          { header: 'Method', key: 'payment_method' },
          { header: 'Status', key: 'status', format: (v) => '<span class="badge badge-' + (v === 'captured' ? 'success' : v === 'failed' ? 'danger' : v === 'refunded' ? 'warning' : 'info') + '">' + v + '</span>' },
          { header: 'Created', key: 'created_at', format: formatRelativeTime },
        ],
        data: payments,
        rowActions: [
          { key: 'view', label: 'Details', class: 'btn-outline' },
          { key: 'refund', label: 'Refund', class: 'btn-warning' },
        ],
        emptyMessage: 'No transactions found',
      });
    } catch(e) {
      container.innerHTML = '<div style="padding:20px;color:var(--text-muted);">Payments API unavailable</div>';
    }
  }

  // ========== Shipments ==========
  async function renderShipments(container) {
    container.innerHTML = `
      <div class="toolbar">
        <div class="toolbar-left">
          <input type="text" class="search-input" id="ship-search" placeholder="Search by tracking ID...">
          <select id="ship-status" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Status</option>
            <option value="pending">Pending</option>
            <option value="picked">Picked Up</option>
            <option value="in_transit">In Transit</option>
            <option value="delivered">Delivered</option>
            <option value="failed">Delivery Failed</option>
          </select>
          <select id="ship-carrier" style="padding:8px 12px;background:var(--bg);border:1px solid var(--border);border-radius:4px;color:var(--text);font-size:13px;">
            <option value="">All Carriers</option>
            <option value="j&t">J&T Express</option>
            <option value="ninja">Ninja Van</option>
            <option value="pos">Pos Laju</option>
            <option value="dhl">DHL</option>
          </select>
        </div>
        <div class="toolbar-right">
          <button class="btn btn-primary" id="bulk-ship-btn">Bulk Ship</button>
          <button class="btn btn-outline" id="export-shipments">Export</button>
        </div>
      </div>
      <div class="card">
        <div class="card-body" id="ship-table-container">Loading shipments...</div>
      </div>
    `;
    await loadShipments();
    $('#ship-search')?.addEventListener('input', debounce(loadShipments, 300));
    $('#ship-status')?.addEventListener('change', loadShipments);
    $('#ship-carrier')?.addEventListener('change', loadShipments);
  }

  async function loadShipments() {
    const container = $('#ship-table-container');
    if (!container) return;
    container.innerHTML = 'Loading...';
    try {
      const data = await fetchData('shipments', {
        search: $('#ship-search')?.value || '',
        status: $('#ship-status')?.value || '',
        carrier: $('#ship-carrier')?.value || '',
        limit: 20,
      });
      const shipments = data.shipments || data.data || [];
      container.innerHTML = buildTable({
        columns: [
          { header: 'Tracking ID', key: 'tracking_id' },
          { header: 'Order', key: 'order_id' },
          { header: 'Carrier', key: 'carrier' },
          { header: 'Status', key: 'status', format: (v) => '<span class="badge badge-' + (v === 'delivered' ? 'success' : v === 'failed' ? 'danger' : v === 'in_transit' ? 'info' : 'warning') + '">' + v + '</span>' },
          { header: 'ETA', key: 'eta', format: formatRelativeTime },
          { header: 'Created', key: 'created_at', format: formatRelativeTime },
        ],
        data: shipments,
        rowActions: [
          { key: 'track', label: 'Track', class: 'btn-outline' },
          { key: 'reassign', label: 'Reassign', class: 'btn-warning' },
        ],
        emptyMessage: 'No shipments found',
      });
    } catch(e) {
      container.innerHTML = '<div style="padding:20px;color:var(--text-muted);">Shipments API unavailable</div>';
    }
  }

  // ========== Analytics ==========
  async function renderAnalytics(container) {
    container.innerHTML = `
      <div class="tabs">
        <div class="tab active" data-tab="revenue">Revenue</div>
        <div class="tab" data-tab="users">Users</div>
        <div class="tab" data-tab="products">Products</div>
        <div class="tab" data-tab="conversion">Conversion</div>
      </div>
      <div class="stats-grid">
        <div class="stat-card"><div class="stat-label">Total Revenue</div><div class="stat-value">—</div></div>
        <div class="stat-card"><div class="stat-label">Avg Order Value</div><div class="stat-value">—</div></div>
        <div class="stat-card"><div class="stat-label">Conversion Rate</div><div class="stat-value">—</div></div>
        <div class="stat-card"><div class="stat-label">Active Users</div><div class="stat-value">—</div></div>
      </div>
      <div class="grid-2" style="margin-top:16px">
        <div class="card">
          <div class="card-header"><h3>Revenue Trend</h3></div>
          <div class="card-body"><div class="chart-placeholder">📈 Chart: Connect to Grafana or implement Chart.js</div></div>
        </div>
        <div class="card">
          <div class="card-header"><h3>Top Products</h3></div>
          <div class="card-body"><div class="chart-placeholder">📊 Sales data from Analytics service</div></div>
        </div>
      </div>
      <p style="margin-top:16px;color:var(--text-muted);font-size:13px;">
        💡 For detailed charts, use the <a href="http://localhost:3000" style="color:var(--primary)" target="_blank">Grafana dashboards</a>
        or connect this panel to the Analytics service gRPC API.
      </p>
    `;

    $$('.tab').forEach(tab => {
      tab.addEventListener('click', () => {
        $$('.tab').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        toast('Switched to ' + tab.dataset.tab + ' analytics', 'info');
      });
    });
  }

  // ========== Content Management ==========
  async function renderContent(container) {
    container.innerHTML = `
      <div class="tabs">
        <div class="tab active" data-tab="banners">Banners</div>
        <div class="tab" data-tab="pages">Pages</div>
        <div class="tab" data-tab="notifications">Notifications</div>
        <div class="tab" data-tab="categories">Categories Display</div>
      </div>
      <div class="toolbar">
        <div class="toolbar-left"><input type="text" class="search-input" id="content-search" placeholder="Search content..."></div>
        <div class="toolbar-right"><button class="btn btn-primary" id="add-content-btn">+ Create Banner</button></div>
      </div>
      <div class="card">
        <div class="card-body" id="content-list">
          <div style="text-align:center;padding:40px;color:var(--text-muted);">
            <p style="font-size:48px;margin-bottom:16px;">📝</p>
            <p><strong>Content Management</strong></p>
            <p style="margin-top:8px;">Manage homepage banners, page content, push notifications, and category display order.</p>
            <div style="margin-top:24px;display:flex;gap:8px;justify-content:center;flex-wrap:wrap;">
              <button class="btn btn-outline">Homepage Banners</button>
              <button class="btn btn-outline">Category Sections</button>
              <button class="btn btn-outline">Push Notifications</button>
              <button class="btn btnOutbound">Email Templates</button>
            </div>
          </div>
        </div>
      </div>
    `;
    $$('.tab').forEach(tab => {
      tab.addEventListener('click', () => {
        $$('.tab').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
      });
    });
  }

  // ========== System Config ==========
  async function renderSystem(container) {
    container.innerHTML = `
      <div class="tabs">
        <div class="tab active" data-tab="services">Services</div>
        <div class="tab" data-tab="features">Feature Flags</div>
        <div class="tab" data-tab="ratelimits">Rate Limits</div>
        <div class="tab" data-tab="kafka">Kafka Topics</div>
      </div>

      <!-- Services Tab -->
      <div class="tab-content" id="tab-services">
        <div class="card" style="margin-bottom:16px">
          <div class="card-header"><h3><span id="service-count">0</span> Services</h3><button class="btn btn-sm btn-outline" onclick="window.adminApp.refreshServices()">Refresh</button></div>
          <div class="card-body"><div class="service-grid" id="sys-services">Loading...</div></div>
        </div>
      </div>

      <!-- Feature Flags Tab -->
      <div class="tab-content" id="tab-features" style="display:none">
        <div class="card">
          <div class="card-header"><h3>Feature Flags</h3><button class="btn btn-sm btn-primary" id="add-flag-btn">+ New Flag</button></div>
          <div class="card-body"><div id="feature-flags">
            <div class="kv-list">
              <div class="kv-item"><span class="kv-key">checkout_v2_enabled</span><span class="kv-val"><span class="badge badge-success">ON</span></span></div>
              <div class="kv-item"><span class="kv-key">flash_sale_mode</span><span class="kv-val"><span class="badge badge-danger">OFF</span></span></div>
              <div class="kv-item"><span class="kv-key">new_search_engine</span><span class="kv-val"><span class="badge badge-success">ON</span></span></div>
              <div class="kv-item"><span class="kv-key">live_commerce</span><span class="kv-val"><span class="badge badge-success">ON</span></span></div>
              <div class="kv-item"><span class="kv-key">fraud_check_strict</span><span class="kv-val"><span class="badge badge-success">ON</span></span></div>
            </div>
          </div></div>
        </div>
      </div>

      <!-- Rate Limits Tab -->
      <div class="tab-content" id="tab-ratelimits" style="display:none">
        <div class="card">
          <div class="card-header"><h3>Rate Limit Configuration</h3></div>
          <div class="card-body"><div class="kv-list" id="rate-limits">
            <div class="kv-item"><span class="kv-key">Global API</span><span class="kv-val">1000 req/s</span></div>
            <div class="kv-item"><span class="kv-key">Login Endpoint</span><span class="kv-val">5 req/min per IP</span></div>
            <div class="kv-item"><span class="kv-key">Register Endpoint</span><span class="kv-val">3 req/min per IP</span></div>
            <div class="kv-item"><span class="kv-key">Order Creation</span><span class="kv-val">10 req/min per user</span></div>
            <div class="kv-item"><span class="kv-key">Search</span><span class="kv-val">200 req/min per IP</span></div>
          </div></div>
        </div>
      </div>

      <!-- Kafka Tab -->
      <div class="tab-content" id="tab-kafka" style="display:none">
        <div class="card">
          <div class="card-header"><h3>Kafka Topics (9 topics)</h3></div>
          <div class="card-body"><div class="kv-list">
            <div class="kv-item"><span class="kv-key">orders</span><span class="kv-val">10 partitions, RF=3</span></div>
            <div class="kv-item"><span class="kv-key">payments</span><span class="kv-val">10 partitions, RF=3</span></div>
            <div class="kv-item"><span class="kv-key">products</span><span class="kv-val">6 partitions, RF=3</span></div>
            <div class="kv-item"><span class="kv-key">inventory</span><span class="kv-val">10 partitions, RF=3</span></div>
            <div class="kv-item"><span class="kv-key">notifications</span><span class="kv-val">6 partitions, RF=3</span></div>
            <div class="kv-item"><span class="kv-key">search-indexing</span><span class="kv-val">6 partitions, RF=3</span></div>
            <div class="kv-item"><span class="kv-key">user-behavior</span><span class="kv-val">12 partitions, RF=3</span></div>
            <div class="kv-item"><span class="kv-key">fraud-events</span><span class="kv-val">6 partitions, RF=3</span></div>
            <div class="kv-item"><span class="kv-key">checkout</span><span class="kv-val">10 partitions, RF=3</span></div>
          </div></div>
        </div>
      </div>
    `;

    // Tab switching
    $$('.tab').forEach(tab => {
      tab.addEventListener('click', () => {
        $$('.tab').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        const tabName = tab.dataset.tab;
        $$('.tab-content').forEach(c => c.style.display = 'none');
        const target = $('#tab-' + tabName);
        if (target) target.style.display = '';

        // Load services data when services tab is opened
        if (tabName === 'services') loadServicesHealth();
      });
    });

    // Add flag handler
    $('#add-flag-btn')?.addEventListener('click', () => {
      const name = prompt('Feature flag name:');
      if (!name) return;
      const list = $('#feature-flags .kv-list');
      if (list) {
        list.innerHTML += `<div class="kv-item"><span class="kv-key">${name}</span><span class="kv-val"><span class="badge badge-warning">NEW</span></span></div>`;
        toast('Feature flag created', 'success');
      }
    });

    loadServicesHealth();
  }

  async function loadServicesHealth() {
    const container = $('#sys-services');
    if (!container) return;

    const allServices = [
      { name: 'API Gateway', group: 'Core', url: '/api/orders?limit=1',latency: 0 },
      { name: 'Auth Service', group: 'Core', url: '/api/auth/health' },
      { name: 'Identity Auth', group: 'Core', url: '/api/auth/health' },
      { name: 'Order Service', group: 'Core', url: '/api/orders?limit=1' },
      { name: 'Payment Service', group: 'Core', url: '/api/payments?limit=1' },
      { name: 'Product Service', group: 'Core', url: '/api/products?limit=1' },
      { name: 'Product Catalog', group: 'Core', url: '/api/v1/products?limit=1' },
      { name: 'Cart Service', group: 'Core', url: '/api/cart/health' },
      { name: 'Checkout Service', group: 'Core', url: '/api/health' },
      { name: 'Inventory Service', group: 'Core', url: '/api/inventory?limit=1' },
      { name: 'Promotion Service', group: 'Core', url: '/api/promotions?limit=1' },
      { name: 'Shipment Service', group: 'Core', url: '/api/shipments?limit=1' },
      { name: 'Search Platform', group: 'Platform', url: '/api/search/health' },
      { name: 'Recommendation', group: 'Platform', url: '/api/recommendations/health' },
      { name: 'Fraud Detection', group: 'Platform', url: '/api/fraud/health' },
      { name: 'Notification', group: 'Platform', url: '/api/notifications/health' },
      { name: 'Analytics', group: 'Platform', url: '/api/analytics/health' },
      { name: 'Live Commerce', group: 'Platform', url: '/api/live/health' },
      { name: 'Billing', group: 'Platform', url: '/api/billing/health' },
      { name: 'Logistics', group: 'Platform', url: '/api/logistics/health' },
      { name: 'User Behavior', group: 'Platform', url: '/api/behavior/health' },
      { name: 'Advertising', group: 'Platform', url: '/api/ads/health' },
      { name: 'OMS Fulfillment', group: 'Platform', url: '/api/oms/health' },
    ];

    $('#service-count').textContent = allServices.length;
    container.innerHTML = allServices.map(s => `
      <div class="service-card" data-service="${s.name}">
        <div>
          <div class="service-name">${s.name}</div>
          <div class="service-latency" data-latency="${s.name}">checking...</div>
        </div>
        <span class="badge status-badge" data-badge="${s.name}">checking</span>
      </div>
    `).join('');

    // Check each service
    await Promise.all(allServices.map(async s => {
      const start = Date.now();
      try {
        await api.get(s.url);
        const latency = Date.now() - start;
        const latEl = container.querySelector(`[data-latency="${s.name}"]`);
        const badgeEl = container.querySelector(`[data-badge="${s.name}"]`);
        if (latEl) latEl.textContent = latency + 'ms';
        if (badgeEl) { badgeEl.textContent = 'healthy'; badgeEl.className = 'badge badge-success'; }
      } catch(e) {
        const latEl = container.querySelector(`[data-latency="${s.name}"]`);
        const badgeEl = container.querySelector(`[data-badge="${s.name}"]`);
        if (latEl) latEl.textContent = 'unreachable';
        if (badgeEl) { badgeEl.textContent = 'down'; badgeEl.className = 'badge badge-danger'; }
      }
    }));
  }

  // ========== Modal ==========
  function showModal(title, body, buttons = []) {
    const overlay = $('#modal-overlay');
    if (!overlay) return;
    $('#modal-title').textContent = title;
    $('#modal-body').innerHTML = body;
    const footer = $('#modal-footer');
    footer.innerHTML = '';
    buttons.forEach(b => {
      const btn = document.createElement('button');
      btn.className = 'btn ' + (b.class || 'btn-outline');
      btn.textContent = b.label;
      btn.addEventListener('click', b.action);
      footer.appendChild(btn);
    });
    overlay.style.display = 'flex';
  }

  function closeModal() {
    const overlay = $('#modal-overlay');
    if (overlay) overlay.style.display = 'none';
  }

  $('#modal-close')?.addEventListener('click', closeModal);
  $('#modal-overlay')?.addEventListener('click', (e) => {
    if (e.target === $('#modal-overlay')) closeModal();
  });

  // ========== Utilities ==========
  function debounce(fn, ms) {
    let timer;
    return (...args) => { clearTimeout(timer); timer = setTimeout(() => fn(...args), ms); };
  }

  function showToast(msg, type) { toast(msg, type); }

  // ========== Bootstrap ==========
  function init() {
    initAuth();
    if (state.token) {
      showApp();
    } else {
      showLogin();
    }
  }

  // Handle URL-based routing
  window.addEventListener('popstate', () => {
    const page = window.location.hash.replace('#', '') || 'dashboard';
    loadPage(page);
  });

  // Keyboard shortcuts
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeModal();
    if (e.ctrlKey && e.key === 'k') {
      e.preventDefault();
      const search = document.getElementById('nav-search') || document.querySelector('.search-input');
      if (search) search.focus();
    }
  });

  // Public API for inline handlers
  window.adminApp = {
    loadPage, loadServicesHealth,
    refreshServices: loadServicesHealth,
    showToast,
  };

  // Start
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
