from dataclasses import dataclass
from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class Account:
    username: str
    password: str
    email: str
    server: str


if __name__ == '__main__':
    user = Account(
        username='test',
        password='test',
        email='test@test.com',
        server='test-server'
    )
