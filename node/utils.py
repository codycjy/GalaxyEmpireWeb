import hashlib
import logging
import requests
import config

SALT = "b6bd8a93c54cc404c80d5a6833ba12eb"


def crypto(url, opt=""):
    """
    :param url: request url
    :param opt: request params
    :return: encrypted url
    """
    opt_w = opt + SALT
    data = opt + "&verifyKey=" + md5(url + opt_w)
    return data


def md5(parm):
    """
    :param parm: data to be encrypted
    :return: encrypted data
    """
    parm = str(parm)
    m = hashlib.md5()
    m.update(bytes(parm, "utf-8"))
    return m.hexdigest()


def updateServerUrl():
    serverListUrl = "http://192.81.130.154/gc2_sl_google.php"
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
    }
    response = requests.get(serverListUrl, timeout=5, headers=headers)
    if response.status_code == 200:
        data = response.json()
    else:
        logging.error("Failed to fetch server list")
        return

    def parse_name(name):
        # Check if name starts with 'g' followed by numbers and colon
        if name.startswith("g") and ":" in name:
            # Extract the g-number part before colon
            g_num = name.split(":")[0]
            return g_num
        else:
            # Take first 2 characters
            return str.lower(name[:2])

    url_map = {parse_name(item["name"]): item["url"] for item in data["lists"]}
    config.serverUrlList = url_map
    logging.info("Server list updated")
