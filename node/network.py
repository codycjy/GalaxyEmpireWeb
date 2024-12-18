from dataclasses import dataclass
import requests
from utils import crypto, md5
from config import serverUrlList
import time
import sys
from model.user import Account

headers = {
    'User-Agent': 'android',
    'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'
}


def addArgs(args):  # TODO: move me later
    ans = ""
    for i in args.items():
        ans += '&' + str(i[0]) + '=' + str(i[1])
    return ans


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
        self.loggingPrefix = ''

    def _post(self, url: str, args={}) -> dict:
        if args is None:
            args = {}
        extra_args = {}
        if "login" not in url:
            extra_args = self.getSession()
        try:
            args.update(extra_args)
            full_url = serverUrlList[self.server] + url + addArgs(args)
        except KeyError as e:
            print("server wrong " + str(e))
            sys.exit(0)
        print(full_url)
        try:
            req = requests.post(full_url, headers=headers, data=crypto(full_url), timeout=5)
            data = req.json()
            if data['status'] != 'error':
                return {'status': 0, 'data': data}
            else:
                if data['err_code'] == 111:
                    print("session expired, relogin")
                    self.login()
                    return self._post(url, args)
                try:
                    print(data['err_msg'])
                    return {'status': -1, 'err_msg': data['err_msg'], 'err_code': data['err_code']}
                except KeyError:
                    print(data)
                    return {'status': -1}
        except Exception as e:
            print("JSONDecodeError " + str(e))
            return {'status': -1}

    def login(self):
        url = f'index.php?page=gamelogin&ver=0.1&tz=7&device_id=51dd0b0337f00c2e03c5bb110a56f818&device_name=OPPO&username={self.username}&password={md5(self.password)}'
        result = self._post(url, {1: 1})
        if result['status'] == 0:
            loginResult = result['data']
            self.ppy_id = loginResult['ppy_id']
            self.ssid = loginResult['ssid']
            print("Login Success")
            return NetworkResponse(status=0, data=loginResult)
        else:
            print("login failed")
            loginResult = result.get('data')
            return NetworkResponse(status=-1, err_msg=result.get('err_msg', "Login failed"), data={})

    def getSession(self):
        return {"sess_id": self.ssid, "ppy_id": self.ppy_id}

    def changePlanet(self, planetId):
        url = f'game.php?page=buildings&mode='
        args = {"cp": planetId}
        print('changePlanet: ' + str(planetId))
        result = self._post(url, args)
        if result['status'] == 0:
            data = result.get('data')
            if data:
                return data
        print("changePlanet failed, sleep 15s and retry")
        time.sleep(15)
        return self.changePlanet(planetId)


if __name__ == '__main__':
    user = Account(server='server', username='username', password='password', email='email')
    network = Network(user)
    network.login()
