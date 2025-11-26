// 时间旅行控制器
class TimeTravelController {
    constructor() {
        this.currentDate = new Date();
        this.minDate = new Date('2020-05-01');
        this.maxDate = new Date();
        this.isPlaying = false;
        this.playInterval = null;
        this.playSpeed = 1000; // 默认1秒/天
        this.isHistoryMode = false;

        this.init();
    }

    init() {
        // 初始化日期选择器
        const datePicker = document.getElementById('date-picker');
        datePicker.value = this.formatDate(this.currentDate);
        datePicker.min = this.formatDate(this.minDate);
        datePicker.max = this.formatDate(this.maxDate);

        // 绑定事件
        this.bindEvents();

        // 初始加载今天的数据
        this.hideHistoryBanner();
    }

    bindEvents() {
        // 日期选择器变化
        document.getElementById('date-picker').addEventListener('change', (e) => {
            const selectedDate = new Date(e.target.value);
            this.goToDate(selectedDate);
        });

        // 前进/后退按钮
        document.getElementById('prev-day-btn').addEventListener('click', () => this.previousDay());
        document.getElementById('next-day-btn').addEventListener('click', () => this.nextDay());
        document.getElementById('prev-month-btn').addEventListener('click', () => this.previousMonth());
        document.getElementById('next-month-btn').addEventListener('click', () => this.nextMonth());

        // 播放/暂停/回到今天
        document.getElementById('play-btn').addEventListener('click', () => this.play());
        document.getElementById('pause-btn').addEventListener('click', () => this.pause());
        document.getElementById('today-btn').addEventListener('click', () => this.goToToday());

        // 速度选择
        document.getElementById('speed-select').addEventListener('change', (e) => {
            this.playSpeed = parseInt(e.target.value);
            if (this.isPlaying) {
                this.pause();
                this.play(); // 重新开始播放以应用新速度
            }
        });

        // 键盘快捷键
        document.addEventListener('keydown', (e) => {
            // 如果在输入框中,不响应快捷键
            if (e.target.tagName === 'INPUT' || e.target.tagName === 'SELECT') {
                return;
            }

            switch(e.key) {
                case 'ArrowLeft':
                    e.preventDefault();
                    this.previousDay();
                    break;
                case 'ArrowRight':
                    e.preventDefault();
                    this.nextDay();
                    break;
                case 'ArrowUp':
                    e.preventDefault();
                    this.previousMonth();
                    break;
                case 'ArrowDown':
                    e.preventDefault();
                    this.nextMonth();
                    break;
                case ' ':
                    e.preventDefault();
                    if (this.isPlaying) {
                        this.pause();
                    } else {
                        this.play();
                    }
                    break;
                case 'Home':
                    e.preventDefault();
                    this.goToDate(this.minDate);
                    break;
                case 'End':
                    e.preventDefault();
                    this.goToToday();
                    break;
            }
        });
    }

    async loadSnapshot(date) {
        try {
            const dateStr = this.formatDate(date);
            console.log('Loading snapshot for:', dateStr);

            const snapshot = await API.getSnapshot(dateStr);

            // 更新UI
            this.updateAccountInfo(snapshot);
            this.updateBTCPositions(snapshot.btcPositions);
            this.updateTodayOrders(snapshot.todayOrders);
            this.updateRecentExecs(snapshot.recentExecs);
            this.updateChart(snapshot.klineData, dateStr);
            this.updateDateDisplay(dateStr);

            // 检查是否是历史模式
            const isToday = this.isSameDay(date, this.maxDate);
            if (!isToday) {
                this.showHistoryBanner(dateStr);
            } else {
                this.hideHistoryBanner();
            }

        } catch (error) {
            console.error('Failed to load snapshot:', error);
            alert('加载历史数据失败: ' + error.message);
        }
    }

    updateAccountInfo(snapshot) {
        // 更新账户信息
        const totalEquity = document.getElementById('total-equity');
        totalEquity.textContent = Format.btc(snapshot.totalEquity);
        totalEquity.className = 'value ' + (snapshot.totalEquity >= 0 ? 'positive' : 'negative');

        document.getElementById('balance').textContent = Format.btc(snapshot.balance);

        const unrealizedPnl = document.getElementById('unrealized-pnl');
        unrealizedPnl.textContent = Format.btc(snapshot.unrealizedPnl);
        unrealizedPnl.className = 'value ' + (snapshot.unrealizedPnl >= 0 ? 'positive' : 'negative');
    }

    updateBTCPositions(positions) {
        const container = document.getElementById('positions-container');

        if (positions.length === 0) {
            container.innerHTML = '<div class="empty-state">No BTC positions on this date</div>';
            return;
        }

        container.innerHTML = positions.map(pos => {
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
    }

    updateTodayOrders(orders) {
        const tbody = document.getElementById('pending-orders-body');
        const count = document.getElementById('pending-count');

        count.textContent = orders.length;

        if (orders.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" class="empty-state">No orders on this date</td></tr>';
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
    }

    updateRecentExecs(executions) {
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
    }

    updateChart(klineData, dateStr) {
        // 使用现有的图表更新函数
        if (typeof chartManager !== 'undefined') {
            chartManager.updateData(klineData, dateStr);
        }
    }

    updateDateDisplay(dateStr) {
        document.getElementById('date-picker').value = dateStr;
        document.getElementById('viewing-date').textContent = dateStr;
    }

    showHistoryBanner(dateStr) {
        const banner = document.getElementById('history-banner');
        banner.style.display = 'block';
        document.getElementById('viewing-date').textContent = dateStr;
        this.isHistoryMode = true;
    }

    hideHistoryBanner() {
        document.getElementById('history-banner').style.display = 'none';
        this.isHistoryMode = false;
    }

    // 导航方法
    previousDay() {
        const prev = new Date(this.currentDate);
        prev.setDate(prev.getDate() - 1);

        if (prev >= this.minDate) {
            this.goToDate(prev);
        }
    }

    nextDay() {
        const next = new Date(this.currentDate);
        next.setDate(next.getDate() + 1);

        if (next <= this.maxDate) {
            this.goToDate(next);
        }
    }

    previousMonth() {
        const prev = new Date(this.currentDate);
        prev.setMonth(prev.getMonth() - 1);

        if (prev < this.minDate) {
            prev.setTime(this.minDate.getTime());
        }

        this.goToDate(prev);
    }

    nextMonth() {
        const next = new Date(this.currentDate);
        next.setMonth(next.getMonth() + 1);

        if (next > this.maxDate) {
            next.setTime(this.maxDate.getTime());
        }

        this.goToDate(next);
    }

    goToDate(date) {
        this.currentDate = new Date(date);
        this.loadSnapshot(this.currentDate);
    }

    goToToday() {
        this.goToDate(this.maxDate);
    }

    // 播放控制
    play() {
        if (this.isPlaying) return;

        this.isPlaying = true;
        document.getElementById('play-btn').style.display = 'none';
        document.getElementById('pause-btn').style.display = 'inline-block';

        this.playInterval = setInterval(() => {
            this.nextDay();

            // 如果到达最新日期,自动暂停
            if (this.isSameDay(this.currentDate, this.maxDate)) {
                this.pause();
            }
        }, this.playSpeed);
    }

    pause() {
        if (!this.isPlaying) return;

        this.isPlaying = false;
        document.getElementById('play-btn').style.display = 'inline-block';
        document.getElementById('pause-btn').style.display = 'none';

        if (this.playInterval) {
            clearInterval(this.playInterval);
            this.playInterval = null;
        }
    }

    // 工具方法
    formatDate(date) {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        return `${year}-${month}-${day}`;
    }

    isSameDay(date1, date2) {
        return date1.getFullYear() === date2.getFullYear() &&
               date1.getMonth() === date2.getMonth() &&
               date1.getDate() === date2.getDate();
    }
}

// 全局实例
let timeTravelController;

// 等待DOM加载完成后初始化
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        timeTravelController = new TimeTravelController();
    });
} else {
    timeTravelController = new TimeTravelController();
}
