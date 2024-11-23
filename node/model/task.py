from dataclasses_json import dataclass_json
from dataclasses import dataclass
from model.user import Account
from model.fleet import Fleet
from model.target import Target
from enum import Enum


class TaskType(Enum):
    ATTACK = 1
    EXPLORE = 4
    ESCAPE = "escape"
    LOGIN = 99


class MissionType(Enum):
    ATTACK = 1
    EXPLORE = 15
    ESCAPE = "escape"  # TODO: check this later


@dataclass_json
@dataclass
class Task:
    task_id: int
    task_type: TaskType
    account: Account
    fleet: Fleet
    repeat: int  # only works for explore and attack tasks
    target: Target


@dataclass_json
@dataclass
class TaskResult:
    task_id: int
    status: int
    task_type: TaskType
    back_ts: int


if __name__ == "__main__":
    pass
