// TradingView 图表管理
let chart = null;
let candlestickSeries = null;
let positionSeries = null;  // 仓位柱形图
let equitySeries = null;    // 总市值柱形图
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

    candlestickSeries = chart.addSeries(LightweightCharts.CandlestickSeries, {
        upColor: '#3fb950',
        downColor: '#f85149',
        borderDownColor: '#f85149',
        borderUpColor: '#3fb950',
        wickDownColor: '#f85149',
        wickUpColor: '#3fb950',
        priceScaleId: 'right',
    });

    // 设置K线图占据上部空间
    chart.priceScale('right').applyOptions({
        scaleMargins: {
            top: 0.05,
            bottom: 0.35,  // 留出35%给下方的两个柱形图
        },
    });

    // 仓位柱形图
    positionSeries = chart.addSeries(LightweightCharts.HistogramSeries, {
        color: '#3fb950',
        priceFormat: {
            type: 'volume',
        },
        priceScaleId: 'position',
    });

    chart.priceScale('position').applyOptions({
        scaleMargins: {
            top: 0.70,
            bottom: 0.18,
        },
    });

    // 总市值柱形图
    equitySeries = chart.addSeries(LightweightCharts.HistogramSeries, {
        color: '#58a6ff',
        priceFormat: {
            type: 'price',
            precision: 4,
            minMove: 0.0001,
        },
        priceScaleId: 'equity',
    });

    chart.priceScale('equity').applyOptions({
        scaleMargins: {
            top: 0.85,
            bottom: 0.02,
        },
    });

    // 响应式调整
    window.addEventListener('resize', () => {
        chart.applyOptions({ width: chartContainer.clientWidth });
    });
}

// 加载K线数据
async function loadChartData(symbol, timeframe) {
    console.log(`加载K线数据: ${symbol} ${timeframe}`);
    try {
        currentSymbol = symbol;
        currentTimeframe = timeframe;

        console.log('  - 请求 API...');
        const klines = await API.getKlines(symbol, timeframe);

        console.log(`  - 收到 ${klines ? klines.length : 0} 条K线数据`);

        if (!klines || klines.length === 0) {
            console.error('  ❌ 没有K线数据!');
            alert(`没有找到 ${symbol} ${timeframe} 的K线数据`);
            return;
        }

        console.log('  - 设置图表数据...');
        candlestickSeries.setData(klines);
        console.log('  ✓ K线图渲染完成');

        // 加载每日仓位数据
        console.log('  - 加载每日仓位数据...');
        await loadDailyPositionData();
        console.log('  ✓ 仓位和市值图渲染完成');

        // 加载订单标记
        console.log('  - 加载订单标记...');
        await loadOrderMarkers(symbol);
        console.log('  ✓ 订单标记完成');

    } catch (error) {
        console.error('  ❌ 加载K线失败:', error);
        alert('加载K线数据失败: ' + error.message);
    }
}

// 加载每日仓位数据
async function loadDailyPositionData() {
    try {
        const dailyPositions = await API.getDailyPosition();

        if (!dailyPositions || dailyPositions.length === 0) {
            console.log('  没有每日仓位数据');
            return;
        }

        // 转换仓位数据 - Long绿色, Short红色, Flat灰色
        const positionData = dailyPositions.map(d => ({
            time: d.time,
            value: Math.abs(d.positionQty),
            color: d.side === 'Long' ? '#3fb950' :
                   d.side === 'Short' ? '#f85149' : '#484f58',
        }));

        // 转换总市值数据 - 蓝色系
        const equityData = dailyPositions.map(d => ({
            time: d.time,
            value: d.totalEquity,
            color: '#58a6ff',
        }));

        positionSeries.setData(positionData);
        equitySeries.setData(equityData);

    } catch (error) {
        console.error('加载每日仓位数据失败:', error);
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
