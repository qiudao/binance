.PHONY: help download update clean backup list

# 默认目标
help:
	@echo "BitMEX 交易记录下载工具"
	@echo ""
	@echo "使用方法:"
	@echo "  make download    - 全量下载所有历史交易记录"
	@echo "  make update      - 增量更新（只下载新记录）"
	@echo "  make list        - 列出所有CSV文件"
	@echo "  make backup      - 备份最新的CSV文件"
	@echo "  make clean       - 删除所有CSV文件（需确认）"
	@echo ""

# 全量下载
download:
	@echo "开始全量下载..."
	go run bitmex_data.go

# 增量更新
update:
	@echo "开始增量更新..."
	go run bitmex_data.go --update

# 列出所有CSV文件
list:
	@echo "CSV文件列表:"
	@ls -lh bitmex_executions_*.csv 2>/dev/null || echo "没有找到CSV文件"

# 备份最新文件
backup:
	@latest=$$(ls -t bitmex_executions_*.csv 2>/dev/null | head -1); \
	if [ -n "$$latest" ]; then \
		backup_name="$${latest%.csv}_backup_$$(date +%Y%m%d_%H%M%S).csv"; \
		cp "$$latest" "$$backup_name"; \
		echo "已备份: $$latest -> $$backup_name"; \
	else \
		echo "没有找到CSV文件"; \
	fi

# 清理所有CSV文件
clean:
	@echo "警告: 这将删除所有 bitmex_executions_*.csv 文件"
	@read -p "确认删除? [y/N] " confirm; \
	if [ "$$confirm" = "y" ] || [ "$$confirm" = "Y" ]; then \
		rm -f bitmex_executions_*.csv; \
		echo "已删除所有CSV文件"; \
	else \
		echo "已取消"; \
	fi
