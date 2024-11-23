from queue import Queue
from network import Network
from model.task import TaskType, TaskResult
from model.user import Account


def login_action(user: Account, result_queue: Queue):
    network = Network(user)
    results = network.login()
    succeed = results['status'] == 0
    result_queue.put(TaskResult(task_id=0, status=succeed, task_type=TaskType.LOGIN, back_ts=results['back_ts']))
