import mmap
import struct
import time
from typing import Dict, NamedTuple
import os
import numpy as np
from collections import defaultdict

LATENCY = defaultdict(list)
LAST_EVENT_TIME = defaultdict(int)


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
        try:
            self.fd = os.open(self.SHM_PATH, os.O_RDWR)
            self.mm = mmap.mmap(self.fd, 0, mmap.MAP_SHARED)
        except FileNotFoundError:
            raise FileNotFoundError(f"Shared memory file not found: {self.SHM_PATH}")

    def get_latency(self, symbol: str, event_time: int) -> int:
        local_time = int(time.time() * 1000)
        latency = local_time - event_time
        LATENCY[symbol].append(latency)
        return latency

    def read_trade(self, symbol: str):
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

        # Calculate latency
        if symbol in LAST_EVENT_TIME:
            last_event_time = LAST_EVENT_TIME[symbol]
            if event_time != last_event_time:
                self.get_latency(symbol, event_time)

        LAST_EVENT_TIME[symbol] = event_time

        # return TradeData(
        #     symbol=symbol,
        #     price=price,
        #     quantity=quantity,
        #     event_time=event_time,
        #     trade_time=trade_time,
        #     trade_id=trade_id,
        #     is_maker=is_maker,
        # )

    def close(self):
        self.mm.close()
        os.close(self.fd)


# 使用示例
if __name__ == "__main__":
    symbols = [
        "ARKMUSDT",
        "ZECUSDT",
        "MANTAUSDT",
        "ENAUSDT",
        "ARKUSDT",
        "RIFUSDT",
        "BEAMXUSDT",
        "METISUSDT",
        "1000SATSUSDT",
        "AMBUSDT",
        "CHZUSDT",
        "RENUSDT",
        "BANANAUSDT",
        "TAOUSDT",
        "RAREUSDT",
        "SXPUSDT",
        "IDUSDT",
        "LQTYUSDT",
        "RPLUSDT",
        "COMBOUSDT",
        "SEIUSDT",
        "RDNTUSDT",
        "BNXUSDT",
        "NKNUSDT",
        "BNBUSDT",
        "APTUSDT",
        "OXTUSDT",
        "LEVERUSDT",
        "AIUSDT",
        "OMNIUSDT",
        "KDAUSDT",
        "ALPACAUSDT",
        "STRKUSDT",
        "FETUSDT",
        "FIDAUSDT",
        "MKRUSDT",
        "GMTUSDT",
        "VIDTUSDT",
        "UMAUSDT",
        "RONINUSDT",
        "FIOUSDT",
        "BALUSDT",
        "IOUSDT",
        "LDOUSDT",
        "KSMUSDT",
        "TURBOUSDT",
        "GUSDT",
        "POLUSDT",
        "XVSUSDT",
        "SUNUSDT",
        "TIAUSDT",
        "LRCUSDT",
        "1MBABYDOGEUSDT",
        "ZKUSDT",
        "ZENUSDT",
        "HOTUSDT",
        "DARUSDT",
        "AXSUSDT",
        "TRXUSDT",
        "LOKAUSDT",
        "LSKUSDT",
        "GLMUSDT",
        "ETHFIUSDT",
        "ONTUSDT",
        "ONGUSDT",
        "CATIUSDT",
        "REZUSDT",
        "NEIROUSDT",
        "VANRYUSDT",
        "ANKRUSDT",
        "ALPHAUSDT",
    ]  # 与 Go 端相同的符号列表
    try:
        reader = TradeReader(symbols)
        while True:
            for symbol in symbols:
                reader.read_trade(symbol)
                # print(trade)
                # time.sleep(0.001)  # 1ms 的轮询间隔
    except KeyboardInterrupt as e:
        print(e)
        reader.close()

        print(f"Latency: {LATENCY}")
        for symbol, latencies in LATENCY.items():
            if latencies:
                avg_latency = np.mean(latencies)
                print(
                    f"Symbol: {symbol}, Avg: {avg_latency:.2f} ms, Median: {np.median(latencies):.2f} ms, Std: {np.std(latencies):.2f} ms, 95%: {np.percentile(latencies, 95):.2f} ms, 99%: {np.percentile(latencies, 99):.2f} ms, Min: {np.min(latencies):.2f} ms, Max: {np.max(latencies):.2f} ms"
                )
