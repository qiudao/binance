// 订单和仓位管理

// 加载未成交订单
async function loadPendingOrders() {
    try {
        const orders = await API.getPendingOrders();
        const tbody = document.getElementById('pending-orders-body');
        const count = document.getElementById('pending-count');

        count.textContent = orders.length;

        if (orders.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="empty-state">No pending orders</td></tr>';
            return;
        }

        tbody.innerHTML = orders.map(order => `
            <tr>
                <td>${Format.datetime(order.timestamp)}</td>
                <td>${order.symbol}</td>
                <td class="side-${order.side.toLowerCase()}">${order.side}</td>
                <td>$${Format.price(order.price)}</td>
                <td>${Format.qty(order.qty)}</td>
                <td>${order.orderType}</td>
                <td class="status-${order.status.toLowerCase()}">${order.status}</td>
            </tr>
        `).join('');

    } catch (error) {
        console.error('Failed to load pending orders:', error);
    }
}

// 加载已成交订单
async function loadExecutions() {
    try {
        const executions = await API.getExecutions();
        const tbody = document.getElementById('executions-body');
        const count = document.getElementById('exec-count');

        count.textContent = executions.length;

        if (executions.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" class="empty-state">No executions</td></tr>';
            return;
        }

        tbody.innerHTML = executions.map(exec => `
            <tr>
                <td>${Format.datetime(exec.timestamp)}</td>
                <td>${exec.symbol}</td>
                <td class="side-${exec.side.toLowerCase()}">${exec.side}</td>
                <td>$${Format.price(exec.price)}</td>
                <td>${Format.qty(exec.qty)}</td>
                <td>${Format.btc(exec.commission)}</td>
            </tr>
        `).join('');

    } catch (error) {
        console.error('Failed to load executions:', error);
    }
}

// 加载当前仓位
async function loadPositions() {
    try {
        const positions = await API.getPositions();
        const container = document.getElementById('positions-container');

        // 只显示BTC相关的币种
        const btcPositions = positions.filter(pos =>
            pos.symbol === 'XBTUSD' || pos.symbol.includes('XBT')
        );

        if (btcPositions.length === 0) {
            container.innerHTML = '<div class="empty-state">No BTC positions</div>';
            return;
        }

        container.innerHTML = btcPositions.map(pos => {
            const sideClass = pos.side.toLowerCase();

            return `
                <div class="position-card ${sideClass}">
                    <div class="symbol">${pos.symbol}</div>
                    <div class="side">${pos.side} ${pos.side === 'Long' ? '🟢' : '🔴'}</div>

                    <div class="info">
                        <span class="label">Quantity:</span>
                        <span>${Format.qty(Math.abs(pos.qty))}</span>
                    </div>

                    <div class="info">
                        <span class="label">Entry Price:</span>
                        <span>$${Format.price(pos.entryPrice)}</span>
                    </div>

                    <div class="info">
                        <span class="label">Current Price:</span>
                        <span>$${Format.price(pos.currentPrice)}</span>
                    </div>
                </div>
            `;
        }).join('');

    } catch (error) {
        console.error('Failed to load positions:', error);
    }
}

// 加载账户信息
async function loadAccountInfo() {
    try {
        const account = await API.getAccount();

        // 总市值
        const totalEquity = document.getElementById('total-equity');
        totalEquity.textContent = Format.btc(account.totalEquity);
        totalEquity.className = 'value ' + (account.totalEquity >= 0 ? 'positive' : 'negative');

        // 余额
        document.getElementById('balance').textContent = Format.btc(account.balance);

        // 未实现盈亏
        const unrealizedPnl = document.getElementById('unrealized-pnl');
        unrealizedPnl.textContent = Format.btc(account.unrealizedPnl);
        unrealizedPnl.className = 'value ' + (account.unrealizedPnl >= 0 ? 'positive' : 'negative');

        // 胜率和交易次数
        document.getElementById('win-rate').textContent = account.winRate.toFixed(1) + '%';
        document.getElementById('total-trades').textContent = account.totalTrades;

    } catch (error) {
        console.error('Failed to load account info:', error);
    }
}
