// 诊断工具

let diagnosticResults = {};

// 运行完整诊断
async function runDiagnostics() {
    const panel = document.getElementById('diagnose-panel');
    const content = document.getElementById('diagnose-content');

    panel.style.display = 'flex';
    content.innerHTML = '<p style="text-align: center; padding: 40px;">🔍 Running diagnostics...</p>';

    diagnosticResults = {
        timestamp: new Date().toISOString(),
        browser: getBrowserInfo(),
        checks: {}
    };

    let html = '';

    // 1. 浏览器环境检查
    html += generateSection('1️⃣ Browser Environment', await checkBrowserEnvironment());

    // 2. API 连接检查
    html += generateSection('2️⃣ API Connectivity', await checkAPIConnectivity());

    // 3. 数据检查
    html += generateSection('3️⃣ Data Availability', await checkDataAvailability());

    // 4. 图表库检查
    html += generateSection('4️⃣ TradingView Library', checkTradingViewLibrary());

    // 5. Console 日志
    html += generateSection('5️⃣ Console Logs', getConsoleLogs());

    // 6. 建议
    html += generateRecommendations();

    // 添加复制按钮
    html += '<button class="copy-btn" onclick="copyDiagnostics()">📋 Copy Full Report</button>';

    content.innerHTML = html;
}

// 获取浏览器信息
function getBrowserInfo() {
    return {
        userAgent: navigator.userAgent,
        platform: navigator.platform,
        language: navigator.language,
        onLine: navigator.onLine,
        cookieEnabled: navigator.cookieEnabled,
        screenResolution: `${window.screen.width}x${window.screen.height}`,
        windowSize: `${window.innerWidth}x${window.innerHeight}`
    };
}

// 检查浏览器环境
async function checkBrowserEnvironment() {
    const checks = {};

    checks['Browser'] = {
        value: navigator.userAgent.match(/(Chrome|Firefox|Safari|Edge)/)?.[1] || 'Unknown',
        status: 'info'
    };

    checks['Online Status'] = {
        value: navigator.onLine ? 'Online ✓' : 'Offline ✗',
        status: navigator.onLine ? 'success' : 'error'
    };

    checks['Cookies Enabled'] = {
        value: navigator.cookieEnabled ? 'Yes ✓' : 'No ✗',
        status: navigator.cookieEnabled ? 'success' : 'warning'
    };

    checks['Screen Resolution'] = {
        value: `${window.screen.width}x${window.screen.height}`,
        status: 'info'
    };

    checks['Window Size'] = {
        value: `${window.innerWidth}x${window.innerHeight}`,
        status: 'info'
    };

    diagnosticResults.checks.browser = checks;
    return checks;
}

// 检查 API 连接
async function checkAPIConnectivity() {
    const checks = {};

    // 测试 K线 API
    try {
        const start = Date.now();
        const response = await fetch('/api/klines?symbol=XBTUSD&timeframe=1d');
        const elapsed = Date.now() - start;
        const data = await response.json();

        checks['Klines API'] = {
            value: `✓ ${data.length} records (${elapsed}ms)`,
            status: 'success',
            details: `Status: ${response.status}, Length: ${data.length}`
        };
    } catch (error) {
        checks['Klines API'] = {
            value: `✗ Failed: ${error.message}`,
            status: 'error',
            details: error.stack
        };
    }

    // 测试订单 API
    try {
        const response = await fetch('/api/orders/pending');
        const data = await response.json();
        checks['Orders API'] = {
            value: `✓ ${data.length} pending orders`,
            status: 'success'
        };
    } catch (error) {
        checks['Orders API'] = {
            value: `✗ ${error.message}`,
            status: 'error'
        };
    }

    // 测试成交 API
    try {
        const response = await fetch('/api/executions');
        const data = await response.json();
        checks['Executions API'] = {
            value: `✓ ${data.length} executions`,
            status: 'success'
        };
    } catch (error) {
        checks['Executions API'] = {
            value: `✗ ${error.message}`,
            status: 'error'
        };
    }

    // 测试账户 API
    try {
        const response = await fetch('/api/account');
        const data = await response.json();
        checks['Account API'] = {
            value: `✓ Balance: ${data.balance.toFixed(4)} BTC`,
            status: 'success'
        };
    } catch (error) {
        checks['Account API'] = {
            value: `✗ ${error.message}`,
            status: 'error'
        };
    }

    diagnosticResults.checks.api = checks;
    return checks;
}

// 检查数据可用性
async function checkDataAvailability() {
    const checks = {};

    try {
        const klines = await API.getKlines('XBTUSD', '1d');
        checks['K-line Data'] = {
            value: klines && klines.length > 0 ?
                `✓ ${klines.length} candles available` :
                '✗ No data',
            status: klines && klines.length > 0 ? 'success' : 'error',
            details: klines ? `First: ${new Date(klines[0].time * 1000).toLocaleDateString()}, Last: ${new Date(klines[klines.length-1].time * 1000).toLocaleDateString()}` : 'No data'
        };
    } catch (error) {
        checks['K-line Data'] = {
            value: `✗ ${error.message}`,
            status: 'error'
        };
    }

    diagnosticResults.checks.data = checks;
    return checks;
}

// 检查 TradingView 库
function checkTradingViewLibrary() {
    const checks = {};

    checks['Library Loaded'] = {
        value: typeof LightweightCharts !== 'undefined' ? '✓ Yes' : '✗ No',
        status: typeof LightweightCharts !== 'undefined' ? 'success' : 'error'
    };

    if (typeof LightweightCharts !== 'undefined') {
        checks['Version'] = {
            value: LightweightCharts.version || 'Unknown',
            status: 'info'
        };

        checks['Chart Instance'] = {
            value: chart ? '✓ Created' : '✗ Not created',
            status: chart ? 'success' : 'warning'
        };

        if (chart && candlestickSeries) {
            checks['Candlestick Series'] = {
                value: '✓ Created',
                status: 'success'
            };
        }
    }

    diagnosticResults.checks.tradingview = checks;
    return checks;
}

// 获取 Console 日志
function getConsoleLogs() {
    // 获取最近的 console 输出
    const logs = window.diagnosticLogs || [];
    return {
        'Recent Logs': {
            value: logs.length > 0 ? `${logs.length} entries` : 'No logs captured',
            status: 'info',
            details: logs.slice(-20).join('\n')
        }
    };
}

// 生成诊断部分 HTML
function generateSection(title, checks) {
    let html = `<div class="diagnose-section">`;
    html += `<h3>${title}</h3>`;

    for (const [label, info] of Object.entries(checks)) {
        const statusClass = info.status || 'info';
        html += `<div class="diagnose-item">`;
        html += `<span class="diagnose-label">${label}:</span>`;
        html += `<span class="diagnose-value ${statusClass}">${info.value}</span>`;
        html += `</div>`;

        if (info.details) {
            html += `<div class="diagnose-logs">${info.details}</div>`;
        }
    }

    html += `</div>`;
    return html;
}

// 生成建议
function generateRecommendations() {
    const issues = [];
    const recommendations = [];

    // 检查是否有错误
    for (const [section, checks] of Object.entries(diagnosticResults.checks)) {
        for (const [key, info] of Object.entries(checks)) {
            if (info.status === 'error') {
                issues.push(`${section}: ${key} - ${info.value}`);
            }
        }
    }

    if (issues.length === 0) {
        recommendations.push('✅ All systems operational!');
    } else {
        recommendations.push('⚠️ Issues detected:');
        recommendations.push(...issues);
        recommendations.push('');
        recommendations.push('💡 Suggestions:');

        if (typeof LightweightCharts === 'undefined') {
            recommendations.push('- Refresh the page to reload TradingView library');
        }

        if (!navigator.onLine) {
            recommendations.push('- Check your internet connection');
        }

        recommendations.push('- Check browser Console (F12) for detailed errors');
        recommendations.push('- Try refreshing the page');
        recommendations.push('- Clear browser cache and reload');
    }

    let html = `<div class="diagnose-section">`;
    html += `<h3>6️⃣ Recommendations</h3>`;
    html += `<div class="diagnose-logs">${recommendations.join('\n')}</div>`;
    html += `</div>`;

    return html;
}

// 复制诊断报告
function copyDiagnostics() {
    const report = generateTextReport();
    navigator.clipboard.writeText(report).then(() => {
        const btn = event.target;
        btn.textContent = '✓ Copied!';
        btn.style.background = '#3fb950';
        setTimeout(() => {
            btn.textContent = '📋 Copy Full Report';
            btn.style.background = '#238636';
        }, 2000);
    });
}

// 生成文本报告
function generateTextReport() {
    let report = '═══════════════════════════════════════\n';
    report += 'BitMEX Trading Dashboard - Diagnostic Report\n';
    report += '═══════════════════════════════════════\n\n';
    report += `Timestamp: ${diagnosticResults.timestamp}\n`;
    report += `Browser: ${diagnosticResults.browser.userAgent}\n\n`;

    for (const [section, checks] of Object.entries(diagnosticResults.checks)) {
        report += `\n${section.toUpperCase()}\n`;
        report += '─'.repeat(40) + '\n';
        for (const [key, info] of Object.entries(checks)) {
            report += `${key}: ${info.value}\n`;
            if (info.details) {
                report += `  Details: ${info.details}\n`;
            }
        }
    }

    return report;
}

// 捕获 console 日志
(function() {
    window.diagnosticLogs = [];
    const originalLog = console.log;
    const originalError = console.error;

    console.log = function(...args) {
        window.diagnosticLogs.push('[LOG] ' + args.join(' '));
        originalLog.apply(console, args);
    };

    console.error = function(...args) {
        window.diagnosticLogs.push('[ERROR] ' + args.join(' '));
        originalError.apply(console, args);
    };
})();
