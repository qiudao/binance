// TradingView 图表管理
let chart = null;
let candlestickSeries = null;
let currentSymbol = 'XBTUSD';
let currentTimeframe = '1d';

// 初始化图表
function initChart() {
    const chartContainer = document.getElementById('chart');

    chart = LightweightCharts.createChart(chartContainer, {
        width: chartContainer.clientWidth,
        height: 600,
        layout: {
            background: { color: '#0d1117' },
            textColor: '#c9d1d9',
        },
        grid: {
            vertLines: { color: '#21262d' },
            horzLines: { color: '#21262d' },
        },
        crosshair: {
            mode: LightweightCharts.CrosshairMode.Normal,
        },
        rightPriceScale: {
            borderColor: '#30363d',
        },
        timeScale: {
            borderColor: '#30363d',
            timeVisible: true,
            secondsVisible: false,
        },
    });

    candlestickSeries = chart.addCandlestickSeries({
        upColor: '#3fb950',
        downColor: '#f85149',
        borderDownColor: '#f85149',
        borderUpColor: '#3fb950',
        wickDownColor: '#f85149',
        wickUpColor: '#3fb950',
    });

    // 响应式调整
    window.addEventListener('resize', () => {
        chart.applyOptions({ width: chartContainer.clientWidth });
    });
}

// 加载K线数据
async function loadChartData(symbol, timeframe) {
    try {
        currentSymbol = symbol;
        currentTimeframe = timeframe;

        const klines = await API.getKlines(symbol, timeframe);

        if (!klines || klines.length === 0) {
            console.error('No kline data available');
            return;
        }

        candlestickSeries.setData(klines);

        // 加载订单标记
        await loadOrderMarkers(symbol);

    } catch (error) {
        console.error('Failed to load chart data:', error);
    }
}

// 加载订单标记
async function loadOrderMarkers(symbol) {
    try {
        const orders = await API.getOrders();
        const executions = await API.getExecutions();

        const markers = [];

        // 添加订单标记
        executions.forEach(exec => {
            if (exec.symbol !== symbol) return;

            markers.push({
                time: exec.timestampUnix,  // 使用 timestampUnix 字段
                position: exec.side === 'Buy' ? 'belowBar' : 'aboveBar',
                color: exec.side === 'Buy' ? '#3fb950' : '#f85149',
                shape: exec.side === 'Buy' ? 'arrowUp' : 'arrowDown',
                text: `${exec.side} ${exec.qty}`,
                size: 1,
            });
        });

        candlestickSeries.setMarkers(markers);

    } catch (error) {
        console.error('Failed to load order markers:', error);
    }
}

// 切换交易对
function changeSymbol(symbol) {
    loadChartData(symbol, currentTimeframe);
}

// 切换时间周期
function changeTimeframe(timeframe) {
    loadChartData(currentSymbol, timeframe);
}
