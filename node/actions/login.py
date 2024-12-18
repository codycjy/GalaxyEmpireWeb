from queue import Queue
from network import Network
from model.task import TaskType, TaskResult, TaskStatus, Task


def login_action(task: Task, result_queue: Queue):
    user = task.account
    uuid = task.uuid
    network = Network(user)
    results = network.login()
    succeed = TaskStatus.SUCCESS if results.status == 0 else TaskStatus.FAILED
    data = results.data
    result_queue.put(TaskResult(task_id=0, status=succeed, task_type=TaskType.LOGIN, uuid=uuid))
