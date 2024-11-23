import hashlib

SALT = "b6bd8a93c54cc404c80d5a6833ba12eb"


def crypto(url, opt=''):
    opt_w = opt + SALT
    data = opt + "&verifyKey=" + md5(url + opt_w)
    return data


def md5(parm):
    parm = str(parm)
    m = hashlib.md5()
    m.update(bytes(parm, 'utf-8'))
    return m.hexdigest()
