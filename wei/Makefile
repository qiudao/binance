.PHONY: help download-trades sync-trades ls-trades clean-trades download-wallet sync-wallet ls-wallet clean-wallet

# 默认显示帮助
.DEFAULT_GOAL := help

help:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "  BitMEX 数据下载工具"
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
	@ls -lh bitmex_executions_*.csv 2>/dev/null || echo "  (无文件)"

clean-trades:
	@echo "⚠️  警告: 将删除所有 bitmex_executions_*.csv"
	@read -p "确认删除? [y/N] " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		rm -f bitmex_executions_*.csv; \
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
	@ls -lh bitmex_wallet_*.csv 2>/dev/null || echo "  (无文件)"

clean-wallet:
	@echo "⚠️  警告: 将删除所有 bitmex_wallet_*.csv"
	@read -p "确认删除? [y/N] " ans; \
	if [ "$$ans" = "y" ] || [ "$$ans" = "Y" ]; then \
		rm -f bitmex_wallet_*.csv; \
		echo "✓ 已清理"; \
	else \
		echo "✗ 已取消"; \
	fi
