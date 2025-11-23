let ws;
let username;
let token;

// Cek session saat load
window.onload = () => {
    token = localStorage.getItem('chat_token');
    username = localStorage.getItem('chat_username');
    if (token && username) {
        verifyToken();
    }
};

// Tab switching
function showTab(tabName) {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
    document.querySelector(`.tab:nth-child(${tabName === 'login' ? 1 : 2})`).classList.add('active');
    document.getElementById(tabName + '-tab').classList.add('active');
    clearErrors();
}

function clearErrors() {
    document.getElementById('login-error').textContent = '';
    document.getElementById('reg-error').textContent = '';
    document.getElementById('reg-success').textContent = '';
}

// Register
async function register() {
    const user = document.getElementById('reg-username').value.trim();
    const pass = document.getElementById('reg-password').value;
    const pass2 = document.getElementById('reg-password2').value;

    clearErrors();

    if (pass !== pass2) {
        document.getElementById('reg-error').textContent = 'Password tidak cocok!';
        return;
    }

    try {
        const res = await fetch('/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: user, password: pass })
        });
        const data = await res.json();

        if (data.success) {
            document.getElementById('reg-success').textContent = data.message;
            document.getElementById('reg-username').value = '';
            document.getElementById('reg-password').value = '';
            document.getElementById('reg-password2').value = '';
            setTimeout(() => showTab('login'), 1500);
        } else {
            document.getElementById('reg-error').textContent = data.message;
        }
    } catch (e) {
        document.getElementById('reg-error').textContent = 'Gagal terhubung ke server';
    }
}

// Login
async function login() {
    const user = document.getElementById('login-username').value.trim();
    const pass = document.getElementById('login-password').value;

    clearErrors();

    try {
        const res = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: user, password: pass })
        });
        const data = await res.json();

        if (data.success) {
            token = data.token;
            username = user;
            localStorage.setItem('chat_token', token);
            localStorage.setItem('chat_username', username);
            enterChat();
        } else {
            document.getElementById('login-error').textContent = data.message;
        }
    } catch (e) {
        document.getElementById('login-error').textContent = 'Gagal terhubung ke server';
    }
}

// Verify token
async function verifyToken() {
    try {
        const res = await fetch('/api/verify', {
            headers: { 'Authorization': token }
        });
        const data = await res.json();

        if (data.success) {
            username = data.username;
            enterChat();
        } else {
            localStorage.removeItem('chat_token');
            localStorage.removeItem('chat_username');
        }
    } catch (e) {
        console.log('Token verification failed');
    }
}

// Logout
async function logout() {
    try {
        await fetch('/api/logout', {
            method: 'POST',
            headers: { 'Authorization': token }
        });
    } catch (e) {}

    localStorage.removeItem('chat_token');
    localStorage.removeItem('chat_username');
    if (ws) ws.close();
    location.reload();
}

// Enter chat
function enterChat() {
    document.getElementById('auth-screen').style.display = 'none';
    document.getElementById('chat-screen').style.display = 'flex';
    document.getElementById('user-info').style.display = 'flex';
    document.getElementById('current-user').textContent = '👤 ' + username;

    connectWebSocket();
}

// WebSocket
function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(protocol + '//' + window.location.host + '/ws');

    ws.onopen = () => {
        ws.send(JSON.stringify({ type: 'join', username: username }));
        document.getElementById('message').focus();
    };

    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'history') {
            displayHistory(data.messages);
        } else {
            displayMessage(data);
        }
    };

    ws.onclose = () => {
        displayMessage({
            type: 'system',
            content: 'Koneksi terputus. Refresh halaman untuk reconnect.'
        });
    };
}

function sendMessage() {
    const input = document.getElementById('message');
    const content = input.value.trim();
    if (!content || !ws) return;

    ws.send(JSON.stringify({
        type: 'message',
        username: username,
        content: content
    }));
    input.value = '';
}

function displayHistory(messages) {
    const container = document.getElementById('messages');
    const chatMessages = messages.filter(m => m.type === 'message');

    if (chatMessages.length === 0) return;

    const divider = document.createElement('div');
    divider.className = 'message system';
    divider.textContent = `── ${chatMessages.length} pesan sebelumnya ──`;
    container.appendChild(divider);

    chatMessages.forEach(msg => displayMessage(msg, true));

    const divider2 = document.createElement('div');
    divider2.className = 'message system';
    divider2.textContent = '── pesan baru ──';
    container.appendChild(divider2);
}

function displayMessage(msg, isHistory = false) {
    const container = document.getElementById('messages');
    const div = document.createElement('div');

    if (msg.type === 'system') {
        div.className = 'message system';
        div.textContent = msg.content;
    } else {
        const isMine = msg.username === username;
        div.className = 'message ' + (isMine ? 'mine' : 'other');
        if (isHistory) div.classList.add('history');
        div.innerHTML = `
            <div class="meta">${isMine ? 'Kamu' : escapeHtml(msg.username)} • ${msg.timestamp}</div>
            <div class="text">${escapeHtml(msg.content)}</div>
        `;
    }

    container.appendChild(div);
    container.scrollTop = container.scrollHeight;
}

function clearChat() {
    if (!confirm('Hapus semua history chat?')) return;

    fetch('/api/clear', { method: 'POST' })
        .then(res => {
            if (res.ok) {
                document.getElementById('messages').innerHTML = '';
                displayMessage({ type: 'system', content: 'Chat history dihapus' });
            }
        })
        .catch(() => alert('Gagal menghapus chat'));
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Event listeners
document.getElementById('login-username').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') document.getElementById('login-password').focus();
});
document.getElementById('login-password').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') login();
});
document.getElementById('reg-username').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') document.getElementById('reg-password').focus();
});
document.getElementById('reg-password').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') document.getElementById('reg-password2').focus();
});
document.getElementById('reg-password2').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') register();
});
document.getElementById('message').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') sendMessage();
});