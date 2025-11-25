#!/usr/bin/env python3
"""
BitMEX 钱包数据可视化分析

依赖: pip install pandas matplotlib seaborn

用法:
    python3 analyze_wallet.py              # 生成所有图表并显示
    python3 analyze_wallet.py --save       # 保存为PNG文件
"""

import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
from datetime import datetime
import sys
import os

# 设置中文字体
plt.rcParams['font.sans-serif'] = ['DejaVu Sans', 'Arial Unicode MS', 'SimHei']
plt.rcParams['axes.unicode_minus'] = False

# 配置
CSV_FILE = "wallet.csv"
SAVE_MODE = "--save" in sys.argv or "-s" in sys.argv
OUTPUT_DIR = "report"

def load_data():
    """加载CSV数据"""
    if not os.path.exists(CSV_FILE):
        print(f"❌ 错误: 找不到 {CSV_FILE} 文件")
        print("请先运行: make download-wallet")
        sys.exit(1)

    print(f"📥 加载数据: {CSV_FILE}")
    df = pd.read_csv(CSV_FILE)
    df['Timestamp'] = pd.to_datetime(df['Timestamp'])
    print(f"✓ 已加载 {len(df)} 条记录\n")
    return df

def plot_balance_trend(df):
    """1. 账户余额变化趋势图"""
    fig, ax = plt.subplots(figsize=(14, 7))

    # 绘制余额曲线
    ax.plot(df['Timestamp'], df['WalletBalance_BTC'],
            linewidth=2, color='#2E86DE', label='Wallet Balance')

    # 标注充币事件
    deposits = df[df['TransactType'] == 'Deposit']
    if not deposits.empty:
        ax.scatter(deposits['Timestamp'], deposits['WalletBalance_BTC'],
                  color='green', s=150, marker='^',
                  label='Deposit', zorder=5, alpha=0.7)

    # 标注提币事件
    withdrawals = df[(df['TransactType'] == 'Withdrawal') & (df['TransactStatus'] == 'Completed')]
    if not withdrawals.empty:
        ax.scatter(withdrawals['Timestamp'], withdrawals['WalletBalance_BTC'],
                  color='red', s=150, marker='v',
                  label='Withdrawal', zorder=5, alpha=0.7)

    # 标注峰值和谷底
    max_idx = df['WalletBalance_BTC'].idxmax()
    min_idx = df.iloc[100:]['WalletBalance_BTC'].idxmin()  # 排除初始阶段

    ax.scatter(df.loc[max_idx, 'Timestamp'], df.loc[max_idx, 'WalletBalance_BTC'],
              color='gold', s=300, marker='*', label='Peak', zorder=6, edgecolors='black')
    ax.scatter(df.loc[min_idx, 'Timestamp'], df.loc[min_idx, 'WalletBalance_BTC'],
              color='purple', s=300, marker='*', label='Bottom', zorder=6, edgecolors='black')

    # 格式化
    ax.set_xlabel('Date', fontsize=12, fontweight='bold')
    ax.set_ylabel('Balance (BTC)', fontsize=12, fontweight='bold')
    ax.set_title('BitMEX Wallet Balance Trend', fontsize=16, fontweight='bold', pad=20)
    ax.grid(True, alpha=0.3, linestyle='--')
    ax.legend(loc='best', fontsize=10)
    ax.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m'))
    plt.xticks(rotation=45)

    # 添加统计信息文本
    start_balance = df.iloc[0]['WalletBalance_BTC']
    end_balance = df.iloc[-1]['WalletBalance_BTC']
    growth = end_balance - start_balance
    growth_pct = (growth / start_balance) * 100

    stats_text = f'Start: {start_balance:.4f} BTC\nEnd: {end_balance:.4f} BTC\nGrowth: +{growth:.4f} BTC ({growth_pct:.1f}%)'
    ax.text(0.02, 0.98, stats_text, transform=ax.transAxes,
           fontsize=10, verticalalignment='top',
           bbox=dict(boxstyle='round', facecolor='wheat', alpha=0.5))

    plt.tight_layout()
    return fig

def plot_pnl_cumulative(df):
    """2. 累计盈亏趋势图"""
    pnl_df = df[df['TransactType'] == 'RealisedPNL'].copy()
    pnl_df['Cumulative_PNL'] = pnl_df['Amount_BTC'].cumsum()

    fig, ax = plt.subplots(figsize=(14, 7))

    # 绘制累计盈亏
    ax.plot(pnl_df['Timestamp'], pnl_df['Cumulative_PNL'],
           linewidth=2.5, color='#00B894', label='Cumulative PNL')

    # 填充正负区域
    ax.fill_between(pnl_df['Timestamp'], pnl_df['Cumulative_PNL'], 0,
                    where=(pnl_df['Cumulative_PNL'] >= 0),
                    color='green', alpha=0.2, label='Profit')
    ax.fill_between(pnl_df['Timestamp'], pnl_df['Cumulative_PNL'], 0,
                    where=(pnl_df['Cumulative_PNL'] < 0),
                    color='red', alpha=0.2, label='Loss')

    ax.axhline(y=0, color='black', linestyle='--', linewidth=1, alpha=0.5)

    ax.set_xlabel('Date', fontsize=12, fontweight='bold')
    ax.set_ylabel('Cumulative PNL (BTC)', fontsize=12, fontweight='bold')
    ax.set_title('Cumulative Realized PNL Trend', fontsize=16, fontweight='bold', pad=20)
    ax.grid(True, alpha=0.3, linestyle='--')
    ax.legend(loc='best', fontsize=10)
    ax.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m'))
    plt.xticks(rotation=45)

    # 添加统计
    total_pnl = pnl_df['Amount_BTC'].sum()
    win_count = (pnl_df['Amount_BTC'] > 0).sum()
    loss_count = (pnl_df['Amount_BTC'] < 0).sum()
    win_rate = (win_count / len(pnl_df)) * 100

    stats_text = f'Total PNL: {total_pnl:.4f} BTC\nWin Rate: {win_rate:.1f}%\nWins: {win_count} | Losses: {loss_count}'
    ax.text(0.02, 0.98, stats_text, transform=ax.transAxes,
           fontsize=10, verticalalignment='top',
           bbox=dict(boxstyle='round', facecolor='lightblue', alpha=0.5))

    plt.tight_layout()
    return fig

def plot_monthly_pnl(df):
    """3. 月度盈亏柱状图"""
    pnl_df = df[df['TransactType'] == 'RealisedPNL'].copy()
    pnl_df['YearMonth'] = pnl_df['Timestamp'].dt.to_period('M')
    monthly_pnl = pnl_df.groupby('YearMonth')['Amount_BTC'].sum()

    fig, ax = plt.subplots(figsize=(14, 7))

    # 绘制柱状图
    colors = ['green' if x >= 0 else 'red' for x in monthly_pnl.values]
    monthly_pnl.plot(kind='bar', ax=ax, color=colors, alpha=0.7, edgecolor='black')

    ax.axhline(y=0, color='black', linestyle='-', linewidth=1)
    ax.set_xlabel('Month', fontsize=12, fontweight='bold')
    ax.set_ylabel('PNL (BTC)', fontsize=12, fontweight='bold')
    ax.set_title('Monthly Realized PNL', fontsize=16, fontweight='bold', pad=20)
    ax.grid(True, alpha=0.3, axis='y', linestyle='--')
    plt.xticks(rotation=45, ha='right')

    # 只显示部分标签
    xticks = ax.get_xticks()
    if len(xticks) > 20:
        step = len(xticks) // 20
        ax.set_xticks(xticks[::step])

    plt.tight_layout()
    return fig

def plot_transaction_types(df):
    """4. 交易类型占比饼图"""
    type_counts = df['TransactType'].value_counts()

    # 只显示主要类型
    main_types = type_counts.head(6)
    if len(type_counts) > 6:
        others_count = type_counts.iloc[6:].sum()
        main_types['Others'] = others_count

    fig, ax = plt.subplots(figsize=(10, 8))

    colors = plt.cm.Set3(range(len(main_types)))
    wedges, texts, autotexts = ax.pie(main_types.values,
                                        labels=main_types.index,
                                        autopct='%1.1f%%',
                                        colors=colors,
                                        startangle=90,
                                        textprops={'fontsize': 11})

    # 美化百分比文本
    for autotext in autotexts:
        autotext.set_color('white')
        autotext.set_fontweight('bold')
        autotext.set_fontsize(10)

    ax.set_title('Transaction Type Distribution', fontsize=16, fontweight='bold', pad=20)

    # 添加图例
    ax.legend(wedges, [f'{label}: {count}' for label, count in zip(main_types.index, main_types.values)],
             loc='best', bbox_to_anchor=(1, 0, 0.5, 1))

    plt.tight_layout()
    return fig

def plot_drawdown(df):
    """5. 回撤分析图"""
    df = df.copy()
    df['Peak'] = df['WalletBalance_BTC'].cummax()
    df['Drawdown'] = ((df['WalletBalance_BTC'] - df['Peak']) / df['Peak']) * 100

    fig, (ax1, ax2) = plt.subplots(2, 1, figsize=(14, 10), sharex=True)

    # 上图：余额和峰值
    ax1.plot(df['Timestamp'], df['WalletBalance_BTC'],
            linewidth=2, color='#2E86DE', label='Balance')
    ax1.plot(df['Timestamp'], df['Peak'],
            linewidth=1.5, color='red', linestyle='--', label='Peak', alpha=0.7)
    ax1.fill_between(df['Timestamp'], df['WalletBalance_BTC'], df['Peak'],
                     where=(df['WalletBalance_BTC'] < df['Peak']),
                     color='red', alpha=0.2)

    ax1.set_ylabel('Balance (BTC)', fontsize=12, fontweight='bold')
    ax1.set_title('Wallet Balance & Drawdown Analysis', fontsize=16, fontweight='bold', pad=20)
    ax1.grid(True, alpha=0.3, linestyle='--')
    ax1.legend(loc='best', fontsize=10)

    # 下图：回撤百分比
    ax2.fill_between(df['Timestamp'], df['Drawdown'], 0,
                     where=(df['Drawdown'] < 0),
                     color='red', alpha=0.5, label='Drawdown')
    ax2.plot(df['Timestamp'], df['Drawdown'],
            linewidth=1.5, color='darkred')

    ax2.axhline(y=0, color='black', linestyle='-', linewidth=1)
    ax2.set_xlabel('Date', fontsize=12, fontweight='bold')
    ax2.set_ylabel('Drawdown (%)', fontsize=12, fontweight='bold')
    ax2.grid(True, alpha=0.3, linestyle='--')
    ax2.legend(loc='best', fontsize=10)
    ax2.xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m'))
    plt.xticks(rotation=45)

    # 标注最大回撤
    max_dd_idx = df['Drawdown'].idxmin()
    max_dd = df.loc[max_dd_idx, 'Drawdown']
    max_dd_time = df.loc[max_dd_idx, 'Timestamp']
    ax2.scatter(max_dd_time, max_dd, color='red', s=200, marker='v',
               zorder=5, edgecolors='black', label=f'Max DD: {max_dd:.2f}%')
    ax2.legend(loc='best', fontsize=10)

    plt.tight_layout()
    return fig

def main():
    """主函数"""
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("  📊 BitMEX 钱包数据可视化分析")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print()

    # 加载数据
    df = load_data()

    # 生成图表
    figures = []
    print("📈 生成图表...")

    print("  1. 账户余额变化趋势图")
    figures.append(('balance_trend.png', plot_balance_trend(df)))

    print("  2. 累计盈亏趋势图")
    figures.append(('pnl_cumulative.png', plot_pnl_cumulative(df)))

    print("  3. 月度盈亏柱状图")
    figures.append(('monthly_pnl.png', plot_monthly_pnl(df)))

    print("  4. 交易类型占比饼图")
    figures.append(('transaction_types.png', plot_transaction_types(df)))

    print("  5. 回撤分析图")
    figures.append(('drawdown_analysis.png', plot_drawdown(df)))

    print()

    if SAVE_MODE:
        # 保存模式
        print("💾 保存图表...")
        os.makedirs(OUTPUT_DIR, exist_ok=True)
        for filename, fig in figures:
            filepath = os.path.join(OUTPUT_DIR, filename)
            fig.savefig(filepath, dpi=300, bbox_inches='tight')
            print(f"  ✓ 已保存: {filepath}")
        print()
        print(f"✓ 所有图表已保存到 {OUTPUT_DIR}/ 目录")
    else:
        # 显示模式
        print("📊 显示图表...")
        print("  (关闭窗口查看下一张图)")
        for _, fig in figures:
            plt.show()

    print()
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("  分析完成")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n❌ 用户中断")
        sys.exit(0)
    except Exception as e:
        print(f"\n❌ 错误: {e}")
        sys.exit(1)
