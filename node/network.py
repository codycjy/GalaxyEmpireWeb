from dataclasses import dataclass
import logging
import requests
from utils import crypto, md5
import time
import sys
import os
from model.user import Account
from config import serverUrlList, PROXY_BASE_URL, PROXY_AUTH_USER, PROXY_AUTH_PASS

# Configure logging
logger = logging.getLogger(__name__)

headers = {
    'User-Agent': 'android',
    'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'
}


def addArgs(args):  # TODO: move me later
    if not args:
        return ""
    return '&' + '&'.join(f"{key}={value}" for key, value in args.items())


@dataclass
class NetworkResponse:
    status: int
    data: dict
    err_msg: str = ""


class Network:
    def __init__(self, user: Account):
        self.server = user.server
        self.username = user.username
        self.password = user.password
        self.ppy_id = None
        self.ssid = None
        self.session = requests.Session()
        self.ready = True
        self.max_login_retries = 3  # Maximum retry attempts
        self.login_retry_count = 0   # Current retry count
        self.proxy = None

        if os.getenv('PROXY', False):
            self.set_proxy()

    def set_proxy(self):
        """
        Set up the proxy for the session.
        """
        self.ready = False
        HTTP_ENDPOINT = '/get/?type=http'
        proxy_url = f"{PROXY_BASE_URL}{HTTP_ENDPOINT}"
        logger.info("Attempting to set proxy...")
        try:
            response = self.session.get(proxy_url, auth=(PROXY_AUTH_USER, PROXY_AUTH_PASS), timeout=5)
            response.raise_for_status()
            proxy = response.json()

            if "proxy" in proxy:
                self.session.proxies = {
                    "http": proxy["proxy"],
                    "https": proxy["proxy"]
                }
                logger.info(f"Proxy set to {proxy['proxy']}")
                self.ready = True
                self.proxy = proxy
            else:
                logger.error("Invalid proxy response format.")
        except requests.exceptions.RequestException as e:
            logger.error(f"Failed to get proxy: {e}")
        except ValueError as e:
            logger.error(f"Invalid JSON response while setting proxy: {e}")

    def _post(self, url: str, args: dict = {}) -> NetworkResponse:
        """
        Internal method to handle POST requests.

        Args:
            url (str): API endpoint.
            args (dict, optional): Parameters for the POST request.

        Returns:
            NetworkResponse: The response wrapped in NetworkResponse.
        """
        logger.debug(f"Making POST request to: {url}, args: {args}")
        if not self.ready:
            logger.error("Network not ready. Cannot make POST request.")
            return NetworkResponse(status=-1, data={}, err_msg="Network not ready")

        extra_args = {}
        if "login" not in url.lower():
            extra_args = self.getSession()

        try:
            args.update(extra_args)
            full_url = serverUrlList[self.server] + url + addArgs(args)
            logger.debug(f"POST Request URL: {full_url}")

        except KeyError as e:
            err_msg = f"Invalid server configuration: {e}"
            logger.error(err_msg)
            return NetworkResponse(status=-1, data={}, err_msg=err_msg)

        try:
            req = self.session.post(full_url, headers=headers, data=crypto(full_url), timeout=5)
            req.raise_for_status()
            data = req.json()
            logger.debug(f"Response JSON: {data}")

            if data.get('status') != 'error':
                self.login_retry_count = 0  # Reset retry count on successful response
                return NetworkResponse(status=0, data=data)

            if data.get('err_code') == 111:
                if self.login_retry_count >= self.max_login_retries:
                    logger.error("Max login retries exceeded")
                    return NetworkResponse(status=-1, data={}, err_msg="Max login retries exceeded")

                logger.warning("Session expired. Attempting to relogin...")
                self.login_retry_count += 1
                login_response = self.login()
                if login_response.status == 0:
                    return self._post(url, args)
                return login_response

            return NetworkResponse(status=-1, data={}, err_msg=data.get('err_msg', ''))

        except requests.exceptions.Timeout:
            logger.error("Request timed out.")
            return NetworkResponse(status=-1, data={}, err_msg="Request timed out.")
        except requests.exceptions.HTTPError as e:
            logger.error(f"HTTP error occurred: {e}")
            return NetworkResponse(status=-1, data={}, err_msg=str(e))
        except requests.exceptions.RequestException as e:
            logger.error(f"Request exception: {e}")
            return NetworkResponse(status=-1, data={}, err_msg=str(e))
        except ValueError as e:
            logger.error(f"JSON decode error: {e}")
            return NetworkResponse(status=-1, data={}, err_msg="Invalid JSON response.")

    def login(self) -> NetworkResponse:
        """
        Perform login to the server.

        Returns:
            NetworkResponse: The response wrapped in NetworkResponse.
        """
        if not self.ready:
            err_msg = "Network not ready."
            logger.error(err_msg)
            return NetworkResponse(status=-1, err_msg=err_msg, data={})

        url = (
            f"index.php?page=gamelogin&ver=0.1&tz=7&device_id=51dd0b0337f00c2e03c5bb110a56f818"
            f"&device_name=OPPO&username={self.username}&password={md5(self.password)}"
        )
        logger.info("Attempting to login...")
        result = self._post(url, {1: 1})

        if result.status == 0:
            loginResult = result.data
            self.ppy_id = loginResult.get('ppy_id')
            self.ssid = loginResult.get('ssid')
            logger.info("Login successful.")
            return NetworkResponse(status=0, data=loginResult)
        else:
            logger.error(f"Login failed: {result.err_msg}")
            return NetworkResponse(status=-1, err_msg=result.err_msg, data={})

    def getSession(self) -> dict:
        """
        Retrieve the current session information.

        Returns:
            dict: Dictionary containing session ID and ppy_id.
        """
        return {"sess_id": self.ssid, "ppy_id": self.ppy_id}

    def changePlanet(self, planetId: int) -> NetworkResponse:
        """
        Change the active planet.

        Args:
            planetId (int): The ID of the planet to switch to.

        Returns:
            NetworkResponse: The response wrapped in NetworkResponse.
        """
        url = 'game.php?page=buildings&mode='
        args = {"cp": planetId}
        logger.info(f"Changing planet to ID: {planetId}")

        result = self._post(url, args)
        if result.status == 0:
            data = result.data.get('data')
            if data:
                logger.info("Planet changed successfully.")
                return NetworkResponse(status=0, data=data)
        logger.error("Failed to change planet. Retrying after 15 seconds.")
        time.sleep(15)
        return self.changePlanet(planetId)

    def __del__(self):
        self.session.close()
        if self.proxy:
            # Delete proxy
            try:
                response = self.session.get(f"{PROXY_BASE_URL}/delete/?proxy={self.proxy}", auth=(PROXY_AUTH_USER, PROXY_AUTH_PASS), timeout=5)
                response.raise_for_status()
                logger.info("Proxy deleted successfully.")
            except requests.exceptions.RequestException as e:
                logger.error(f"Failed to delete proxy: {e}")


if __name__ == "__main__":
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s [%(levelname)s] %(name)s: %(message)s',
        handlers=[
            logging.StreamHandler(sys.stdout),
            logging.FileHandler("network.log")
        ]
    )
    user = Account("saltfish", "A123456", "", "g26")
    network = Network(user)
    print(network.login())
    print(network.ppy_id, network.ssid)
