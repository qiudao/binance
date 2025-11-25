// 应用主逻辑

// 初始化应用
async function initApp() {
    console.log('Initializing BitMEX Trading Dashboard...');

    // 初始化图表
    initChart();

    // 加载初始数据
    await loadAllData();

    // 设置事件监听
    setupEventListeners();

    console.log('Dashboard initialized successfully!');
}

// 加载所有数据
async function loadAllData() {
    const symbol = document.getElementById('symbol-select').value;
    const timeframe = document.getElementById('timeframe-select').value;

    // 显示加载状态
    document.body.classList.add('loading');

    try {
        // 并行加载所有数据
        await Promise.all([
            loadChartData(symbol, timeframe),
            loadPendingOrders(),
            loadExecutions(),
            loadPositions(),
            loadAccountInfo()
        ]);
    } catch (error) {
        console.error('Failed to load data:', error);
        alert('Failed to load data. Please check console for details.');
    } finally {
        document.body.classList.remove('loading');
    }
}

// 设置事件监听
function setupEventListeners() {
    // 交易对切换
    document.getElementById('symbol-select').addEventListener('change', (e) => {
        changeSymbol(e.target.value);
    });

    // 时间周期切换
    document.getElementById('timeframe-select').addEventListener('change', (e) => {
        changeTimeframe(e.target.value);
    });

    // 刷新按钮
    document.getElementById('refresh-btn').addEventListener('click', () => {
        loadAllData();
    });

    // 自动刷新 (每30秒)
    setInterval(() => {
        loadPendingOrders();
        loadExecutions();
        loadPositions();
        loadAccountInfo();
    }, 30000);
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    initApp();
});
