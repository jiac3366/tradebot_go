import mmap
import struct
import time
from typing import Dict, NamedTuple
import os


class TradeData(NamedTuple):
    symbol: str
    price: str
    quantity: str
    event_time: int
    trade_time: int
    trade_id: int
    is_maker: bool


class TradeReader:
    TRADE_SIZE = 256
    SHM_PATH = "/dev/shm/binance_trades"  # 使用 /dev/shm 路径

    def __init__(self, symbols: list):
        self.symbols = {s: i for i, s in enumerate(symbols)}
        self.fd = os.open(self.SHM_PATH, os.O_RDWR)
        self.mm = mmap.mmap(self.fd, 0, mmap.MAP_SHARED)

    def read_trade(self, symbol: str) -> TradeData:
        if symbol not in self.symbols:
            raise ValueError(f"Unknown symbol: {symbol}")

        offset = self.symbols[symbol] * self.TRADE_SIZE

        # 读取数据
        data = self.mm[offset : offset + self.TRADE_SIZE]

        # 解析数据
        symbol = data[:32].decode().strip("\x00")
        price = data[32:64].decode().strip("\x00")
        quantity = data[64:96].decode().strip("\x00")

        event_time = struct.unpack("<Q", data[96:104])[0]
        trade_time = struct.unpack("<Q", data[104:112])[0]
        trade_id = struct.unpack("<Q", data[112:120])[0]
        is_maker = bool(data[120])

        return TradeData(
            symbol=symbol,
            price=price,
            quantity=quantity,
            event_time=event_time,
            trade_time=trade_time,
            trade_id=trade_id,
            is_maker=is_maker,
        )

    def close(self):
        self.mm.close()
        os.close(self.fd)


# 使用示例
if __name__ == "__main__":
    symbols = ["BTCUSDT", "ETHUSDT"]  # 与 Go 端相同的符号列表
    reader = TradeReader(symbols)

    try:
        while True:
            # 读取 BTCUSDT 的最新交易数据
            trade = reader.read_trade("BTCUSDT")
            print(trade)
            time.sleep(0.001)  # 1ms 的轮询间隔
    finally:
        reader.close()
