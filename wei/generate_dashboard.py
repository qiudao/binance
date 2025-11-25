#!/usr/bin/env python3
"""
生成 BitMEX 钱包资金分析的 HTML 交互式仪表板

用法: python3 generate_dashboard.py
输出: wallet_dashboard.html
"""

import pandas as pd
import json
from datetime import datetime
import sys
import os

CSV_FILE = "wallet.csv"
OUTPUT_DIR = "report"
OUTPUT_HTML = os.path.join(OUTPUT_DIR, "wallet_dashboard.html")

def load_and_process_data():
    """加载并处理数据"""
    if not os.path.exists(CSV_FILE):
        print(f"❌ 错误: 找不到 {CSV_FILE} 文件")
        sys.exit(1)

    print(f"📥 加载数据: {CSV_FILE}")
    df = pd.read_csv(CSV_FILE)
    df['Timestamp'] = pd.to_datetime(df['Timestamp'])

    # 基础统计
    start_balance = df.iloc[0]['WalletBalance_BTC']
    end_balance = df.iloc[-1]['WalletBalance_BTC']
    growth = end_balance - start_balance
    growth_pct = (growth / start_balance) * 100

    # 充提币统计
    deposits = df[(df['TransactType'] == 'Deposit') & (df['TransactStatus'] == 'Completed')]
    withdrawals = df[(df['TransactType'] == 'Withdrawal') & (df['TransactStatus'] == 'Completed')]
    deposit_sum = deposits['Amount_BTC'].sum()
    withdrawal_sum = withdrawals['Amount_BTC'].sum()
    net_flow = deposit_sum + withdrawal_sum

    # 盈亏统计
    pnl_df = df[df['TransactType'] == 'RealisedPNL']
    total_pnl = pnl_df['Amount_BTC'].sum()
    win_count = (pnl_df['Amount_BTC'] > 0).sum()
    loss_count = (pnl_df['Amount_BTC'] < 0).sum()
    win_rate = (win_count / len(pnl_df) * 100) if len(pnl_df) > 0 else 0

    # 资金费
    funding_df = df[df['TransactType'] == 'Funding']
    funding_sum = funding_df['Amount_BTC'].sum()

    # 峰值谷底
    max_balance = df['WalletBalance_BTC'].max()
    min_balance = df.iloc[100:]['WalletBalance_BTC'].min()
    max_time = df.loc[df['WalletBalance_BTC'].idxmax(), 'Timestamp'].strftime('%Y-%m-%d')
    min_time = df.iloc[100:].loc[df.iloc[100:]['WalletBalance_BTC'].idxmin(), 'Timestamp'].strftime('%Y-%m-%d')

    drawdown = ((max_balance - min_balance) / max_balance) * 100

    # 准备时间序列数据
    balance_data = df[['Timestamp', 'WalletBalance_BTC']].copy()
    balance_data['Timestamp'] = balance_data['Timestamp'].dt.strftime('%Y-%m-%d %H:%M:%S')

    # 累计盈亏数据
    pnl_cumsum = pnl_df[['Timestamp', 'Amount_BTC']].copy()
    pnl_cumsum['Cumulative_PNL'] = pnl_cumsum['Amount_BTC'].cumsum()
    pnl_cumsum['Timestamp'] = pnl_cumsum['Timestamp'].dt.strftime('%Y-%m-%d %H:%M:%S')

    # 月度盈亏
    pnl_monthly = pnl_df.copy()
    pnl_monthly['YearMonth'] = pnl_monthly['Timestamp'].dt.to_period('M').astype(str)
    monthly_pnl = pnl_monthly.groupby('YearMonth')['Amount_BTC'].sum().reset_index()

    # 充提币事件
    deposit_events = deposits[['Timestamp', 'Amount_BTC', 'WalletBalance_BTC']].copy()
    deposit_events['Timestamp'] = deposit_events['Timestamp'].dt.strftime('%Y-%m-%d')

    withdrawal_events = withdrawals[['Timestamp', 'Amount_BTC', 'WalletBalance_BTC']].copy()
    withdrawal_events['Timestamp'] = withdrawal_events['Timestamp'].dt.strftime('%Y-%m-%d')

    stats = {
        'start_balance': round(start_balance, 8),
        'end_balance': round(end_balance, 8),
        'growth': round(growth, 8),
        'growth_pct': round(growth_pct, 2),
        'deposit_sum': round(deposit_sum, 8),
        'withdrawal_sum': round(withdrawal_sum, 8),
        'net_flow': round(net_flow, 8),
        'total_pnl': round(total_pnl, 8),
        'win_count': int(win_count),
        'loss_count': int(loss_count),
        'win_rate': round(win_rate, 2),
        'funding_sum': round(funding_sum, 8),
        'max_balance': round(max_balance, 8),
        'min_balance': round(min_balance, 8),
        'max_time': max_time,
        'min_time': min_time,
        'drawdown': round(drawdown, 2),
        'total_records': len(df),
        'pnl_trades': len(pnl_df),
        'funding_payments': len(funding_df)
    }

    chart_data = {
        'balance': balance_data.to_dict('records'),
        'pnl_cumsum': pnl_cumsum[['Timestamp', 'Cumulative_PNL']].to_dict('records'),
        'monthly_pnl': monthly_pnl.to_dict('records'),
        'deposits': deposit_events.to_dict('records'),
        'withdrawals': withdrawal_events.to_dict('records')
    }

    return stats, chart_data

def generate_html(stats, chart_data):
    """生成 HTML 内容"""
    html_content = f'''<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>BitMEX Wallet Dashboard</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        * {{
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }}

        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
            color: #333;
        }}

        .container {{
            max-width: 1400px;
            margin: 0 auto;
        }}

        h1 {{
            text-align: center;
            color: white;
            font-size: 2.5em;
            margin-bottom: 30px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }}

        .metrics-grid {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }}

        .metric-card {{
            background: white;
            padding: 25px;
            border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            transition: transform 0.3s;
        }}

        .metric-card:hover {{
            transform: translateY(-5px);
        }}

        .metric-title {{
            font-size: 0.9em;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
        }}

        .metric-value {{
            font-size: 2em;
            font-weight: bold;
            color: #2E86DE;
            margin-bottom: 5px;
        }}

        .metric-subtitle {{
            font-size: 0.9em;
            color: #999;
        }}

        .positive {{
            color: #00B894;
        }}

        .negative {{
            color: #D63031;
        }}

        .chart-container {{
            background: white;
            padding: 30px;
            border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            margin-bottom: 30px;
        }}

        .chart-title {{
            font-size: 1.5em;
            font-weight: bold;
            margin-bottom: 20px;
            color: #333;
        }}

        canvas {{
            max-height: 400px;
        }}

        footer {{
            text-align: center;
            color: white;
            margin-top: 50px;
            padding: 20px;
            font-size: 0.9em;
        }}

        .timestamp {{
            color: rgba(255,255,255,0.8);
            font-size: 0.85em;
        }}
    </style>
</head>
<body>
    <div class="container">
        <h1>📊 BitMEX Wallet Dashboard</h1>

        <!-- 关键指标卡片 -->
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-title">Current Balance</div>
                <div class="metric-value">{stats['end_balance']} BTC</div>
                <div class="metric-subtitle">From {stats['start_balance']} BTC</div>
            </div>

            <div class="metric-card">
                <div class="metric-title">Total Growth</div>
                <div class="metric-value positive">+{stats['growth']} BTC</div>
                <div class="metric-subtitle">{stats['growth_pct']}% Increase</div>
            </div>

            <div class="metric-card">
                <div class="metric-title">Realized PNL</div>
                <div class="metric-value {'positive' if stats['total_pnl'] > 0 else 'negative'}">{stats['total_pnl']:+.4f} BTC</div>
                <div class="metric-subtitle">{stats['pnl_trades']} Trades</div>
            </div>

            <div class="metric-card">
                <div class="metric-title">Win Rate</div>
                <div class="metric-value">{stats['win_rate']}%</div>
                <div class="metric-subtitle">{stats['win_count']} Wins / {stats['loss_count']} Losses</div>
            </div>

            <div class="metric-card">
                <div class="metric-title">Net Deposits</div>
                <div class="metric-value {'positive' if stats['net_flow'] > 0 else 'negative'}">{stats['net_flow']:+.4f} BTC</div>
                <div class="metric-subtitle">In: {stats['deposit_sum']:.4f} | Out: {stats['withdrawal_sum']:.4f}</div>
            </div>

            <div class="metric-card">
                <div class="metric-title">Max Drawdown</div>
                <div class="metric-value negative">{stats['drawdown']:.2f}%</div>
                <div class="metric-subtitle">Peak: {stats['max_balance']:.4f} BTC</div>
            </div>

            <div class="metric-card">
                <div class="metric-title">Funding Income</div>
                <div class="metric-value {'positive' if stats['funding_sum'] > 0 else 'negative'}">{stats['funding_sum']:+.4f} BTC</div>
                <div class="metric-subtitle">{stats['funding_payments']} Payments</div>
            </div>

            <div class="metric-card">
                <div class="metric-title">Total Records</div>
                <div class="metric-value">{stats['total_records']}</div>
                <div class="metric-subtitle">All Transactions</div>
            </div>
        </div>

        <!-- 图表 1: 账户余额趋势 -->
        <div class="chart-container">
            <div class="chart-title">Account Balance Trend</div>
            <canvas id="balanceChart"></canvas>
        </div>

        <!-- 图表 2: 累计盈亏 -->
        <div class="chart-container">
            <div class="chart-title">Cumulative PNL</div>
            <canvas id="pnlChart"></canvas>
        </div>

        <!-- 图表 3: 月度盈亏 -->
        <div class="chart-container">
            <div class="chart-title">Monthly PNL</div>
            <canvas id="monthlyChart"></canvas>
        </div>

        <footer>
            <div class="timestamp">Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</div>
            <div>BitMEX Wallet Analysis Dashboard</div>
        </footer>
    </div>

    <script>
        // 数据
        const stats = {json.dumps(stats, indent=2)};
        const chartData = {json.dumps(chart_data, indent=2)};

        // 图表配置
        Chart.defaults.font.family = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';

        // 1. 余额趋势图
        const balanceCtx = document.getElementById('balanceChart').getContext('2d');
        const balanceChart = new Chart(balanceCtx, {{
            type: 'line',
            data: {{
                labels: chartData.balance.map(d => d.Timestamp),
                datasets: [{{
                    label: 'Wallet Balance (BTC)',
                    data: chartData.balance.map(d => d.WalletBalance_BTC),
                    borderColor: '#2E86DE',
                    backgroundColor: 'rgba(46, 134, 222, 0.1)',
                    borderWidth: 2,
                    fill: true,
                    pointRadius: 0,
                    tension: 0.1
                }}]
            }},
            options: {{
                responsive: true,
                maintainAspectRatio: true,
                plugins: {{
                    legend: {{
                        display: true,
                        position: 'top'
                    }},
                    tooltip: {{
                        mode: 'index',
                        intersect: false,
                        callbacks: {{
                            label: function(context) {{
                                return 'Balance: ' + context.parsed.y.toFixed(4) + ' BTC';
                            }}
                        }}
                    }}
                }},
                scales: {{
                    x: {{
                        display: true,
                        ticks: {{
                            maxRotation: 45,
                            minRotation: 45,
                            maxTicksLimit: 20
                        }}
                    }},
                    y: {{
                        display: true,
                        title: {{
                            display: true,
                            text: 'Balance (BTC)'
                        }}
                    }}
                }}
            }}
        }});

        // 2. 累计PNL图
        const pnlCtx = document.getElementById('pnlChart').getContext('2d');
        const pnlChart = new Chart(pnlCtx, {{
            type: 'line',
            data: {{
                labels: chartData.pnl_cumsum.map(d => d.Timestamp),
                datasets: [{{
                    label: 'Cumulative PNL (BTC)',
                    data: chartData.pnl_cumsum.map(d => d.Cumulative_PNL),
                    borderColor: '#00B894',
                    backgroundColor: 'rgba(0, 184, 148, 0.1)',
                    borderWidth: 2,
                    fill: true,
                    pointRadius: 0,
                    tension: 0.1
                }}]
            }},
            options: {{
                responsive: true,
                maintainAspectRatio: true,
                plugins: {{
                    legend: {{
                        display: true,
                        position: 'top'
                    }},
                    tooltip: {{
                        mode: 'index',
                        intersect: false,
                        callbacks: {{
                            label: function(context) {{
                                return 'PNL: ' + context.parsed.y.toFixed(4) + ' BTC';
                            }}
                        }}
                    }}
                }},
                scales: {{
                    x: {{
                        display: true,
                        ticks: {{
                            maxRotation: 45,
                            minRotation: 45,
                            maxTicksLimit: 20
                        }}
                    }},
                    y: {{
                        display: true,
                        title: {{
                            display: true,
                            text: 'PNL (BTC)'
                        }}
                    }}
                }}
            }}
        }});

        // 3. 月度PNL柱状图
        const monthlyCtx = document.getElementById('monthlyChart').getContext('2d');
        const monthlyPnlData = chartData.monthly_pnl.map(d => d.Amount_BTC);
        const monthlyColors = monthlyPnlData.map(v => v >= 0 ? 'rgba(0, 184, 148, 0.7)' : 'rgba(214, 48, 49, 0.7)');

        const monthlyChart = new Chart(monthlyCtx, {{
            type: 'bar',
            data: {{
                labels: chartData.monthly_pnl.map(d => d.YearMonth),
                datasets: [{{
                    label: 'Monthly PNL (BTC)',
                    data: monthlyPnlData,
                    backgroundColor: monthlyColors,
                    borderColor: monthlyColors.map(c => c.replace('0.7', '1')),
                    borderWidth: 1
                }}]
            }},
            options: {{
                responsive: true,
                maintainAspectRatio: true,
                plugins: {{
                    legend: {{
                        display: true,
                        position: 'top'
                    }},
                    tooltip: {{
                        callbacks: {{
                            label: function(context) {{
                                return 'PNL: ' + context.parsed.y.toFixed(4) + ' BTC';
                            }}
                        }}
                    }}
                }},
                scales: {{
                    x: {{
                        display: true,
                        ticks: {{
                            maxRotation: 45,
                            minRotation: 45
                        }}
                    }},
                    y: {{
                        display: true,
                        title: {{
                            display: true,
                            text: 'PNL (BTC)'
                        }}
                    }}
                }}
            }}
        }});
    </script>
</body>
</html>'''

    return html_content

def main():
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("  📊 生成 BitMEX 钱包仪表板")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print()

    # 加载并处理数据
    stats, chart_data = load_and_process_data()
    print(f"✓ 已处理 {stats['total_records']} 条记录")
    print()

    # 生成 HTML
    print("📝 生成 HTML...")
    html_content = generate_html(stats, chart_data)

    # 写入文件
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    with open(OUTPUT_HTML, 'w', encoding='utf-8') as f:
        f.write(html_content)

    file_size = os.path.getsize(OUTPUT_HTML) / 1024
    print(f"✓ 已生成: {OUTPUT_HTML} ({file_size:.1f} KB)")
    print()
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print(f"  在浏览器中打开: {OUTPUT_HTML}")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n❌ 用户中断")
        sys.exit(0)
    except Exception as e:
        print(f"\n❌ 错误: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
