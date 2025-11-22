.PHONY: all build clean fetch-1m fetch-5m fetch-15m fetch-1h fetch-4h fetch-1d save-1m save-5m save-15m save-1h save-4h save-1d

all: build

build:
	go build -o bin/binance-kline main.go

# 查看数据（不保存）
fetch-1m:
	go run main.go -interval 1m

fetch-5m:
	go run main.go -interval 5m

fetch-15m:
	go run main.go -interval 15m

fetch-1h:
	go run main.go -interval 1h

fetch-4h:
	go run main.go -interval 4h

fetch-1d:
	go run main.go -interval 1d

# 保存数据到CSV文件
save-1m:
	go run main.go -interval 1m -output data/klines_1m.csv
	#go run main.go -interval 1m -limit 500 -output data/klines_1m.csv

save-5m:
	go run main.go -interval 5m -output data/klines_5m.csv

save-15m:
	go run main.go -interval 15m -output data/klines_15m.csv

save-1h:
	go run main.go -interval 1h -output data/klines_1h.csv

save-4h:
	go run main.go -interval 4h -output data/klines_4h.csv

save-1d:
	go run main.go -interval 1d -output data/klines_1d.csv

clean:
	rm -rf bin/ data/
