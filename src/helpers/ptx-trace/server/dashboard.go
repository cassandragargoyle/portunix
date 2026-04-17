/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package server

// dashboardHTML contains the embedded web dashboard
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PTX-TRACE Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: #18181b;
            color: #e4e4e7;
            min-height: 100vh;
        }
        .header {
            background: #27272a;
            padding: 1rem 2rem;
            border-bottom: 1px solid #3f3f46;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 {
            font-size: 1.5rem;
            font-weight: 600;
            color: #a1a1aa;
        }
        .header .status {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: #22c55e;
        }
        .status-dot.disconnected {
            background: #ef4444;
        }
        .container {
            display: grid;
            grid-template-columns: 300px 1fr;
            height: calc(100vh - 60px);
        }
        .sidebar {
            background: #27272a;
            border-right: 1px solid #3f3f46;
            overflow-y: auto;
        }
        .sidebar-header {
            padding: 1rem;
            border-bottom: 1px solid #3f3f46;
            font-weight: 600;
            color: #a1a1aa;
            text-transform: uppercase;
            font-size: 0.75rem;
            letter-spacing: 0.05em;
        }
        .session-list {
            list-style: none;
        }
        .session-item {
            padding: 0.75rem 1rem;
            border-bottom: 1px solid #3f3f46;
            cursor: pointer;
            transition: background 0.15s;
            position: relative;
        }
        .session-item:hover {
            background: #3f3f46;
        }
        .session-item.active {
            background: #52525b;
            color: white;
        }
        .delete-btn {
            position: absolute;
            top: 0.5rem;
            right: 0.5rem;
            width: 24px;
            height: 24px;
            border: none;
            background: transparent;
            color: #71717a;
            cursor: pointer;
            border-radius: 4px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 1rem;
            opacity: 0;
            transition: all 0.15s;
        }
        .session-item:hover .delete-btn {
            opacity: 1;
        }
        .delete-btn:hover {
            background: #ef4444;
            color: white;
        }
        .session-name {
            font-weight: 500;
            margin-bottom: 0.25rem;
        }
        .session-meta {
            font-size: 0.75rem;
            color: #71717a;
            display: flex;
            gap: 0.5rem;
        }
        .session-item.active .session-meta {
            color: rgba(255,255,255,0.7);
        }
        .main {
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        .stats-bar {
            display: flex;
            gap: 1rem;
            padding: 1rem;
            background: #27272a;
            border-bottom: 1px solid #3f3f46;
        }
        .stat-card {
            background: #18181b;
            padding: 0.75rem 1rem;
            border-radius: 0.5rem;
            min-width: 120px;
        }
        .stat-label {
            font-size: 0.75rem;
            color: #71717a;
            text-transform: uppercase;
            margin-bottom: 0.25rem;
        }
        .stat-value {
            font-size: 1.5rem;
            font-weight: 600;
        }
        .stat-value.success { color: #22c55e; }
        .stat-value.error { color: #ef4444; }
        .stat-value.warning { color: #f59e0b; }
        .events-container {
            flex: 1;
            overflow-y: auto;
            padding: 1rem;
        }
        .event-item {
            background: #27272a;
            border-radius: 0.5rem;
            padding: 0.75rem 1rem;
            margin-bottom: 0.5rem;
            border-left: 3px solid #71717a;
        }
        .event-item.level-error {
            border-left-color: #ef4444;
        }
        .event-item.level-warning {
            border-left-color: #f59e0b;
        }
        .event-item.level-info {
            border-left-color: #71717a;
        }
        .event-item.level-debug {
            border-left-color: #71717a;
        }
        .event-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 0.5rem;
        }
        .event-operation {
            font-weight: 500;
            color: #f4f4f5;
        }
        .event-time {
            font-size: 0.75rem;
            color: #71717a;
        }
        .event-details {
            font-size: 0.875rem;
            color: #a1a1aa;
        }
        .event-error {
            margin-top: 0.5rem;
            padding: 0.5rem;
            background: rgba(239, 68, 68, 0.1);
            border-radius: 0.25rem;
            color: #fca5a5;
            font-family: monospace;
            font-size: 0.75rem;
        }
        .badge {
            display: inline-block;
            padding: 0.125rem 0.5rem;
            border-radius: 9999px;
            font-size: 0.625rem;
            font-weight: 600;
            text-transform: uppercase;
        }
        .badge-error { background: #ef4444; color: white; }
        .badge-warning { background: #f59e0b; color: white; }
        .badge-info { background: #71717a; color: white; }
        .badge-debug { background: #71717a; color: white; }
        .badge-active { background: #22c55e; color: white; }
        .badge-completed { background: #71717a; color: white; }
        .empty-state {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100%;
            color: #71717a;
        }
        .empty-state svg {
            width: 64px;
            height: 64px;
            margin-bottom: 1rem;
            opacity: 0.5;
        }
        .filters {
            display: flex;
            gap: 0.5rem;
            padding: 0.5rem 1rem;
            background: #27272a;
            border-bottom: 1px solid #3f3f46;
        }
        .filter-btn {
            padding: 0.375rem 0.75rem;
            border: 1px solid #3f3f46;
            background: transparent;
            color: #a1a1aa;
            border-radius: 0.375rem;
            cursor: pointer;
            font-size: 0.75rem;
            transition: all 0.15s;
        }
        .filter-btn:hover {
            background: #3f3f46;
        }
        .filter-btn.active {
            background: #52525b;
            border-color: #52525b;
            color: white;
        }
        .modal-overlay {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.7);
            display: none;
            align-items: center;
            justify-content: center;
            z-index: 1000;
        }
        .modal-overlay.active {
            display: flex;
        }
        .modal {
            background: #27272a;
            border-radius: 0.5rem;
            padding: 1.5rem;
            max-width: 400px;
            width: 90%;
            border: 1px solid #3f3f46;
        }
        .modal-title {
            font-size: 1.125rem;
            font-weight: 600;
            margin-bottom: 0.5rem;
        }
        .modal-message {
            color: #a1a1aa;
            margin-bottom: 1.5rem;
        }
        .modal-actions {
            display: flex;
            justify-content: flex-end;
            gap: 0.75rem;
        }
        .modal-btn {
            padding: 0.5rem 1rem;
            border-radius: 0.375rem;
            cursor: pointer;
            font-size: 0.875rem;
            border: 1px solid #3f3f46;
            transition: all 0.15s;
        }
        .modal-btn-cancel {
            background: transparent;
            color: #a1a1aa;
        }
        .modal-btn-cancel:hover {
            background: #3f3f46;
        }
        .modal-btn-delete {
            background: #ef4444;
            color: white;
            border-color: #ef4444;
        }
        .modal-btn-delete:hover {
            background: #dc2626;
        }
        .modal-btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>PTX-TRACE Dashboard</h1>
        <div class="status">
            <span class="status-dot" id="ws-status"></span>
            <span id="ws-status-text">Connecting...</span>
        </div>
    </div>
    <div class="container">
        <div class="sidebar">
            <div class="sidebar-header">Sessions</div>
            <ul class="session-list" id="session-list"></ul>
        </div>
        <div class="main">
            <div class="stats-bar" id="stats-bar">
                <div class="stat-card">
                    <div class="stat-label">Total Events</div>
                    <div class="stat-value" id="stat-total">0</div>
                </div>
                <div class="stat-card">
                    <div class="stat-label">Success</div>
                    <div class="stat-value success" id="stat-success">0</div>
                </div>
                <div class="stat-card">
                    <div class="stat-label">Errors</div>
                    <div class="stat-value error" id="stat-errors">0</div>
                </div>
                <div class="stat-card">
                    <div class="stat-label">Warnings</div>
                    <div class="stat-value warning" id="stat-warnings">0</div>
                </div>
            </div>
            <div class="filters">
                <button class="filter-btn active" data-level="all">All</button>
                <button class="filter-btn" data-level="error">Errors</button>
                <button class="filter-btn" data-level="warning">Warnings</button>
                <button class="filter-btn" data-level="info">Info</button>
                <button class="filter-btn" data-level="debug">Debug</button>
            </div>
            <div class="events-container" id="events-container">
                <div class="empty-state">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                    </svg>
                    <p>Select a session to view events</p>
                </div>
            </div>
        </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div class="modal-overlay" id="delete-modal">
        <div class="modal">
            <div class="modal-title">Delete Session</div>
            <div class="modal-message" id="delete-modal-message">
                Are you sure you want to delete this session? This action cannot be undone.
            </div>
            <div class="modal-actions">
                <button class="modal-btn modal-btn-cancel" id="delete-cancel">Cancel</button>
                <button class="modal-btn modal-btn-delete" id="delete-confirm">Delete</button>
            </div>
        </div>
    </div>

    <script>
        let ws = null;
        let currentSession = null;
        let currentFilter = 'all';
        let events = [];

        // WebSocket connection
        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            ws = new WebSocket(protocol + '//' + window.location.host + '/ws');

            ws.onopen = () => {
                document.getElementById('ws-status').classList.remove('disconnected');
                document.getElementById('ws-status-text').textContent = 'Connected';
            };

            ws.onclose = () => {
                document.getElementById('ws-status').classList.add('disconnected');
                document.getElementById('ws-status-text').textContent = 'Disconnected';
                setTimeout(connectWebSocket, 3000);
            };

            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                handleWebSocketMessage(data);
            };
        }

        function handleWebSocketMessage(data) {
            // Handle session deletion event
            if (data.type === 'session_deleted') {
                if (data.session_id === currentSession) {
                    currentSession = null;
                    events = [];
                    renderEvents();
                    clearStats();
                }
                loadSessions();
                return;
            }

            if (data.session_id === currentSession) {
                events.unshift(data);
                renderEvents();
            }
            loadSessions();
        }

        // Clear stats display
        function clearStats() {
            document.getElementById('stat-total').textContent = '0';
            document.getElementById('stat-success').textContent = '0';
            document.getElementById('stat-errors').textContent = '0';
            document.getElementById('stat-warnings').textContent = '0';
        }

        // Delete modal handling
        let deleteSessionId = null;

        function showDeleteModal(sessionId, sessionName) {
            deleteSessionId = sessionId;
            document.getElementById('delete-modal-message').textContent =
                'Are you sure you want to delete session "' + sessionName + '"? This action cannot be undone.';
            document.getElementById('delete-modal').classList.add('active');
        }

        function hideDeleteModal() {
            deleteSessionId = null;
            document.getElementById('delete-modal').classList.remove('active');
        }

        async function confirmDelete() {
            if (!deleteSessionId) return;

            const btn = document.getElementById('delete-confirm');
            btn.disabled = true;
            btn.textContent = 'Deleting...';

            try {
                const response = await fetch('/api/sessions/' + deleteSessionId, {
                    method: 'DELETE'
                });

                if (!response.ok) {
                    const error = await response.text();
                    throw new Error(error);
                }

                // If we deleted the current session, clear the view
                if (deleteSessionId === currentSession) {
                    currentSession = null;
                    events = [];
                    renderEvents();
                    clearStats();
                }

                hideDeleteModal();
                loadSessions();
            } catch (error) {
                alert('Failed to delete session: ' + error.message);
            } finally {
                btn.disabled = false;
                btn.textContent = 'Delete';
            }
        }

        // Modal event listeners
        document.getElementById('delete-cancel').addEventListener('click', hideDeleteModal);
        document.getElementById('delete-confirm').addEventListener('click', confirmDelete);
        document.getElementById('delete-modal').addEventListener('click', (e) => {
            if (e.target.id === 'delete-modal') hideDeleteModal();
        });

        // Load sessions
        async function loadSessions() {
            try {
                const response = await fetch('/api/sessions');
                const sessions = await response.json();
                renderSessions(sessions);
            } catch (error) {
                console.error('Failed to load sessions:', error);
            }
        }

        function renderSessions(sessions) {
            const list = document.getElementById('session-list');
            list.innerHTML = sessions.map(session => {
                const isActive = session.id === currentSession;
                const statusBadge = session.status === 'active'
                    ? '<span class="badge badge-active">Active</span>'
                    : '<span class="badge badge-completed">Completed</span>';
                return '<li class="session-item' + (isActive ? ' active' : '') + '" data-id="' + session.id + '" data-name="' + escapeHtml(session.name) + '">' +
                    '<button class="delete-btn" title="Delete session" data-delete="' + session.id + '"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14M10 11v6M14 11v6"/></svg></button>' +
                    '<div class="session-name">' + escapeHtml(session.name) + '</div>' +
                    '<div class="session-meta">' +
                        statusBadge +
                        '<span>' + formatTime(session.started_at) + '</span>' +
                    '</div>' +
                '</li>';
            }).join('');

            // Add click handlers for session selection
            list.querySelectorAll('.session-item').forEach(item => {
                item.addEventListener('click', (e) => {
                    if (!e.target.classList.contains('delete-btn')) {
                        selectSession(item.dataset.id);
                    }
                });
            });

            // Add click handlers for delete buttons
            list.querySelectorAll('.delete-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    e.stopPropagation();
                    const sessionId = btn.dataset.delete;
                    const sessionName = btn.closest('.session-item').dataset.name;
                    showDeleteModal(sessionId, sessionName);
                });
            });
        }

        async function selectSession(sessionId) {
            currentSession = sessionId;
            loadSessions();

            try {
                // Load session stats
                const statsResponse = await fetch('/api/sessions/' + sessionId + '/stats');
                const stats = await statsResponse.json();
                updateStats(stats);

                // Load events
                const eventsResponse = await fetch('/api/sessions/' + sessionId + '/events?limit=100');
                events = await eventsResponse.json();
                renderEvents();
            } catch (error) {
                console.error('Failed to load session:', error);
            }
        }

        function updateStats(stats) {
            if (!stats) return;
            document.getElementById('stat-total').textContent = stats.total_events || 0;
            document.getElementById('stat-success').textContent = stats.by_status?.success || 0;
            document.getElementById('stat-errors').textContent = stats.by_status?.error || 0;
            document.getElementById('stat-warnings').textContent = stats.by_status?.warning || 0;
        }

        function renderEvents() {
            const container = document.getElementById('events-container');

            const filtered = currentFilter === 'all'
                ? events
                : events.filter(e => e.level === currentFilter);

            if (filtered.length === 0) {
                container.innerHTML = '<div class="empty-state">' +
                    '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">' +
                        '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />' +
                    '</svg>' +
                    '<p>No events found</p>' +
                '</div>';
                return;
            }

            container.innerHTML = filtered.map(event => {
                const level = event.level || 'info';
                const operation = event.operation?.name || 'Unknown';
                const errorHtml = event.error
                    ? '<div class="event-error">' + escapeHtml(event.error.code + ': ' + event.error.message) + '</div>'
                    : '';
                const duration = event.duration_us
                    ? '<span>' + (event.duration_us / 1000).toFixed(2) + 'ms</span>'
                    : '';

                return '<div class="event-item level-' + level + '">' +
                    '<div class="event-header">' +
                        '<span class="event-operation">' +
                            '<span class="badge badge-' + level + '">' + level + '</span> ' +
                            escapeHtml(operation) +
                        '</span>' +
                        '<span class="event-time">' + formatTime(event.timestamp) + ' ' + duration + '</span>' +
                    '</div>' +
                    '<div class="event-details">' +
                        (event.operation?.category ? '<span>Category: ' + escapeHtml(event.operation.category) + '</span>' : '') +
                    '</div>' +
                    errorHtml +
                '</div>';
            }).join('');
        }

        // Filter buttons
        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                currentFilter = btn.dataset.level;
                renderEvents();
            });
        });

        // Utility functions
        function escapeHtml(text) {
            if (!text) return '';
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function formatTime(timestamp) {
            if (!timestamp) return '';
            const date = new Date(timestamp);
            return date.toLocaleTimeString();
        }

        // Initialize
        connectWebSocket();
        loadSessions();
    </script>
</body>
</html>`
