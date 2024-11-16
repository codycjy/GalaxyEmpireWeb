import json

from model.user import Account
from model.fleet import Fleet
from model.target import Target
from model.task import Task
if  __name__ == "__main__":
    user = Account.from_dict({"server": "ze",
                            "username": "FZ920",
                            "password": "fan920",
                            "email": "email"
                            })
    fleet = Fleet.from_dict({'de': 500})
    target = Target.from_dict({'galaxy': 74, 'system': 24, 'planet': 8})
    task = Task.from_dict({'task_id': 1, 'task_type': 1, 'account': user, 'fleet': fleet, 'repeat': 1, 'target': target})

    print(task)
    with open("test.json", "r") as f:
        data = f.read()

    json_data = json.loads(data)
    task = Task.from_dict(json_data)
    print(task)
