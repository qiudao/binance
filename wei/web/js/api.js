// API 基础URL
const API_BASE = window.location.origin;

// API调用封装
const API = {
    // 获取K线数据
    async getKlines(symbol = 'XBTUSD', timeframe = '1d') {
        const response = await fetch(`${API_BASE}/api/klines?symbol=${symbol}&timeframe=${timeframe}`);
        return response.json();
    },

    // 获取所有订单
    async getOrders(status = '') {
        const url = status ? `${API_BASE}/api/orders?status=${status}` : `${API_BASE}/api/orders`;
        const response = await fetch(url);
        return response.json();
    },

    // 获取未成交订单
    async getPendingOrders() {
        const response = await fetch(`${API_BASE}/api/orders/pending`);
        return response.json();
    },

    // 获取成交记录
    async getExecutions() {
        const response = await fetch(`${API_BASE}/api/executions`);
        return response.json();
    },

    // 获取当前仓位
    async getPositions() {
        const response = await fetch(`${API_BASE}/api/positions`);
        return response.json();
    },

    // 获取账户信息
    async getAccount() {
        const response = await fetch(`${API_BASE}/api/account`);
        return response.json();
    },

    // 获取历史快照
    async getSnapshot(date) {
        const response = await fetch(`${API_BASE}/api/snapshot?date=${date}`);
        return response.json();
    }
};

// 格式化函数
const Format = {
    // 格式化时间
    datetime(timestamp) {
        if (typeof timestamp === 'string') {
            timestamp = new Date(timestamp).getTime() / 1000;
        }
        const date = new Date(timestamp * 1000);
        return date.toISOString().replace('T', ' ').substr(0, 19);
    },

    // 格式化日期
    date(timestamp) {
        if (typeof timestamp === 'string') {
            timestamp = new Date(timestamp).getTime() / 1000;
        }
        const date = new Date(timestamp * 1000);
        return date.toISOString().substr(0, 10);
    },

    // 格式化价格
    price(value) {
        return value.toFixed(2);
    },

    // 格式化BTC数量
    btc(value) {
        return value.toFixed(8) + ' BTC';
    },

    // 格式化百分比
    percent(value) {
        const sign = value >= 0 ? '+' : '';
        return sign + value.toFixed(2) + '%';
    },

    // 格式化数量
    qty(value) {
        return value.toLocaleString();
    }
};
