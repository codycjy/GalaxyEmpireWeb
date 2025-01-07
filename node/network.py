import os
import sys
import time
import logging
import requests
from dataclasses import dataclass
from model.user import Account
from config import serverUrlList
from utils import crypto, md5
from proxy_pool.proxy_pool import ProxyPool

# Configure logging
logger = logging.getLogger(__name__)

headers = {
    "User-Agent": "android",
    "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
}


def addArgs(args):  # TODO: move me later
    if not args:
        return ""
    return "&" + "&".join(f"{key}={value}" for key, value in args.items())


@dataclass
class NetworkResponse:
    status: int
    data: dict
    err_msg: str = ""


class Network:
    def __init__(self, user: Account, proxy_pool: ProxyPool):
        self.server = user.server
        self.username = user.username
        self.password = user.password
        self.ppy_id = None
        self.ssid = None
        self.session = requests.Session()
        self.ready = True
        self.max_login_retries = 3  # Maximum retry attempts
        self.login_retry_count = 0  # Current retry count
        self.planet_id_table = {}
        self.proxy_pool = proxy_pool

        if os.getenv("PROXY", False):
            self.ready = False
            self.set_proxy()
        else:
            self.ready = True

    def set_proxy(self, max_retries: int = 3, initial_delay: float = 1):
        """
        Set up the proxy for the session with exponential backoff.

        Args:
            max_retries (int): Maximum number of retry attempts
            initial_delay (float): Initial delay in seconds before first retry
        """
        self.ready = False

        for attempt in range(max_retries + 1):
            proxy = self.proxy_pool.get_proxy()
            if proxy:
                try:
                    self.session.proxies = {
                        "http": proxy,
                    }
                    # Test the proxy with a simple request
                    self.session.get("http://example.com", timeout=5)
                    logger.info(f"Proxy set successfully to {proxy}")
                    self.ready = True
                    return
                except requests.exceptions.RequestException as e:
                    logger.warning(f"Proxy {proxy} failed: {e}")
                    self.proxy_pool.delete_proxy(proxy)

                    if attempt < max_retries:
                        delay = initial_delay * (2**attempt)  # Exponential backoff
                        logger.warning(f"Retrying proxy setup in {delay} seconds...")
                        time.sleep(delay)
                    continue

        logger.error("Failed to set up proxy after all attempts")

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
            req = self.session.post(
                full_url, headers=headers, data=crypto(full_url), timeout=5
            )
            req.raise_for_status()
            data = req.json()
            logger.debug(f"Response JSON: {data}")

            if data.get("status") != "error":
                self.login_retry_count = 0  # Reset retry count on successful response
                return NetworkResponse(status=0, data=data)

            if data.get("err_code") == 111:
                if self.login_retry_count >= self.max_login_retries:
                    logger.error("Max login retries exceeded")
                    return NetworkResponse(
                        status=-1, data={}, err_msg="Max login retries exceeded"
                    )

                logger.warning("Session expired. Attempting to relogin...")
                self.login_retry_count += 1
                login_response = self.login()
                if login_response.status == 0:
                    return self._post(url, args)
                return login_response

            return NetworkResponse(status=-1, data={}, err_msg=data.get("err_msg", ""))

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
            self.ppy_id = loginResult.get("ppy_id")
            self.ssid = loginResult.get("ssid")
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

    def change_planet(
        self, planetId: int = 0, max_retries: int = 3, initial_delay: float = 5
    ) -> NetworkResponse:
        """
        Change the active planet with exponential backoff retry mechanism.

        Args:
            planetId (int): The ID of the planet to switch to.
            max_retries (int): Maximum number of retry attempts.
            initial_delay (float): Initial delay in seconds before first retry.

        Returns:
            NetworkResponse: The response wrapped in NetworkResponse.
        """
        url = "game.php?page=buildings"
        args = {}
        logging.info("Updating planet ID table...")
        if planetId:
            logger.info("Changing planet to ID: %s", planetId)
            args["cp"] = planetId

        for attempt in range(max_retries + 1):
            logger.info(
                f"Changing planet to ID: {planetId} (Attempt {attempt + 1}/{max_retries + 1})"
            )

            result = self._post(url, args)
            if result.status == 0:
                data = result.data.get("result")
                if data:
                    logger.info("Planet changed successfully.")
                    self.update_planet_id_table(data)
                    return NetworkResponse(status=0, data=data)
            logger.error(f"Failed to change planet: {result.err_msg}")

            if attempt < max_retries:
                delay = initial_delay * (2**attempt)  # Exponential backoff
                logger.warning(
                    f"Failed to change planet. Retrying in {delay} seconds..."
                )
                time.sleep(delay)

        logger.error(f"Failed to change planet after {max_retries + 1} attempts.")
        return NetworkResponse(
            status=-1,
            data={},
            err_msg=f"Failed to change planet after {max_retries + 1} attempts",
        )

    def update_planet_id_table(self, full_data: dict):
        logger.debug("Updating planet ID table...")

        planets_data = (
            full_data.get("buildInfo", {}).get("result", {}).get("Planets", {})
        )
        planet_id_table = {}
        for planet_id, planet_data in planets_data.items():
            position = ":".join(
                [
                    str(planet_data["galaxy"]),
                    str(planet_data["system"]),
                    str(planet_data["planet"]),
                    str(int(int(planet_data["planet_type"]) == 3)),
                ]
            )
            planet_id_table[position] = planet_id
            planet_id_table[planet_id] = position
        self.planet_id_table = planet_id_table


if __name__ == "__main__":
    logging.basicConfig(
        level=logging.DEBUG,
        format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
        handlers=[
            logging.StreamHandler(sys.stdout),
            logging.FileHandler("network.log"),
        ],
    )
