.PHONY: help download-trades sync-trades ls-trades clean-trades download-wallet sync-wallet ls-wallet clean-wallet download-orders sync-orders ls-orders clean-orders analyze plot dashboard

# 默认显示帮助
.DEFAULT_GOAL := help

help:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  BitMEX 数据下载与分析工具"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📊 交易记录 (Trades)"
	@echo "  download-trades    全量下载交易记录"
	@echo "  sync-trades        增量同步交易记录"
	@echo "  ls-trades          列出所有文件"
	@echo "  clean-trades       清理所有文件（需确认）"
	@echo ""
	@echo "💰 钱包历史 (Wallet)"
	@echo "  download-wallet    全量下载钱包历史"
	@echo "  sync-wallet        增量同步钱包历史"
	@echo "  ls-wallet          列出所有文件"
	@echo "  clean-wallet       清理所有文件（需确认）"
	@echo ""
	@echo "📋 订单记录 (Orders)"
	@echo "  download-orders    全量下载订单记录"
	@echo "  sync-orders        增量同步订单记录"
	@echo "  ls-orders          列出所有文件"
	@echo "  clean-orders       清理所有文件（需确认）"
	@echo ""
	@echo "📈 数据分析 (Analysis)"
	@echo "  analyze            生成文本分析报告"
	@echo "  plot               生成Python可视化图表（需安装pandas）"
	@echo "  dashboard          生成HTML交互式仪表板（需安装pandas）"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ============================================================
# 交易记录 (Trades)
# ============================================================

download-trades:
	@echo "📥 全量下载交易记录..."
	@go run bitmex_data.go

sync-trades:
	@echo "🔄 增量同步交易记录..."
	@go run bitmex_data.go --update

ls-trades:
	@echo "📄 交易记录文件:"
	@ls -lh executions.csv executions.csv.bak 2>/dev/null || echo "  (无文件)"

clean-trades:
	@echo "⚠️  警告: 将删除 executions.csv 和 executions.csv.bak"
	@read -p "确认删除? [y/N] " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		rm -f executions.csv executions.csv.bak; \
		echo "✓ 已清理"; \
	else \
		echo "✗ 已取消"; \
	fi

# ============================================================
# 钱包历史 (Wallet)
# ============================================================

download-wallet:
	@echo "📥 全量下载钱包历史..."
	@go run bitmex_wallet.go

sync-wallet:
	@echo "🔄 增量同步钱包历史..."
	@go run bitmex_wallet.go --update

ls-wallet:
	@echo "📄 钱包历史文件:"
	@ls -lh wallet.csv wallet.csv.bak 2>/dev/null || echo "  (无文件)"

clean-wallet:
	@echo "⚠️  警告: 将删除 wallet.csv 和 wallet.csv.bak"
	@read -p "确认删除? [y/N] " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		rm -f wallet.csv wallet.csv.bak; \
		echo "✓ 已清理"; \
	else \
		echo "✗ 已取消"; \
	fi

# ============================================================
# 订单记录 (Orders)
# ============================================================

download-orders:
	@echo "📥 全量下载订单记录..."
	@go run bitmex_orders.go

sync-orders:
	@echo "🔄 增量同步订单记录..."
	@go run bitmex_orders.go --update

ls-orders:
	@echo "📄 订单记录文件:"
	@ls -lh orders.csv orders.csv.bak 2>/dev/null || echo "  (无文件)"

clean-orders:
	@echo "⚠️  警告: 将删除 orders.csv 和 orders.csv.bak"
	@read -p "确认删除? [y/N] " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		rm -f orders.csv orders.csv.bak; \
		echo "✓ 已清理"; \
	else \
		echo "✗ 已取消"; \
	fi

# ============================================================
# 数据分析 (Analysis)
# ============================================================

analyze:
	@echo "📊 生成钱包资金分析报告..."
	@if [ ! -f wallet.csv ]; then \
		echo "❌ 错误: 找不到 wallet.csv"; \
		echo "请先运行: make download-wallet"; \
		exit 1; \
	fi
	@mkdir -p report
	@./analyze_wallet.sh | tee report/wallet_report.txt
	@echo ""
	@echo "✓ 报告已保存: report/wallet_report.txt"

plot:
	@echo "📈 生成Python可视化图表..."
	@if [ ! -f wallet.csv ]; then \
		echo "❌ 错误: 找不到 wallet.csv"; \
		echo "请先运行: make download-wallet"; \
		exit 1; \
	fi
	@if ! python3 -c "import pandas, matplotlib" 2>/dev/null; then \
		echo ""; \
		echo "❌ 缺少依赖: pandas 或 matplotlib"; \
		echo ""; \
		echo "安装方法:"; \
		echo "  sudo apt install python3-pandas python3-matplotlib"; \
		echo "  或"; \
		echo "  python3 -m pip install --user pandas matplotlib"; \
		echo ""; \
		exit 1; \
	fi
	@mkdir -p report
	@./analyze_wallet.py --save
	@echo ""
	@echo "✓ 图表已保存到 report/ 目录:"
	@ls -lh report/*.png 2>/dev/null || true

dashboard:
	@echo "📊 生成HTML交互式仪表板..."
	@if [ ! -f wallet.csv ]; then \
		echo "❌ 错误: 找不到 wallet.csv"; \
		echo "请先运行: make download-wallet"; \
		exit 1; \
	fi
	@if ! python3 -c "import pandas" 2>/dev/null; then \
		echo ""; \
		echo "❌ 缺少依赖: pandas"; \
		echo ""; \
		echo "安装方法:"; \
		echo "  sudo apt install python3-pandas"; \
		echo "  或"; \
		echo "  python3 -m pip install --user pandas"; \
		echo ""; \
		exit 1; \
	fi
	@mkdir -p report
	@python3 generate_dashboard.py
	@echo ""
	@echo "✓ 已生成: report/wallet_dashboard.html"
	@echo "  在浏览器中打开查看"
