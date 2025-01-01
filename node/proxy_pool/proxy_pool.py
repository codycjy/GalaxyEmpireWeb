from datetime import datetime
import requests
from queue import PriorityQueue
from dataclasses import dataclass, field
from typing import Optional
from proxy_pool.config import API_URL


@dataclass(order=True)
class Proxy:
    sort_index: float  # This will store the timestamp
    ip: str = field(compare=False)
    port: int = field(compare=False)
    endtime: datetime = field(compare=False)
    city: str = field(compare=False)
    rosname: str = field(compare=False)

    def __str__(self):
        return f"{self.ip}:{self.port}"


class ProxyPool:
    def __init__(self):
        self.proxies = PriorityQueue()
        self.min_pool_size = 5
        self.deleted_proxies = set()
        self.refresh_pool()

    def refresh_pool(self):
        if self.proxies.qsize() >= self.min_pool_size:
            return

        try:
            response = requests.get(API_URL)  # pyright:ignore
            data = response.json()

            if data['success'] and data['data']:
                for proxy_data in data['data']:
                    endtime = datetime.strptime(proxy_data['endtime'], '%Y/%m/%d %H:%M:%S')
                    timestamp = endtime.timestamp()
                    proxy = Proxy(
                        sort_index=timestamp,  # Use timestamp as sort index
                        ip=proxy_data['ip'],
                        port=proxy_data['port'],
                        endtime=endtime,
                        city=proxy_data['city'],
                        rosname=proxy_data['rosname']
                    )
                    self.proxies.put(proxy)  # Now we can put proxy directly
        except Exception as e:
            print(f"Error refreshing proxy pool: {e}")

    def get_proxy(self) -> Optional[str]:
        self.refresh_pool()

        while not self.proxies.empty():
            proxy = self.proxies.get()
            proxy_str = str(proxy)

            if proxy.endtime <= datetime.now():
                self.deleted_proxies.discard(proxy_str)
                continue

            if proxy_str in self.deleted_proxies:
                continue

            return proxy_str

        return None

    def delete_proxy(self, proxy_str: str):
        """Mark a proxy as deleted without removing from queue"""
        self.deleted_proxies.add(proxy_str)
