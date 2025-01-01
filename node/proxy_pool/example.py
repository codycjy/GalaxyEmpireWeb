from proxy_pool.proxy_pool import ProxyPool
import time


def main():
    # Initialize the proxy pool
    pool = ProxyPool()

    # Example of getting and using proxies
    for _ in range(3):
        proxy = pool.get_proxy()
        if proxy:
            print(f"Got proxy: {proxy}")

            # Simulate using the proxy
            try:
                # Here you would typically make a request using the proxy
                # For demonstration, let's just simulate a failed proxy
                if _ == 1:  # Simulate failure for second proxy
                    raise Exception("Proxy failed")

                print(f"Successfully used proxy {proxy}")

            except Exception as e:
                print(f"Proxy {proxy} failed: {e}")
                # Mark the proxy as deleted
                pool.delete_proxy(proxy)
        else:
            print("No valid proxies available")

        time.sleep(1)  # Small delay between requests

    # Show how many proxies are still in the pool
    print(f"\nDeleted proxies: {pool.deleted_proxies}")


if __name__ == "__main__":
    time.sleep(3)
    main()
