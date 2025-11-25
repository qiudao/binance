// 应用主逻辑

// 初始化应用
async function initApp() {
    console.log('=== BitMEX Trading Dashboard 初始化 ===');
    console.log('1. 检查 TradingView 库...');

    if (typeof LightweightCharts === 'undefined') {
        console.error('❌ TradingView Lightweight Charts 未加载!');
        alert('TradingView 图表库加载失败，请刷新页面重试');
        return;
    }
    console.log('✓ TradingView 库已加载');

    console.log('2. 初始化图表...');
    try {
        initChart();
        console.log('✓ 图表初始化成功');
    } catch (error) {
        console.error('❌ 图表初始化失败:', error);
        return;
    }

    console.log('3. 加载数据...');
    try {
        await loadAllData();
        console.log('✓ 数据加载完成');
    } catch (error) {
        console.error('❌ 数据加载失败:', error);
    }

    console.log('4. 设置事件监听...');
    setupEventListeners();
    console.log('✓ 事件监听已设置');

    console.log('=== Dashboard 初始化完成! ===');
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

    // 诊断按钮
    document.getElementById('diagnose-btn').addEventListener('click', () => {
        runDiagnostics();
    });

    // 关闭诊断面板
    document.getElementById('close-diagnose').addEventListener('click', () => {
        document.getElementById('diagnose-panel').style.display = 'none';
    });

    // 点击面板外部关闭
    document.getElementById('diagnose-panel').addEventListener('click', (e) => {
        if (e.target.id === 'diagnose-panel') {
            document.getElementById('diagnose-panel').style.display = 'none';
        }
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
